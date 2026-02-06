package color

import (
	"os/exec"
	"runtime"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Theme holds all application colors.
type Theme struct {
	// Directory colors (fsnav: blue)
	DirColor    rl.Color
	DirHover    rl.Color
	DirSelected rl.Color

	// File colors (fsnav: dark red)
	FileColor    rl.Color
	FileHover    rl.Color
	FileSelected rl.Color

	// Special types
	SymlinkColor rl.Color
	OtherColor   rl.Color
	ErrorColor   rl.Color

	// UI chrome
	Background    rl.Color
	SidebarBg     rl.Color
	TextPrimary   rl.Color
	TextSecondary rl.Color
	TextDim       rl.Color
	SelectionBg   rl.Color
	HoverBg       rl.Color
	BorderColor   rl.Color

	// 3D scene
	GridColor      rl.Color
	GridMajor      rl.Color
	WireframeColor rl.Color

	// Accent
	DirAccent  rl.Color // directory name color in sidebar
	LinkAccent rl.Color // current breadcrumb segment
}

var darkTheme = Theme{
	DirColor:    rl.NewColor(60, 160, 170, 255),  // teal directories
	DirHover:    rl.NewColor(80, 200, 210, 255),
	DirSelected: rl.NewColor(100, 220, 230, 255),

	FileColor:    rl.NewColor(200, 155, 60, 255),  // warm amber files
	FileHover:    rl.NewColor(230, 185, 80, 255),
	FileSelected: rl.NewColor(255, 210, 100, 255),

	SymlinkColor: rl.NewColor(170, 130, 210, 255),
	OtherColor:   rl.NewColor(100, 100, 110, 180),
	ErrorColor:   rl.NewColor(220, 70, 70, 255),

	Background:    rl.NewColor(16, 18, 22, 255),
	SidebarBg:     rl.NewColor(22, 24, 30, 255),
	TextPrimary:   rl.NewColor(220, 222, 228, 255),
	TextSecondary: rl.NewColor(140, 145, 158, 255),
	TextDim:       rl.NewColor(75, 80, 92, 255),
	SelectionBg:   rl.NewColor(30, 60, 70, 255),
	HoverBg:       rl.NewColor(30, 34, 44, 255),
	BorderColor:   rl.NewColor(40, 44, 52, 255),

	GridColor:      rl.NewColor(28, 32, 38, 255),
	GridMajor:      rl.NewColor(40, 46, 56, 255),
	WireframeColor: rl.NewColor(100, 220, 230, 255),

	DirAccent:  rl.NewColor(90, 200, 200, 255),   // teal for dir names
	LinkAccent: rl.NewColor(100, 180, 240, 255),
}

var lightTheme = Theme{
	DirColor:    rl.NewColor(40, 130, 140, 255),
	DirHover:    rl.NewColor(30, 110, 120, 255),
	DirSelected: rl.NewColor(20, 150, 160, 255),

	FileColor:    rl.NewColor(180, 120, 40, 255),
	FileHover:    rl.NewColor(200, 140, 50, 255),
	FileSelected: rl.NewColor(220, 160, 60, 255),

	SymlinkColor: rl.NewColor(120, 80, 170, 255),
	OtherColor:   rl.NewColor(150, 150, 155, 200),
	ErrorColor:   rl.NewColor(200, 60, 60, 255),

	Background:    rl.NewColor(242, 242, 245, 255),
	SidebarBg:     rl.NewColor(234, 234, 238, 255),
	TextPrimary:   rl.NewColor(28, 30, 36, 255),
	TextSecondary: rl.NewColor(85, 88, 96, 255),
	TextDim:       rl.NewColor(148, 150, 158, 255),
	SelectionBg:   rl.NewColor(180, 220, 225, 255),
	HoverBg:       rl.NewColor(218, 222, 228, 255),
	BorderColor:   rl.NewColor(198, 200, 208, 255),

	GridColor:      rl.NewColor(218, 220, 226, 255),
	GridMajor:      rl.NewColor(198, 200, 208, 255),
	WireframeColor: rl.NewColor(20, 150, 160, 255),

	DirAccent:  rl.NewColor(30, 130, 135, 255),
	LinkAccent: rl.NewColor(40, 110, 190, 255),
}

// Active is the currently active theme. Set by InitTheme().
var Active = darkTheme

// Convenience accessors (keep backward compat with existing code)
var (
	DirColor       = Active.DirColor
	DirHover       = Active.DirHover
	DirSelected    = Active.DirSelected
	FileColor      = Active.FileColor
	FileHover      = Active.FileHover
	FileSelected   = Active.FileSelected
	SymlinkColor   = Active.SymlinkColor
	OtherColor     = Active.OtherColor
	ErrorColor     = Active.ErrorColor
	Background     = Active.Background
	SidebarBg      = Active.SidebarBg
	TextPrimary    = Active.TextPrimary
	TextSecondary  = Active.TextSecondary
	TextDim        = Active.TextDim
	SelectionBg    = Active.SelectionBg
	HoverBg        = Active.HoverBg
	BorderColor    = Active.BorderColor
	GridColor      = Active.GridColor
	GridMajor      = Active.GridMajor
	WireframeColor = Active.WireframeColor
)

// InitTheme sets the active theme. Pass "" to auto-detect, "dark", or "light".
func InitTheme(preference string) {
	switch preference {
	case "dark":
		Active = darkTheme
	case "light":
		Active = lightTheme
	default:
		if detectDarkMode() {
			Active = darkTheme
		} else {
			Active = lightTheme
		}
	}
	// Update convenience vars
	DirColor = Active.DirColor
	DirHover = Active.DirHover
	DirSelected = Active.DirSelected
	FileColor = Active.FileColor
	FileHover = Active.FileHover
	FileSelected = Active.FileSelected
	SymlinkColor = Active.SymlinkColor
	OtherColor = Active.OtherColor
	ErrorColor = Active.ErrorColor
	Background = Active.Background
	SidebarBg = Active.SidebarBg
	TextPrimary = Active.TextPrimary
	TextSecondary = Active.TextSecondary
	TextDim = Active.TextDim
	SelectionBg = Active.SelectionBg
	HoverBg = Active.HoverBg
	BorderColor = Active.BorderColor
	GridColor = Active.GridColor
	GridMajor = Active.GridMajor
	WireframeColor = Active.WireframeColor
}

func detectDarkMode() bool {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
		if err != nil {
			return false // Light mode (command fails when not set to dark)
		}
		return strings.TrimSpace(string(out)) == "Dark"
	case "linux":
		// Check GTK theme
		out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
		if err == nil && strings.Contains(string(out), "dark") {
			return true
		}
		// Check GTK theme name
		out, err = exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme").Output()
		if err == nil && strings.Contains(strings.ToLower(string(out)), "dark") {
			return true
		}
		return false
	case "windows":
		// Check Windows registry for dark mode
		out, err := exec.Command("reg", "query",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
			"/v", "AppsUseLightTheme").Output()
		if err == nil && strings.Contains(string(out), "0x0") {
			return true // 0 means dark mode
		}
		return false
	default:
		return true // Default to dark
	}
}
