package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
	"github.com/Crank-Git/FSNRedux/internal/fs"
)

// fileTypeEntry maps a file extension to an icon label and category.
type fileTypeEntry struct {
	Icon     string
	Category string
}

var fileTypeMap = map[string]fileTypeEntry{
	// Source code
	".go":    {"Go", "Source Code"},
	".py":    {"Py", "Source Code"},
	".js":    {"JS", "Source Code"},
	".ts":    {"TS", "Source Code"},
	".tsx":   {"TSX", "Source Code"},
	".jsx":   {"JSX", "Source Code"},
	".rs":    {"Rs", "Source Code"},
	".c":     {"C", "Source Code"},
	".cpp":   {"C++", "Source Code"},
	".cc":    {"C++", "Source Code"},
	".h":     {"H", "Header File"},
	".hpp":   {"H++", "Header File"},
	".java":  {"Jv", "Source Code"},
	".kt":    {"Kt", "Source Code"},
	".swift": {"Sw", "Source Code"},
	".rb":    {"Rb", "Source Code"},
	".php":   {"PHP", "Source Code"},
	".cs":    {"C#", "Source Code"},
	".lua":   {"Lua", "Source Code"},
	".zig":   {"Zig", "Source Code"},
	".dart":  {"Drt", "Source Code"},
	".scala": {"Scl", "Source Code"},
	".ex":    {"Ex", "Source Code"},
	".exs":   {"Exs", "Source Code"},
	".erl":   {"Erl", "Source Code"},
	".hs":    {"Hs", "Source Code"},
	".ml":    {"ML", "Source Code"},
	".r":     {"R", "Source Code"},
	".m":     {"OC", "Source Code"},
	// Shell / Scripts
	".sh":   {"Sh", "Shell Script"},
	".bash": {"Sh", "Shell Script"},
	".zsh":  {"Sh", "Shell Script"},
	".fish": {"Sh", "Shell Script"},
	".ps1":  {"PS", "PowerShell Script"},
	".bat":  {"Bat", "Batch Script"},
	// Markup / Config
	".html": {"HTM", "Markup"},
	".htm":  {"HTM", "Markup"},
	".xml":  {"XML", "Markup"},
	".svg":  {"SVG", "Vector Image"},
	".css":  {"CSS", "Stylesheet"},
	".scss": {"SCS", "Stylesheet"},
	".less": {"Les", "Stylesheet"},
	".json": {"JSN", "Data (JSON)"},
	".yaml": {"YML", "Data (YAML)"},
	".yml":  {"YML", "Data (YAML)"},
	".toml": {"TML", "Data (TOML)"},
	".ini":  {"INI", "Configuration"},
	".cfg":  {"CFG", "Configuration"},
	".env":  {"ENV", "Configuration"},
	// Documents
	".md":   {"MD", "Markdown"},
	".txt":  {"TXT", "Plain Text"},
	".rst":  {"RST", "Markup Document"},
	".pdf":  {"PDF", "PDF Document"},
	".doc":  {"DOC", "Word Document"},
	".docx": {"DOC", "Word Document"},
	".xls":  {"XLS", "Spreadsheet"},
	".xlsx": {"XLS", "Spreadsheet"},
	".csv":  {"CSV", "Comma-Separated"},
	".ppt":  {"PPT", "Presentation"},
	".pptx": {"PPT", "Presentation"},
	// Images
	".png":  {"PNG", "Image"},
	".jpg":  {"JPG", "Image"},
	".jpeg": {"JPG", "Image"},
	".gif":  {"GIF", "Image"},
	".bmp":  {"BMP", "Image"},
	".webp": {"WBP", "Image"},
	".ico":  {"ICO", "Icon"},
	".tiff": {"TIF", "Image"},
	// Audio
	".mp3":  {"MP3", "Audio"},
	".wav":  {"WAV", "Audio"},
	".flac": {"FLC", "Audio"},
	".ogg":  {"OGG", "Audio"},
	".aac":  {"AAC", "Audio"},
	".m4a":  {"M4A", "Audio"},
	// Video
	".mp4":  {"MP4", "Video"},
	".mkv":  {"MKV", "Video"},
	".avi":  {"AVI", "Video"},
	".mov":  {"MOV", "Video"},
	".webm": {"WBM", "Video"},
	".wmv":  {"WMV", "Video"},
	// Archives
	".zip":  {"ZIP", "Archive"},
	".tar":  {"TAR", "Archive"},
	".gz":   {"GZ", "Archive"},
	".bz2":  {"BZ2", "Archive"},
	".xz":   {"XZ", "Archive"},
	".7z":   {"7Z", "Archive"},
	".rar":  {"RAR", "Archive"},
	".zst":  {"ZST", "Archive"},
	// Binary / Executable
	".exe":  {"EXE", "Executable"},
	".dll":  {"DLL", "Library"},
	".so":   {"SO", "Shared Library"},
	".dylib": {"DYL", "Shared Library"},
	".bin":  {"BIN", "Binary"},
	".o":    {"OBJ", "Object File"},
	".a":    {"LIB", "Static Library"},
	".wasm": {"WSM", "WebAssembly"},
	// Database
	".db":     {"DB", "Database"},
	".sqlite": {"SQL", "Database"},
	".sql":    {"SQL", "SQL Script"},
	// Build / Lock
	".lock": {"LCK", "Lock File"},
	".sum":  {"SUM", "Checksum"},
	".mod":  {"MOD", "Module File"},
}

