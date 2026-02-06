package fs

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// ScanProgress reports the scanner's current state.
type ScanProgress struct {
	DirsScanned int64
	FilesFound  int64
	BytesTotal  int64
	Done        bool
}

// ScanResult is the final output of a scan.
type ScanResult struct {
	Tree  *Tree
	Error error
}

// ScannerOptions configures the scanner.
type ScannerOptions struct {
	WorkerCount    int      // number of concurrent directory readers (default: NumCPU * 2)
	MaxDepth       int      // maximum recursion depth (0 = unlimited)
	IgnorePatterns []string // glob patterns to skip
	ShowHidden     bool     // if false, skip dotfiles/dotdirs (default: false)
}

// Scanner performs concurrent filesystem scanning.
type Scanner struct {
	workerCount    int
	maxDepth       int
	ignorePatterns []string
	showHidden     bool

	// Atomic counters for progress
	dirsScanned atomic.Int64
	filesFound  atomic.Int64
	bytesTotal  atomic.Int64
}

// NewScanner creates a configured scanner.
func NewScanner(opts ScannerOptions) *Scanner {
	workers := opts.WorkerCount
	if workers <= 0 {
		workers = runtime.NumCPU() * 2
	}

	patterns := opts.IgnorePatterns
	if len(patterns) == 0 {
		patterns = defaultIgnorePatterns()
	}

	return &Scanner{
		workerCount:    workers,
		maxDepth:       opts.MaxDepth,
		ignorePatterns: patterns,
		showHidden:     opts.ShowHidden,
	}
}

// defaultIgnorePatterns returns patterns that are skipped by default.
func defaultIgnorePatterns() []string {
	return []string{
		".git", ".hg", ".svn",
		"node_modules",
		".DS_Store", "Thumbs.db",
		"$RECYCLE.BIN", "System Volume Information",
		"proc", "sys", "dev",
		".Trash", ".Spotlight-V100", ".fseventsd",
		".DocumentRevisions-V100", ".TemporaryItems",
	}
}

