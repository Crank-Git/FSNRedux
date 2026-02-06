package app

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
	"github.com/Crank-Git/FSNRedux/internal/fs"
	"github.com/Crank-Git/FSNRedux/internal/input"
	"github.com/Crank-Git/FSNRedux/internal/layout"
	"github.com/Crank-Git/FSNRedux/internal/renderer"
	"github.com/Crank-Git/FSNRedux/internal/scene"
	"github.com/Crank-Git/FSNRedux/internal/ui"
)

// Config holds application configuration from CLI flags.
type Config struct {
	RootPath   string
	Width      int
	Height     int
	MaxDepth   int
	Theme      string
	ShowHidden bool
}

// App is the main application that wires all subsystems together.
type App struct {
	config Config

	// Subsystems
	scanner    *fs.Scanner
	renderer   *renderer.Renderer
	inputState *input.InputState

	// State
	tree          *fs.Tree
	graph         *scene.Graph
	treeViewState *ui.TreeViewState
	scanning      bool
	scanResult    <-chan fs.ScanResult
	selectedPath  string
	expandedPaths map[string]bool // tracks which dirs are expanded in 3D view

	// Input bar (path entry / search)
	inputBar      ui.InputBar
	searchResults []string // paths matching current search
	searchIndex   int      // current search result index

	// Inspect panel
	inspectOpen bool
	inspectInfo *fs.InspectInfo

	// Settings menu
	settings *ui.SettingsState

	// File preview
	preview ui.PreviewState
}

// New creates the application with the given config.
func New(cfg Config) *App {
	return &App{
		config:        cfg,
		scanner:       fs.NewScanner(fs.ScannerOptions{MaxDepth: 1, ShowHidden: cfg.ShowHidden}),
		renderer:      renderer.New(),
		inputState:    input.NewInputState(),
		expandedPaths: make(map[string]bool),
		settings:      ui.NewSettingsState(cfg.ShowHidden, cfg.Theme, cfg.MaxDepth, true),
	}
}

// Run is the main entry point - initializes window and runs the main loop.
func (a *App) Run() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(a.config.Width), int32(a.config.Height),
		fmt.Sprintf("FSNRedux - %s", a.config.RootPath))
	defer rl.CloseWindow()
	color.InitTheme(a.config.Theme)
	ui.LoadFont()
	defer ui.UnloadFont()
	rl.SetTargetFPS(60)
	rl.SetExitKey(0) // Disable Escape-to-quit so Escape works for in-app actions

	// Start initial scan
	a.startScan()

	for !rl.WindowShouldClose() {
		a.update()
		a.draw()
	}
}

// startScan kicks off an async filesystem scan.
func (a *App) startScan() {
	a.scanning = true
	a.tree = nil
	a.graph = nil
	a.scanResult = a.scanner.Scan(context.Background(), a.config.RootPath)
}