// FileTypeIcon returns a short icon label and category for a filename.
func FileTypeIcon(name string, isDir bool) (icon string, category string) {
	if isDir {
		return "DIR", "Directory"
	}
	ext := strings.ToLower(filepath.Ext(name))
	if entry, ok := fileTypeMap[ext]; ok {
		return entry.Icon, entry.Category
	}
	if ext != "" {
		return strings.ToUpper(strings.TrimPrefix(ext, ".")), "File"
	}
	return "---", "File"
}

// FileTypeIconColor returns a color for the file type icon badge.
func FileTypeIconColor(icon string) rl.Color {
	switch icon {
	case "Go":
		return rl.NewColor(0, 173, 216, 255)   // cyan
	case "Py":
		return rl.NewColor(55, 118, 171, 255)   // blue
	case "JS", "JSX":
		return rl.NewColor(247, 223, 30, 255)   // yellow
	case "TS", "TSX":
		return rl.NewColor(49, 120, 198, 255)   // blue
	case "Rs":
		return rl.NewColor(222, 165, 132, 255)  // rust orange
	case "C", "C++", "H", "H++":
		return rl.NewColor(85, 85, 255, 255)    // blue
	case "Jv":
		return rl.NewColor(248, 152, 32, 255)   // java orange
	case "Rb":
		return rl.NewColor(204, 52, 45, 255)    // ruby red
	case "Sh":
		return rl.NewColor(78, 154, 6, 255)     // green
	case "HTM", "CSS", "SCS":
		return rl.NewColor(228, 77, 38, 255)    // html orange
	case "MD", "TXT", "RST":
		return rl.NewColor(180, 180, 180, 255)  // light gray
	case "JSN", "YML", "TML", "XML":
		return rl.NewColor(160, 160, 80, 255)   // olive
	case "PNG", "JPG", "GIF", "BMP", "SVG", "WBP", "ICO":
		return rl.NewColor(140, 200, 60, 255)   // green
	case "MP3", "WAV", "FLC", "OGG", "AAC", "M4A":
		return rl.NewColor(230, 126, 34, 255)   // orange
	case "MP4", "MKV", "AVI", "MOV", "WBM":
		return rl.NewColor(155, 89, 182, 255)   // purple
	case "ZIP", "TAR", "GZ", "RAR", "7Z":
		return rl.NewColor(127, 140, 141, 255)  // gray
	case "PDF", "DOC", "XLS", "PPT":
		return rl.NewColor(192, 57, 43, 255)    // dark red
	case "DIR":
		return rl.NewColor(255, 193, 7, 255)    // amber
	default:
		return rl.NewColor(149, 165, 166, 255)  // silver
	}
}

