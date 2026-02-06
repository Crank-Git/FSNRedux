package ui

import (
	"os"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
)

var (
	SidebarWidth     int32   = 260
	BreadcrumbHeight int32   = 32
	InfoPanelHeight  int32   = 140
	RowHeight        float32 = 18
	IndentWidth      float32 = 16
	FontSize         float32 = 14
	SmallFontSize    float32 = 11

	AppFont    rl.Font
	fontLoaded bool
)

// LoadFont attempts to load a system TTF font. Call after rl.InitWindow().
func LoadFont() {
	candidates := systemFontPaths()
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			// Load at large atlas size (48px) for crisp rendering at all display sizes.
			// Text is drawn at 11-16px but the high-res atlas + bilinear filter
			// produces sharp glyphs, especially on HiDPI/Retina screens.
			AppFont = rl.LoadFontEx(path, 48, nil, 256)
			if AppFont.BaseSize > 0 {
				rl.SetTextureFilter(AppFont.Texture, rl.FilterBilinear)
				fontLoaded = true
				return
			}
		}
	}
	// Fallback: use raylib default
	AppFont = rl.GetFontDefault()
	fontLoaded = true
}

// UnloadFont frees the loaded font. Call before rl.CloseWindow().
func UnloadFont() {
	if fontLoaded && AppFont.BaseSize > 0 {
		rl.UnloadFont(AppFont)
	}
}

func systemFontPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts/SFNSMono.ttf",
			"/System/Library/Fonts/SFNSText.ttf",
			"/System/Library/Fonts/SFNS.ttf",
			"/System/Library/Fonts/Helvetica.ttc",
			"/System/Library/Fonts/Menlo.ttc",
			"/Library/Fonts/Arial.ttf",
		}
	case "linux":
		return []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
			"/usr/share/fonts/noto/NotoSans-Regular.ttf",
		}
	case "windows":
		return []string{
			`C:\Windows\Fonts\segoeui.ttf`,
			`C:\Windows\Fonts\arial.ttf`,
			`C:\Windows\Fonts\consola.ttf`,
		}
	default:
		return nil
	}
}

// DrawPanel draws a filled rectangle with border.
func DrawPanel(x, y, w, h int32, bg rl.Color) {
	rl.DrawRectangle(x, y, w, h, bg)
	rl.DrawRectangleLines(x, y, w, h, color.BorderColor)
}

// DrawTextUI is a helper that draws text using the loaded font.
func DrawTextUI(text string, x, y int32, size float32, clr rl.Color) {
	rl.DrawTextEx(AppFont, text, rl.NewVector2(float32(x), float32(y)), size, 0.5, clr)
}

// MeasureTextUI measures text width using the loaded font.
func MeasureTextUI(text string, size float32) int32 {
	v := rl.MeasureTextEx(AppFont, text, size, 0.5)
	return int32(v.X)
}