// update handles input and checks for scan completion.
func (a *App) update() {
	// Check if scan completed
	if a.scanning && a.scanResult != nil {
		select {
		case result, ok := <-a.scanResult:
			if ok {
				a.scanning = false
				if result.Error == nil && result.Tree != nil {
					a.tree = result.Tree
					a.treeViewState = ui.NewTreeViewState(a.tree.Root.Path)
					a.expandedPaths[a.tree.Root.Path] = true
					a.rebuildLayout(true)
				}
			}
		default:
			// Still scanning
		}
	}

	// Sync text input state to disable camera/shortcut keys
	sidebarSearchActive := a.treeViewState != nil && a.treeViewState.SearchActive
	textActive := a.inputBar.Active || sidebarSearchActive
	modalOpen := a.inspectOpen || a.settings.Open || a.preview.Open
	a.inputState.TextInputActive = textActive || modalOpen
	a.inputState.Camera.KeyboardEnabled = !textActive && !modalOpen

	// Check sidebar search submit
	if a.treeViewState != nil && a.treeViewState.SearchSubmit != "" {
		a.searchFor(a.treeViewState.SearchSubmit)
		a.treeViewState.SearchSubmit = ""
	}

	// Handle input bar
	if a.inputBar.Active {
		if a.inputBar.Update() {
			a.handleInputBarSubmit()
		}
		return // input bar consumes all keyboard input
	}

	// Handle inspect panel (consumes input when open)
	if a.inspectOpen {
		if rl.IsKeyPressed(rl.KeySpace) || rl.IsKeyPressed(rl.KeyEscape) {
			a.inspectOpen = false
			a.inspectInfo = nil
		}
		return
	}

	// Handle preview panel (consumes input when open)
	if a.preview.Open {
		if a.preview.Update() {
			a.preview.Close()
		}
		// O key opens file with default app even from preview
		if rl.IsKeyPressed(rl.KeyO) {
			a.openWithDefault(a.preview.FilePath)
		}
		return
	}

	// Handle settings menu (consumes input when open)
	if a.settings.Open {
		if rl.IsKeyPressed(rl.KeyComma) || rl.IsKeyPressed(rl.KeyEscape) {
			a.settings.Open = false
		}
		return
	}

	// Process 3D input
	if a.graph != nil {
		clickedPath := a.inputState.Update(a.graph, ui.SidebarWidth)
		if clickedPath != "" {
			a.handleClickedPath(clickedPath)
		}

		// Path bar (Ctrl+L)
		if a.inputState.PathBarRequested {
			initial := a.config.RootPath
			if a.selectedPath != "" {
				initial = a.selectedPath
			}
			a.inputBar.Open(ui.InputBarPath, initial)
			return
		}

		// Search (F key -> sidebar search)
		if a.inputState.SearchRequested {
			if a.treeViewState != nil {
				a.treeViewState.SearchActive = true
				a.treeViewState.SearchText = ""
				a.treeViewState.SearchCursor = 0
			}
			return
		}

		// Enter = expand selected directory
		if a.inputState.ExpandRequested {
			if sel := a.inputState.Picker.SelectedNode; sel != nil && sel.Entry != nil && sel.Entry.IsDir() {
				if !a.expandedPaths[sel.Entry.Path] {
					a.expandDir(sel.Entry.Path, sel)
				}
			}
		}

		// Escape = collapse selected dir / go to parent
		if a.inputState.BackRequested {
			// First clear search results if active
			if len(a.searchResults) > 0 {
				a.searchResults = nil
				a.searchIndex = 0
			} else if sel := a.inputState.Picker.SelectedNode; sel != nil {
				if sel.Entry != nil && sel.Entry.IsDir() && a.expandedPaths[sel.Entry.Path] {
					// Collapse current dir
					delete(a.expandedPaths, sel.Entry.Path)
					if a.treeViewState != nil {
						delete(a.treeViewState.ExpandedDirs, sel.Entry.Path)
					}
					a.selectedPath = sel.Entry.Path
					a.rebuildLayout(false)
				} else if sel.Parent != nil {
					// Go to parent
					a.inputState.Picker.SelectedNode = sel.Parent
					a.selectedPath = ""
					if sel.Parent.Entry != nil {
						a.selectedPath = sel.Parent.Entry.Path
					}
					a.inputState.FocusOnNode(sel.Parent)
				}
			}
		}

		// Home = focus on root
		if a.inputState.HomeRequested && a.graph.Root != nil {
			a.inputState.Picker.SelectedNode = a.graph.Root
			if a.graph.Root.Entry != nil {
				a.selectedPath = a.graph.Root.Entry.Path
			}
			a.inputState.FocusOnNode(a.graph.Root)
		}

		// B = birdseye view
		if a.inputState.BirdseyeRequested {
			a.birdseyeView()
		}

		// Tab / Shift+Tab = cycle through visible nodes
		if a.inputState.NextNodeRequested {
			a.selectNextVisible(1)
		}
		if a.inputState.PrevNodeRequested {
			a.selectNextVisible(-1)
		}

		// Space = inspect/preview selected node
		if a.inputState.InspectRequested {
			if sel := a.inputState.Picker.SelectedNode; sel != nil && sel.Entry != nil {
				if sel.Entry.IsDir() {
					// Directories get the inspect panel
					info := sel.Entry.Inspect()
					a.inspectInfo = &info
					a.inspectOpen = true
				} else {
					// Files get the preview panel
					a.preview.OpenPreview(sel.Entry.Path)
				}
			}
		}

		// O = open selected file with default application
		if a.inputState.OpenFileRequested {
			if sel := a.inputState.Picker.SelectedNode; sel != nil && sel.Entry != nil {
				a.openWithDefault(sel.Entry.Path)
			}
		}

		// Comma = open settings
		if a.inputState.SettingsRequested {
			a.settings.Open = true
		}

		// Search result navigation: N=next, P=prev
		if len(a.searchResults) > 0 && !a.inputState.TextInputActive {
			if rl.IsKeyPressed(rl.KeyN) {
				a.navigateToSearchResult((a.searchIndex + 1) % len(a.searchResults))
			}
			if rl.IsKeyPressed(rl.KeyP) {
				idx := a.searchIndex - 1
				if idx < 0 {
					idx = len(a.searchResults) - 1
				}
				a.navigateToSearchResult(idx)
			}
		}
	}
}