// DrawInfoPanel renders file/directory info at the bottom of the sidebar.
func DrawInfoPanel(entry *fs.Entry, screenHeight int32) {
	panelX := int32(0)
	panelY := screenHeight - InfoPanelHeight
	panelW := SidebarWidth
	panelH := InfoPanelHeight

	DrawPanel(panelX, panelY, panelW, panelH, color.SidebarBg)

	// Top separator line
	rl.DrawRectangle(panelX+8, panelY, panelW-16, 1, color.BorderColor)

	if entry == nil {
		DrawTextUI("No selection", panelX+8, panelY+12, FontSize, color.TextDim)
		DrawTextUI("Click a file or directory", panelX+8, panelY+32, SmallFontSize, color.TextDim)
		return
	}

	y := panelY + 6

	// Icon badge + name
	icon, _ := FileTypeIcon(entry.Name, entry.IsDir())
	if entry.Type == fs.TypeSymlink {
		icon = "LNK"
	}
	badgeW := drawIconBadge(icon, panelX+8, y)

	nameColor := color.TextPrimary
	if entry.Type == fs.TypeDir {
		nameColor = color.Active.DirAccent
	}
	name := entry.Name
	if len(name) > 26 {
		name = name[:24] + ".."
	}
	DrawTextUI(name, panelX+8+badgeW+6, y, FontSize, nameColor)
	y += 18

	// Full path (truncated)
	pathStr := entry.Path
	maxPathChars := int((float32(panelW) - 16) / 7)
	if len(pathStr) > maxPathChars {
		pathStr = "..." + pathStr[len(pathStr)-maxPathChars+3:]
	}
	DrawTextUI(pathStr, panelX+8, y, SmallFontSize, color.TextDim)
	y += 14

	// Type + extension
	typeStr := entry.Type.String()
	if entry.Type == fs.TypeFile {
		ext := filepath.Ext(entry.Name)
		if ext != "" {
			typeStr = fmt.Sprintf("%s (%s)", typeStr, ext)
		}
	}
	DrawTextUI(typeStr, panelX+8, y, SmallFontSize, color.TextSecondary)

	// Size on the right
	sizeStr := FormatSize(entry.Size)
	sizeW := MeasureTextUI(sizeStr, SmallFontSize)
	DrawTextUI(sizeStr, panelW-sizeW-8, y, SmallFontSize, color.TextSecondary)
	y += 14

	// Modified + age on same line
	if !entry.ModTime.IsZero() {
		modStr := entry.ModTime.Format("2006-01-02 15:04")
		DrawTextUI(modStr, panelX+8, y, SmallFontSize, color.TextSecondary)

		age := time.Since(entry.ModTime)
		ageStr := formatAge(age)
		ageW := MeasureTextUI(ageStr, SmallFontSize)
		DrawTextUI(ageStr, panelW-ageW-8, y, SmallFontSize, color.TextDim)
		y += 14
	}

	// For directories: children info
	if entry.Type == fs.TypeDir {
		dirs := 0
		files := 0
		for _, child := range entry.Children {
			if child.Type == fs.TypeDir {
				dirs++
			} else {
				files++
			}
		}
		childStr := fmt.Sprintf("%d dirs, %d files", dirs, files)
		if !entry.Loaded {
			childStr = "not expanded"
		}
		DrawTextUI(childStr, panelX+8, y, SmallFontSize, color.TextSecondary)
	}
}

// DrawSelectedTooltip renders a floating info card near a selected 3D node.
func DrawSelectedTooltip(entry *fs.Entry, screenX, screenY float32) {
	if entry == nil {
		return
	}

	// Position below and to the right of the node's screen position
	tx := int32(screenX) + 12
	ty := int32(screenY) + 12

	// Build info lines
	name := entry.Name
	if len(name) > 24 {
		name = name[:22] + ".."
	}
	icon, _ := FileTypeIcon(entry.Name, entry.IsDir())
	if entry.Type == fs.TypeSymlink {
		icon = "LNK"
	}
	sizeStr := FormatSize(entry.Size)
	line2 := fmt.Sprintf("%s  %s", entry.Type.String(), sizeStr)

	// Measure (account for badge + gap + name)
	badgeTextW := MeasureTextUI(icon, SmallFontSize) + 12 + 8 // padding + gap
	w1 := badgeTextW + MeasureTextUI(name, FontSize)
	w2 := MeasureTextUI(line2, SmallFontSize)
	boxW := w1
	if w2 > boxW {
		boxW = w2
	}
	boxW += 16
	boxH := int32(36)

	// Keep on screen
	sw := int32(rl.GetScreenWidth())
	sh := int32(rl.GetScreenHeight())
	if tx+boxW > sw-4 {
		tx = sw - boxW - 4
	}
	if ty+boxH > sh-4 {
		ty = int32(screenY) - boxH - 4
	}

	// Draw card
	rl.DrawRectangle(tx, ty, boxW, boxH, rl.NewColor(
		color.Active.SidebarBg.R,
		color.Active.SidebarBg.G,
		color.Active.SidebarBg.B,
		230,
	))
	rl.DrawRectangleLines(tx, ty, boxW, boxH, color.BorderColor)

	nameColor := color.TextPrimary
	if entry.Type == fs.TypeDir {
		nameColor = color.Active.DirAccent
	}
	bw := drawIconBadge(icon, tx+8, ty+4)
	DrawTextUI(name, tx+8+bw+6, ty+4, FontSize, nameColor)
	DrawTextUI(line2, tx+8, ty+20, SmallFontSize, color.TextSecondary)
}