// shouldIgnore checks if a name matches any ignore pattern.
func (s *Scanner) shouldIgnore(name string) bool {
	for _, pattern := range s.ignorePatterns {
		if strings.EqualFold(name, pattern) {
			return true
		}
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

// Progress returns the current scan progress (safe for concurrent reads).
func (s *Scanner) Progress() ScanProgress {
	return ScanProgress{
		DirsScanned: s.dirsScanned.Load(),
		FilesFound:  s.filesFound.Load(),
		BytesTotal:  s.bytesTotal.Load(),
	}
}

// Scan starts a background scan rooted at the given path.
// Returns immediately. Results arrive on the returned channel.
func (s *Scanner) Scan(ctx context.Context, root string) <-chan ScanResult {
	resultCh := make(chan ScanResult, 1)

	go func() {
		defer close(resultCh)

		// Reset counters
		s.dirsScanned.Store(0)
		s.filesFound.Store(0)
		s.bytesTotal.Store(0)

		tree, err := s.scanSync(ctx, root)
		resultCh <- ScanResult{Tree: tree, Error: err}
	}()

	return resultCh
}

// ScanSync performs a blocking scan (useful for tests).
func (s *Scanner) ScanSync(ctx context.Context, root string) (*Tree, error) {
	return s.scanSync(ctx, root)
}

func (s *Scanner) scanSync(ctx context.Context, root string) (*Tree, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &os.PathError{Op: "scan", Path: absRoot, Err: os.ErrInvalid}
	}

	rootEntry := &Entry{
		Name:    filepath.Base(absRoot),
		Path:    absRoot,
		Type:    TypeDir,
		ModTime: info.ModTime(),
		Depth:   0,
	}

	sem := make(chan struct{}, s.workerCount)
	var wg sync.WaitGroup

	wg.Add(1)
	s.walkDir(ctx, rootEntry, sem, &wg)
	wg.Wait()

	tree := buildTree(rootEntry)
	return tree, nil
}

// walkDir recursively scans a directory using bounded concurrency.
func (s *Scanner) walkDir(ctx context.Context, parent *Entry, sem chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	if ctx.Err() != nil {
		return
	}

	if s.maxDepth > 0 && parent.Depth >= s.maxDepth {
		return
	}

	// Acquire semaphore
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return
	}

	dirEntries, err := os.ReadDir(parent.Path)
	<-sem // Release semaphore

	if err != nil {
		parent.Error = err.Error()
		s.dirsScanned.Add(1)
		return
	}

	s.dirsScanned.Add(1)

	children := make([]*Entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		if ctx.Err() != nil {
			return
		}

		if s.shouldIgnore(de.Name()) {
			continue
		}

		// Skip hidden files/dirs (dotfiles) unless configured to show them
		if !s.showHidden && strings.HasPrefix(de.Name(), ".") {
			continue
		}

		child := &Entry{
			Name:  de.Name(),
			Path:  filepath.Join(parent.Path, de.Name()),
			Depth: parent.Depth + 1,
		}

		switch {
		case de.Type()&os.ModeSymlink != 0:
			child.Type = TypeSymlink
			// Try to get symlink target info for size
			if info, err := os.Stat(child.Path); err == nil {
				child.Size = info.Size()
				child.ModTime = info.ModTime()
			} else {
				// Broken symlink - use lstat info
				if linfo, lerr := os.Lstat(child.Path); lerr == nil {
					child.ModTime = linfo.ModTime()
				}
			}
			s.filesFound.Add(1)

		case de.IsDir():
			child.Type = TypeDir
			if info, err := de.Info(); err == nil {
				child.ModTime = info.ModTime()
			}
			wg.Add(1)
			go s.walkDir(ctx, child, sem, wg)

		case de.Type().IsRegular():
			child.Type = TypeFile
			if info, err := de.Info(); err == nil {
				child.Size = info.Size()
				child.ModTime = info.ModTime()
				s.bytesTotal.Add(child.Size)
			}
			s.filesFound.Add(1)

		default:
			child.Type = TypeOther
			if info, err := de.Info(); err == nil {
				child.ModTime = info.ModTime()
			}
			s.filesFound.Add(1)
		}

		children = append(children, child)
	}

	parent.Children = children
	parent.Loaded = true
}

// LoadDir synchronously scans a single directory's immediate children.
// Used for lazy/on-demand loading when the user expands a directory.
func (s *Scanner) LoadDir(entry *Entry) error {
	if entry.Type != TypeDir || entry.Loaded {
		return nil
	}

	dirEntries, err := os.ReadDir(entry.Path)
	if err != nil {
		entry.Error = err.Error()
		entry.Loaded = true
		return err
	}

	children := make([]*Entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		if s.shouldIgnore(de.Name()) {
			continue
		}
		if !s.showHidden && strings.HasPrefix(de.Name(), ".") {
			continue
		}

		child := &Entry{
			Name:  de.Name(),
			Path:  filepath.Join(entry.Path, de.Name()),
			Depth: entry.Depth + 1,
		}

		switch {
		case de.Type()&os.ModeSymlink != 0:
			child.Type = TypeSymlink
			if info, err := os.Stat(child.Path); err == nil {
				child.Size = info.Size()
				child.ModTime = info.ModTime()
			}
		case de.IsDir():
			child.Type = TypeDir
			if info, err := de.Info(); err == nil {
				child.ModTime = info.ModTime()
			}
		case de.Type().IsRegular():
			child.Type = TypeFile
			if info, err := de.Info(); err == nil {
				child.Size = info.Size()
				child.ModTime = info.ModTime()
			}
		default:
			child.Type = TypeOther
		}

		children = append(children, child)
	}

	sort.Slice(children, func(i, j int) bool {
		return children[i].Size > children[j].Size
	})

	entry.Children = children
	entry.Loaded = true

	var total int64
	for _, c := range children {
		total += c.Size
	}
	entry.Size = total
	return nil
}