// selectNextVisible cycles selection through visible nodes.
func (a *App) selectNextVisible(direction int) {
	if a.graph == nil {
		return
	}

	// Build flat list of visible nodes
	var visible []*scene.SceneNode
	a.graph.Traverse(func(node *scene.SceneNode) bool {
		visible = append(visible, node)
		return true
	})
	if len(visible) == 0 {
		return
	}

	// Find current index
	current := -1
	for i, n := range visible {
		if n == a.inputState.Picker.SelectedNode {
			current = i
			break
		}
	}

	// Move
	next := current + direction
	if next < 0 {
		next = len(visible) - 1
	} else if next >= len(visible) {
		next = 0
	}

	node := visible[next]
	a.inputState.Picker.SelectedNode = node
	if node.Entry != nil {
		a.selectedPath = node.Entry.Path
		if a.treeViewState != nil {
			a.treeViewState.SelectedPath = node.Entry.Path
		}
	}
	a.inputState.FocusOnNode(node)
}

// handleClickedPath processes a double-clicked path (expand/collapse dirs).
func (a *App) handleClickedPath(clickedPath string) {
	a.selectedPath = clickedPath
	if a.treeViewState != nil {
		a.treeViewState.SelectedPath = clickedPath
	}
	// Expand/collapse directories on double-click
	if node := a.graph.FindByPath(clickedPath); node != nil && node.Entry != nil && node.Entry.IsDir() {
		if a.expandedPaths[clickedPath] {
			// Collapse
			delete(a.expandedPaths, clickedPath)
			if a.treeViewState != nil {
				delete(a.treeViewState.ExpandedDirs, clickedPath)
			}
			a.rebuildLayout(false)
		} else {
			// Expand
			a.expandDir(clickedPath, node)
		}
	}
}

// expandDir expands a directory node, loading children if needed.
func (a *App) expandDir(path string, node *scene.SceneNode) {
	a.expandedPaths[path] = true
	if a.treeViewState != nil {
		a.treeViewState.ExpandedDirs[path] = true
	}
	if !node.Entry.Loaded {
		a.scanner.LoadDir(node.Entry)
	}
	a.selectedPath = path
	a.rebuildLayout(false)
	if newNode := a.graph.FindByPath(path); newNode != nil {
		a.inputState.FocusOnNode(newNode)
	}
}

// handleInputBarSubmit processes the input bar when the user presses Enter.
func (a *App) handleInputBarSubmit() {
	text := strings.TrimSpace(a.inputBar.Text)
	mode := a.inputBar.Mode
	a.inputBar.Close()

	if text == "" {
		return
	}

	switch mode {
	case ui.InputBarPath:
		a.navigateToPath(text)
	case ui.InputBarSearch:
		a.searchFor(text)
	}
}