// formatAge converts a duration to a human-readable age string.
func formatAge(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months", int(d.Hours()/(24*30)))
	default:
		years := d.Hours() / (24 * 365)
		return fmt.Sprintf("%.1f years", years)
	}
}

// drawIconBadge draws a colored badge with a short label at (bx, by).
// Returns the width consumed so the caller can offset the name text.
func drawIconBadge(icon string, bx, by int32) int32 {
	badgeColor := FileTypeIconColor(icon)
	padding := int32(6)
	fontSize := SmallFontSize + 1
	textW := MeasureTextUI(icon, fontSize)
	badgeW := textW + padding*2
	badgeH := int32(18)

	// Badge background with border
	rl.DrawRectangle(bx, by, badgeW, badgeH, badgeColor)
	rl.DrawRectangleLines(bx, by, badgeW, badgeH, rl.NewColor(
		badgeColor.R/2, badgeColor.G/2, badgeColor.B/2, 255,
	))
	// Badge text (dark on bright, white on dark)
	lum := int(badgeColor.R)*299 + int(badgeColor.G)*587 + int(badgeColor.B)*114
	textClr := rl.NewColor(0, 0, 0, 255)
	if lum < 128000 {
		textClr = rl.NewColor(255, 255, 255, 255)
	}
	DrawTextUI(icon, bx+padding, by+3, fontSize, textClr)
	return badgeW
}

// DrawInspectPanel renders a centered overlay with detailed file/directory info.
func DrawInspectPanel(info *fs.InspectInfo, screenW, screenH int32) {
	if info == nil {
		return
	}

	panelW := int32(400)
	panelH := int32(280)
	if !info.IsDir {
		panelH = 220
	}
	panelX := (screenW - panelW) / 2
	panelY := (screenH - panelH) / 2

	// Dimmed background
	rl.DrawRectangle(0, 0, screenW, screenH, rl.NewColor(0, 0, 0, 120))

	// Panel with accent border
	rl.DrawRectangle(panelX, panelY, panelW, panelH, color.SidebarBg)
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, color.Active.LinkAccent)

	x := panelX + 16
	y := panelY + 12

	// Icon badge + Name
	icon, category := FileTypeIcon(info.Name, info.IsDir)
	badgeW := drawIconBadge(icon, x, y)

	name := info.Name
	if len(name) > 36 {
		name = name[:34] + ".."
	}
	nameColor := color.TextPrimary
	if info.IsDir {
		nameColor = color.Active.DirAccent
	}
	DrawTextUI(name, x+badgeW+8, y, FontSize+2, nameColor)
	y += 22

	// Category subtitle
	DrawTextUI(category, x, y, SmallFontSize, color.TextDim)
	y += 16
	rl.DrawRectangle(x, y, panelW-32, 1, color.BorderColor)
	y += 8

	// Helper to draw a label: value row
	drawRow := func(label, value string) {
		DrawTextUI(label, x, y, SmallFontSize, color.TextDim)
		valW := MeasureTextUI(value, SmallFontSize)
		DrawTextUI(value, panelX+panelW-16-valW, y, SmallFontSize, color.TextSecondary)
		y += 18
	}

	// Path (may need truncation)
	pathStr := info.Path
	maxChars := int((float32(panelW) - 32) / 6.5)
	if len(pathStr) > maxChars {
		pathStr = "..." + pathStr[len(pathStr)-maxChars+3:]
	}
	DrawTextUI(pathStr, x, y, SmallFontSize, color.TextDim)
	y += 20

	if !info.IsDir && info.Extension != "" {
		drawRow("Extension:", info.Extension)
	}

	drawRow("Size:", FormatSize(info.Size))
	drawRow("Permissions:", info.Perms)

	if !info.ModTime.IsZero() {
		drawRow("Modified:", info.ModTime.Format("2006-01-02 15:04:05"))
		age := time.Since(info.ModTime)
		drawRow("Age:", formatAge(age))
	}

	if info.IsDir {
		if info.Loaded {
			drawRow("Files:", fmt.Sprintf("%d", info.FileCount))
			drawRow("Subdirectories:", fmt.Sprintf("%d", info.DirCount))
			drawRow("Direct children:", fmt.Sprintf("%d", info.ChildCount))
		} else {
			drawRow("Children:", "not expanded")
		}
	}

	// Dismiss hint
	y = panelY + panelH - 20
	hint := "Press Space or Escape to close"
	hintW := MeasureTextUI(hint, SmallFontSize)
	DrawTextUI(hint, panelX+(panelW-hintW)/2, y, SmallFontSize, color.TextDim)
}

// DrawModeIndicator draws the current visualization mode in the corner with a background pill.
func DrawModeIndicator(mode string, screenWidth int32) {
	text := fmt.Sprintf("Mode: %s", mode)
	textWidth := MeasureTextUI(text, FontSize)
	x := screenWidth - textWidth - 20
	y := int32(BreadcrumbHeight + 8)
	// Background pill
	rl.DrawRectangle(x-6, y-3, textWidth+12, 20, rl.NewColor(
		color.Active.SidebarBg.R,
		color.Active.SidebarBg.G,
		color.Active.SidebarBg.B,
		220,
	))
	rl.DrawRectangleLines(x-6, y-3, textWidth+12, 20, color.BorderColor)
	DrawTextUI(text, x, y, FontSize, color.TextSecondary)
}

// DrawScanProgress shows scanning progress overlay.
func DrawScanProgress(dirsScanned, filesFound int64, bytesTotal int64, screenWidth, screenHeight int32) {
	text := fmt.Sprintf("Scanning... %d dirs, %d files (%s)",
		dirsScanned, filesFound, FormatSize(bytesTotal))
	textWidth := MeasureTextUI(text, FontSize+2)
	x := (screenWidth - textWidth) / 2
	y := screenHeight / 2

	// Background box
	rl.DrawRectangle(x-12, y-8, textWidth+24, 32, rl.NewColor(20, 20, 25, 220))
	rl.DrawRectangleLines(x-12, y-8, textWidth+24, 32, color.BorderColor)
	DrawTextUI(text, x, y, FontSize+2, color.TextPrimary)
}

// DrawHelpText shows keyboard shortcuts in a readable panel.
func DrawHelpText(screenWidth, screenHeight int32) {
	lines := []struct {
		key  string
		desc string
	}{
		{"Mouse L-drag", "Orbit camera"},
		{"Mouse R-drag", "Zoom"},
		{"Scroll / +/-", "Zoom in/out"},
		{"WASD / Arrows", "Pan camera"},
		{"Click", "Select node"},
		{"Double-click", "Expand/collapse dir"},
		{"Enter", "Expand selected dir"},
		{"Space", "Inspect dir / preview file"},
		{"O", "Open with default app"},
		{"Escape", "Collapse dir / parent"},
		{"Tab / Shift+Tab", "Next / prev node"},
		{"Home", "Go to root"},
		{"F", "Search"},
		{"Ctrl+L", "Go to path"},
		{"N / P", "Next / prev search result"},
		{"B", "Birdseye view"},
		{",", "Settings"},
		{"H", "Toggle this help"},
	}

	// Panel dimensions
	panelW := int32(260)
	panelH := int32(len(lines)*16 + 20)
	panelX := screenWidth - panelW - 8
	panelY := screenHeight - panelH - 8

	// Background panel
	rl.DrawRectangle(panelX, panelY, panelW, panelH, rl.NewColor(
		color.Active.SidebarBg.R,
		color.Active.SidebarBg.G,
		color.Active.SidebarBg.B,
		230,
	))
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, color.BorderColor)

	// Title
	DrawTextUI("Keyboard Shortcuts", panelX+8, panelY+4, SmallFontSize, color.TextSecondary)
	rl.DrawRectangle(panelX+8, panelY+16, panelW-16, 1, color.BorderColor)

	y := panelY + 20
	for _, line := range lines {
		// Key in accent color
		DrawTextUI(line.key, panelX+8, y, SmallFontSize, color.Active.LinkAccent)
		// Description aligned right-ish
		descW := MeasureTextUI(line.desc, SmallFontSize)
		DrawTextUI(line.desc, panelX+panelW-descW-8, y, SmallFontSize, color.TextDim)
		y += 16
	}
}