// navigateToPath changes the root to a new filesystem path.
func (a *App) navigateToPath(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		return
	}

	// Check if path is within current tree - just navigate to it
	if a.graph != nil {
		if node := a.graph.FindByPath(absPath); node != nil {
			a.selectedPath = absPath
			a.inputState.Picker.SelectedNode = node
			a.inputState.FocusOnNode(node)
			// Expand parent chain
			a.expandParentChain(absPath)
			return
		}
	}

	// New root - restart scan
	a.config.RootPath = absPath
	a.expandedPaths = map[string]bool{absPath: true}
	a.selectedPath = ""
	rl.SetWindowTitle(fmt.Sprintf("FSNRedux - %s", absPath))
	a.startScan()
}

// expandParentChain ensures all ancestors of the given path are expanded.
func (a *App) expandParentChain(path string) {
	for path != a.config.RootPath && path != "/" && path != "." {
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		if node := a.graph.FindByPath(parent); node != nil && node.Entry != nil && node.Entry.IsDir() {
			if !a.expandedPaths[parent] {
				a.expandDir(parent, node)
			}
		}
		path = parent
	}
}

// searchFor finds entries matching the query and navigates to the first result.
func (a *App) searchFor(query string) {
	if a.tree == nil || a.tree.Root == nil {
		return
	}

	// Search through loaded entries
	a.searchResults = nil
	a.searchIndex = 0
	q := strings.ToLower(query)
	a.searchEntries(a.tree.Root, q)

	// Sort results by path for consistent ordering
	sort.Strings(a.searchResults)

	if len(a.searchResults) > 0 {
		a.navigateToSearchResult(0)
	}
}

// searchEntries recursively searches loaded entries for name matches.
func (a *App) searchEntries(entry *fs.Entry, query string) {
	if strings.Contains(strings.ToLower(entry.Name), query) {
		a.searchResults = append(a.searchResults, entry.Path)
	}
	if entry.Loaded {
		for _, child := range entry.Children {
			a.searchEntries(child, query)
		}
	}
}

// navigateToSearchResult navigates to the n-th search result.
func (a *App) navigateToSearchResult(index int) {
	if index < 0 || index >= len(a.searchResults) {
		return
	}
	a.searchIndex = index
	path := a.searchResults[index]

	// Expand parent chain to make the result visible
	a.expandParentChain(path)

	// After expanding parents, rebuild may have happened - find the node
	if node := a.graph.FindByPath(path); node != nil {
		a.selectedPath = path
		a.inputState.Picker.SelectedNode = node
		a.inputState.FocusOnNode(node)
		if a.treeViewState != nil {
			a.treeViewState.SelectedPath = path
		}
	}
}

// rebuildLayout recomputes the layout and scene graph from the current tree.
// autoFrame controls whether the camera is repositioned to show everything.
func (a *App) rebuildLayout(autoFrame bool) {
	if a.tree == nil {
		return
	}
	opts := layout.DefaultOptions(layout.ModeTreeV)
	opts.ExpandedPaths = a.expandedPaths
	layoutRoot := layout.Compute(a.tree, opts)
	a.graph = scene.NewGraph(layoutRoot, a.expandedPaths)

	// Restore selection pointer after rebuild
	if a.selectedPath != "" {
		a.inputState.Picker.SelectedNode = a.graph.FindByPath(a.selectedPath)
	}
	a.inputState.Picker.HoveredNode = nil

	if autoFrame {
		a.frameCamera()
	}
}

// frameCamera positions the camera to see the entire scene.
func (a *App) frameCamera() {
	if a.graph == nil || a.graph.Root == nil {
		return
	}
	minBounds := rl.NewVector3(float32(1e30), float32(1e30), float32(1e30))
	maxBounds := rl.NewVector3(float32(-1e30), float32(-1e30), float32(-1e30))

	a.graph.Traverse(func(node *scene.SceneNode) bool {
		if node.Bounds.Min.X < minBounds.X {
			minBounds.X = node.Bounds.Min.X
		}
		if node.Bounds.Min.Y < minBounds.Y {
			minBounds.Y = node.Bounds.Min.Y
		}
		if node.Bounds.Min.Z < minBounds.Z {
			minBounds.Z = node.Bounds.Min.Z
		}
		if node.Bounds.Max.X > maxBounds.X {
			maxBounds.X = node.Bounds.Max.X
		}
		if node.Bounds.Max.Y > maxBounds.Y {
			maxBounds.Y = node.Bounds.Max.Y
		}
		if node.Bounds.Max.Z > maxBounds.Z {
			maxBounds.Z = node.Bounds.Max.Z
		}
		return true
	})

	a.inputState.Camera.FrameScene(minBounds, maxBounds)
}

// draw renders one frame.
func (a *App) draw() {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	rl.BeginDrawing()
	rl.ClearBackground(color.Background)

	// 3D viewport
	rl.BeginMode3D(a.inputState.Camera.Camera)
	renderer.DrawGround()
	if a.graph != nil {
		a.renderer.DrawScene(a.graph, a.inputState.Picker.SelectedNode, a.inputState.Picker.HoveredNode)
	}
	rl.EndMode3D()

	// 3D labels + file icons projected to 2D (drawn after EndMode3D so they're always facing camera)
	// Uses shared placement tracker to prevent overlapping text/icons
	if a.graph != nil {
		var placed []screenRect
		placed = a.drawSceneLabels(placed)
		a.drawFileIcons(placed)
	}

	// Floating tooltip for hovered 3D node
	if a.inputState.Picker.HoveredNode != nil && a.inputState.Picker.HoveredNode.Entry != nil {
		hNode := a.inputState.Picker.HoveredNode
		screenPos := rl.GetWorldToScreen(rl.NewVector3(
			hNode.Position.X, hNode.Position.Y+hNode.Size.Y/2, hNode.Position.Z,
		), a.inputState.Camera.Camera)
		ui.DrawSelectedTooltip(hNode.Entry, screenPos.X, screenPos.Y)
	}

	// 2D UI overlay
	// Breadcrumb
	selectedEntry := a.getSelectedEntry()
	breadcrumbPath := a.config.RootPath
	if selectedEntry != nil {
		breadcrumbPath = selectedEntry.Path
	}
	clickedBreadcrumb := ui.DrawBreadcrumb(breadcrumbPath, a.config.RootPath, screenW)
	if clickedBreadcrumb != "" {
		a.inputState.FocusOnPath(a.graph, clickedBreadcrumb)
	}

	// Sidebar
	if a.tree != nil && a.treeViewState != nil {
		sidebarClicked := ui.DrawSidebar(a.tree, a.treeViewState, screenH)
		if sidebarClicked != "" {
			a.selectedPath = sidebarClicked
			a.inputState.FocusOnPath(a.graph, sidebarClicked)
		}
	}

	// Info panel
	ui.DrawInfoPanel(selectedEntry, screenH)

	// Input bar overlay
	a.inputBar.Draw(screenW)

	// Search results indicator
	if len(a.searchResults) > 0 {
		searchText := fmt.Sprintf("Search: %d/%d matches (N=next, P=prev, Esc=clear)",
			a.searchIndex+1, len(a.searchResults))
		stw := ui.MeasureTextUI(searchText, ui.SmallFontSize)
		sx := screenW - stw - 12
		sy := ui.BreadcrumbHeight + 30
		rl.DrawRectangle(sx-4, sy-1, stw+8, 15, rl.NewColor(0, 0, 0, 180))
		ui.DrawTextUI(searchText, sx, sy, ui.SmallFontSize, color.Active.LinkAccent)
	}

	// Inspect panel overlay
	if a.inspectOpen && a.inspectInfo != nil {
		ui.DrawInspectPanel(a.inspectInfo, screenW, screenH)
	}

	// Preview panel overlay
	if a.preview.Open {
		ui.DrawPreviewPanel(&a.preview, screenW, screenH)
	}

	// Settings panel overlay
	if a.settings.Open {
		action := ui.DrawSettingsPanel(a.settings, screenW, screenH)
		a.applySettingsAction(action)
	}

	// Scanning overlay
	if a.scanning {
		progress := a.scanner.Progress()
		ui.DrawScanProgress(progress.DirsScanned, progress.FilesFound,
			progress.BytesTotal, screenW, screenH)
	}

	// Help text (keep settings and H key toggle in sync)
	a.settings.ShowLegend = a.inputState.ShowHelp
	if a.inputState.ShowHelp {
		ui.DrawHelpText(screenW, screenH)
	}

	rl.EndDrawing()
}

// screenRect tracks a placed 2D element to prevent overlaps.
type screenRect struct {
	x, y, w, h int32
}

// rectsOverlap returns true if two rectangles overlap (with padding).
func rectsOverlap(a, b screenRect) bool {
	pad := int32(2)
	return a.x-pad < b.x+b.w+pad && a.x+a.w+pad > b.x-pad &&
		a.y-pad < b.y+b.h+pad && a.y+a.h+pad > b.y-pad
}

// anyOverlap returns true if r overlaps any rect in the list.
func anyOverlap(r screenRect, placed []screenRect) bool {
	for _, p := range placed {
		if rectsOverlap(r, p) {
			return true
		}
	}
	return false
}

// drawSceneLabels renders nearby directory names as 2D text projected from 3D positions.
// Returns updated placement list for downstream consumers.
func (a *App) drawSceneLabels(placed []screenRect) []screenRect {
	cam := a.inputState.Camera.Camera
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	labelsDrawn := 0
	maxLabels := 40

	a.graph.Traverse(func(node *scene.SceneNode) bool {
		if labelsDrawn >= maxLabels {
			return false
		}
		if node.Entry == nil || !node.Entry.IsDir() {
			return true
		}

		// Distance check first (cheap)
		dx := cam.Position.X - node.Position.X
		dy := cam.Position.Y - node.Position.Y
		dz := cam.Position.Z - node.Position.Z
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if dist > 50 {
			return true
		}

		// Position label above the pedestal
		labelPos := rl.NewVector3(
			node.Position.X,
			node.Position.Y+node.Size.Y/2+0.15,
			node.Position.Z,
		)
		screenPos := rl.GetWorldToScreen(labelPos, cam)

		if screenPos.X < 0 || screenPos.X > sw || screenPos.Y < 0 || screenPos.Y > sh {
			return true
		}

		alpha := uint8(255)
		if dist > 25 {
			alpha = uint8(255.0 * (1.0 - (dist-25.0)/25.0))
		}

		name := node.Entry.Name
		if len(name) > 18 {
			name = name[:16] + ".."
		}

		fontSize := float32(12)
		textWidth := ui.MeasureTextUI(name, fontSize)
		x := int32(screenPos.X) - textWidth/2
		y := int32(screenPos.Y)

		// Check for overlap with already-placed elements
		rect := screenRect{x - 2, y - 1, textWidth + 4, 14}
		if anyOverlap(rect, placed) {
			return true // skip this label
		}

		rl.DrawRectangle(rect.x, rect.y, rect.w, rect.h, rl.NewColor(0, 0, 0, alpha/2))
		ui.DrawTextUI(name, x, y, fontSize, rl.NewColor(
			color.Active.TextPrimary.R,
			color.Active.TextPrimary.G,
			color.Active.TextPrimary.B,
			alpha,
		))
		placed = append(placed, rect)
		labelsDrawn++

		return true
	})

	return placed
}

// getSelectedEntry returns the fs.Entry for the currently selected node.
func (a *App) getSelectedEntry() *fs.Entry {
	if a.inputState.Picker.SelectedNode != nil {
		return a.inputState.Picker.SelectedNode.Entry
	}
	return nil
}

// applySettingsAction handles a setting change from the settings panel.
func (a *App) applySettingsAction(action ui.SettingsAction) {
	switch action {
	case ui.SettingsToggleHidden:
		a.config.ShowHidden = a.settings.ShowHidden
		a.scanner = fs.NewScanner(fs.ScannerOptions{MaxDepth: 1, ShowHidden: a.config.ShowHidden})
		a.expandedPaths = map[string]bool{a.config.RootPath: true}
		a.selectedPath = ""
		a.inputState.Picker.SelectedNode = nil
		a.startScan()

	case ui.SettingsCycleTheme:
		a.config.Theme = a.settings.Theme
		color.InitTheme(a.config.Theme)

	case ui.SettingsToggleLegend:
		a.inputState.ShowHelp = a.settings.ShowLegend

	case ui.SettingsDepthUp, ui.SettingsDepthDown:
		a.config.MaxDepth = a.settings.MaxDepth
		// Rebuild layout with new depth (no re-scan needed)
		a.rebuildLayout(false)
	}
}

// drawFileIcons renders simple unicolor 2D icons on top of file pedestals.
func (a *App) drawFileIcons(placed []screenRect) {
	cam := a.inputState.Camera.Camera
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	iconsDrawn := 0
	maxIcons := 80

	a.graph.Traverse(func(node *scene.SceneNode) bool {
		if iconsDrawn >= maxIcons {
			return false
		}
		if node.Entry == nil || node.Entry.IsDir() {
			return true
		}

		// Distance check
		dx := cam.Position.X - node.Position.X
		dy := cam.Position.Y - node.Position.Y
		dz := cam.Position.Z - node.Position.Z
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if dist > 30 {
			return true
		}

		// Project top-center of pedestal to screen
		labelPos := rl.NewVector3(
			node.Position.X,
			node.Position.Y+node.Size.Y/2+0.02,
			node.Position.Z,
		)
		sp := rl.GetWorldToScreen(labelPos, cam)
		if sp.X < 0 || sp.X > sw || sp.Y < 0 || sp.Y > sh {
			return true
		}

		// Scale icon size by distance
		scale := 1.0 - (dist / 30.0)
		if scale < 0.3 {
			scale = 0.3
		}
		iconSize := int32(float32(10) * scale)
		if iconSize < 4 {
			return true
		}

		cx := int32(sp.X)
		cy := int32(sp.Y)

		// Check for overlap with placed labels/icons
		rect := screenRect{cx - iconSize, cy - iconSize, iconSize * 2, iconSize * 2}
		if anyOverlap(rect, placed) {
			return true
		}

		alpha := uint8(255)
		if dist > 15 {
			alpha = uint8(255.0 * (1.0 - (dist-15.0)/15.0))
		}

		icon, _ := ui.FileTypeIcon(node.Entry.Name, false)
		iconColor := ui.FileTypeIconColor(icon)
		iconColor.A = alpha

		drawSimpleIcon(icon, cx, cy, iconSize, iconColor)
		placed = append(placed, rect)
		iconsDrawn++
		return true
	})
}

// drawSimpleIcon draws a small unicolor geometric shape representing a file type.
func drawSimpleIcon(icon string, cx, cy, size int32, clr rl.Color) {
	s := size
	switch icon {
	case "Go", "Py", "JS", "TS", "TSX", "JSX", "Rs", "C", "C++", "Jv", "Rb",
		"Sw", "Kt", "Lua", "C#", "PHP", "Zig", "Drt", "Ex", "Hs", "ML", "R", "OC",
		"H", "H++", "Scl", "Exs", "Erl":
		// Code: angle brackets < >
		rl.DrawLine(cx-s, cy, cx-s/2, cy-s/2, clr)
		rl.DrawLine(cx-s, cy, cx-s/2, cy+s/2, clr)
		rl.DrawLine(cx+s, cy, cx+s/2, cy-s/2, clr)
		rl.DrawLine(cx+s, cy, cx+s/2, cy+s/2, clr)

	case "PNG", "JPG", "GIF", "BMP", "SVG", "WBP", "ICO", "TIF":
		// Image: small rectangle with triangle inside
		rl.DrawRectangleLines(cx-s, cy-s*3/4, s*2, s*3/2, clr)
		rl.DrawTriangle(
			rl.NewVector2(float32(cx-s/2), float32(cy+s/2)),
			rl.NewVector2(float32(cx), float32(cy-s/4)),
			rl.NewVector2(float32(cx+s/2), float32(cy+s/2)),
			clr,
		)

	case "MP3", "WAV", "FLC", "OGG", "AAC", "M4A":
		// Audio: note shape (circle + stem)
		rl.DrawCircle(cx-s/4, cy+s/4, float32(s)/3, clr)
		rl.DrawLine(cx-s/4+s/3, cy+s/4, cx-s/4+s/3, cy-s*3/4, clr)

	case "MP4", "MKV", "AVI", "MOV", "WBM", "WMV":
		// Video: play triangle
		rl.DrawTriangle(
			rl.NewVector2(float32(cx-s/2), float32(cy-s*3/4)),
			rl.NewVector2(float32(cx-s/2), float32(cy+s*3/4)),
			rl.NewVector2(float32(cx+s*3/4), float32(cy)),
			clr,
		)

	case "ZIP", "TAR", "GZ", "RAR", "7Z", "BZ2", "XZ", "ZST":
		// Archive: box with zipper line
		rl.DrawRectangleLines(cx-s, cy-s*3/4, s*2, s*3/2, clr)
		rl.DrawLine(cx, cy-s*3/4, cx, cy+s*3/4, clr)

	case "PDF", "DOC", "XLS", "PPT":
		// Document: page with folded corner
		rl.DrawRectangleLines(cx-s*3/4, cy-s, s*3/2, s*2, clr)
		rl.DrawLine(cx+s*3/4-s/2, cy-s, cx+s*3/4, cy-s+s/2, clr)

	case "MD", "TXT", "RST":
		// Text: horizontal lines
		rl.DrawLine(cx-s, cy-s/2, cx+s, cy-s/2, clr)
		rl.DrawLine(cx-s, cy, cx+s/2, cy, clr)
		rl.DrawLine(cx-s, cy+s/2, cx+s*3/4, cy+s/2, clr)

	case "Sh", "Bat", "PS":
		// Shell: prompt >_
		rl.DrawLine(cx-s, cy-s/3, cx, cy, clr)
		rl.DrawLine(cx-s, cy+s/3, cx, cy, clr)
		rl.DrawLine(cx, cy+s/2, cx+s, cy+s/2, clr)

	case "JSN", "YML", "TML", "XML", "INI", "CFG", "ENV":
		// Config: gear-like (small diamond)
		rl.DrawLine(cx, cy-s, cx+s, cy, clr)
		rl.DrawLine(cx+s, cy, cx, cy+s, clr)
		rl.DrawLine(cx, cy+s, cx-s, cy, clr)
		rl.DrawLine(cx-s, cy, cx, cy-s, clr)

	case "DB", "SQL":
		// Database: stacked ellipses (simplified as lines)
		rl.DrawEllipseLines(cx, cy-s/2, float32(s), float32(s)/3, clr)
		rl.DrawLine(cx-s, cy-s/2, cx-s, cy+s/2, clr)
		rl.DrawLine(cx+s, cy-s/2, cx+s, cy+s/2, clr)
		rl.DrawEllipseLines(cx, cy+s/2, float32(s), float32(s)/3, clr)

	default:
		// Generic file: simple rectangle
		rl.DrawRectangleLines(cx-s*3/4, cy-s, s*3/2, s*2, clr)
	}
}

// birdseyeView positions the camera overhead to show all expanded directories.
func (a *App) birdseyeView() {
	if a.graph == nil || a.graph.Root == nil {
		return
	}
	minBounds := rl.NewVector3(float32(1e30), float32(1e30), float32(1e30))
	maxBounds := rl.NewVector3(float32(-1e30), float32(-1e30), float32(-1e30))

	a.graph.Traverse(func(node *scene.SceneNode) bool {
		if node.Bounds.Min.X < minBounds.X {
			minBounds.X = node.Bounds.Min.X
		}
		if node.Bounds.Min.Z < minBounds.Z {
			minBounds.Z = node.Bounds.Min.Z
		}
		if node.Bounds.Max.X > maxBounds.X {
			maxBounds.X = node.Bounds.Max.X
		}
		if node.Bounds.Max.Z > maxBounds.Z {
			maxBounds.Z = node.Bounds.Max.Z
		}
		return true
	})

	a.inputState.Camera.Birdseye(minBounds, maxBounds)
}

// openWithDefault opens a file or directory with the OS default application.
func (a *App) openWithDefault(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		return
	}
	cmd.Start()
}
