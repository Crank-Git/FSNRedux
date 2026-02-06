package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
)

// PreviewState holds the state for the file preview panel.
type PreviewState struct {
	Open     bool
	FilePath string
	FileName string
	Kind     PreviewKind

	// Text preview
	Lines      []string
	ScrollY    int
	TotalLines int

	// Image preview
	Texture   rl.Texture2D
	ImgWidth  int32
	ImgHeight int32
	ImgLoaded bool
}

// PreviewKind distinguishes what type of preview to show.
type PreviewKind int

const (
	PreviewNone PreviewKind = iota
	PreviewText
	PreviewImage
	PreviewUnsupported
)

// maxPreviewLines limits how many lines we read from text files.
const maxPreviewLines = 500

// maxPreviewBytes limits how many bytes we read (1MB).
const maxPreviewBytes = 1024 * 1024

// textExtensions are extensions we treat as text-previewable.
var textExtensions = map[string]bool{
	".txt": true, ".md": true, ".go": true, ".py": true, ".js": true,
	".ts": true, ".tsx": true, ".jsx": true, ".rs": true, ".c": true,
	".cpp": true, ".h": true, ".hpp": true, ".java": true, ".rb": true,
	".sh": true, ".bash": true, ".zsh": true, ".json": true, ".yaml": true,
	".yml": true, ".toml": true, ".xml": true, ".html": true, ".css": true,
	".scss": true, ".sql": true, ".lua": true, ".swift": true, ".kt": true,
	".ini": true, ".cfg": true, ".env": true, ".csv": true, ".log": true,
	".mod": true, ".sum": true, ".lock": true, ".gitignore": true,
	".dockerfile": true, ".makefile": true, ".cmake": true, ".bat": true,
	".ps1": true, ".r": true, ".m": true, ".ex": true, ".exs": true,
	".erl": true, ".hs": true, ".ml": true, ".rst": true, ".tex": true,
	".vim": true, ".conf": true, ".properties": true,
}

// imageExtensions are extensions we treat as image-previewable.
var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".bmp": true, ".gif": true,
	".tga": true, ".hdr": true, ".psd": true,
}

// classifyPreview determines what kind of preview is appropriate for a file.
func classifyPreview(name string) PreviewKind {
	ext := strings.ToLower(filepath.Ext(name))
	if imageExtensions[ext] {
		return PreviewImage
	}
	if textExtensions[ext] {
		return PreviewText
	}
	// Try no-extension files as text (e.g. Makefile, Dockerfile)
	baseLower := strings.ToLower(filepath.Base(name))
	if ext == "" || baseLower == "makefile" || baseLower == "dockerfile" || baseLower == "readme" || baseLower == "license" {
		return PreviewText
	}
	return PreviewUnsupported
}

// OpenPreview loads a file for preview.
func (p *PreviewState) OpenPreview(path string) {
	p.Close() // clean up any previous preview

	p.FilePath = path
	p.FileName = filepath.Base(path)
	p.Kind = classifyPreview(path)
	p.ScrollY = 0

	switch p.Kind {
	case PreviewText:
		p.loadText(path)
	case PreviewImage:
		p.loadImage(path)
	case PreviewUnsupported:
		p.Lines = []string{"Preview not available for this file type.", "", "Press O to open with default application."}
		p.TotalLines = 3
	}

	p.Open = true
}

func (p *PreviewState) loadText(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		p.Lines = []string{"Error reading file: " + err.Error()}
		p.TotalLines = 1
		return
	}
	if len(data) > maxPreviewBytes {
		data = data[:maxPreviewBytes]
	}
	allLines := strings.Split(string(data), "\n")
	if len(allLines) > maxPreviewLines {
		allLines = allLines[:maxPreviewLines]
	}
	p.Lines = allLines
	p.TotalLines = len(allLines)
}

func (p *PreviewState) loadImage(path string) {
	img := rl.LoadImage(path)
	if img.Width == 0 || img.Height == 0 {
		p.Kind = PreviewUnsupported
		p.Lines = []string{"Failed to load image."}
		p.TotalLines = 1
		return
	}
	// Scale down large images to fit preview
	maxDim := int32(512)
	if img.Width > maxDim || img.Height > maxDim {
		if img.Width > img.Height {
			rl.ImageResize(img, maxDim, img.Height*maxDim/img.Width)
		} else {
			rl.ImageResize(img, img.Width*maxDim/img.Height, maxDim)
		}
	}
	p.Texture = rl.LoadTextureFromImage(img)
	p.ImgWidth = img.Width
	p.ImgHeight = img.Height
	p.ImgLoaded = true
	rl.UnloadImage(img)
}

// Close cleans up preview state and unloads resources.
func (p *PreviewState) Close() {
	if p.ImgLoaded {
		rl.UnloadTexture(p.Texture)
		p.ImgLoaded = false
	}
	p.Open = false
	p.Lines = nil
	p.TotalLines = 0
	p.ScrollY = 0
	p.Kind = PreviewNone
}

// Update handles scroll input. Returns true if preview should close.
func (p *PreviewState) Update() bool {
	if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeySpace) {
		return true
	}

	// Scroll for text preview
	if p.Kind == PreviewText {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			p.ScrollY -= int(wheel * 3)
		}
		if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyJ) {
			p.ScrollY += 3
		}
		if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyK) {
			p.ScrollY -= 3
		}
		if rl.IsKeyPressed(rl.KeyPageDown) {
			p.ScrollY += 20
		}
		if rl.IsKeyPressed(rl.KeyPageUp) {
			p.ScrollY -= 20
		}
		// Clamp
		if p.ScrollY < 0 {
			p.ScrollY = 0
		}
		maxScroll := p.TotalLines - 20
		if maxScroll < 0 {
			maxScroll = 0
		}
		if p.ScrollY > maxScroll {
			p.ScrollY = maxScroll
		}
	}
	return false
}

// DrawPreviewPanel renders the file preview overlay.
func DrawPreviewPanel(p *PreviewState, screenW, screenH int32) {
	if p == nil || !p.Open {
		return
	}

	panelW := screenW * 2 / 3
	if panelW < 400 {
		panelW = 400
	}
	if panelW > 800 {
		panelW = 800
	}
	panelH := screenH * 3 / 4
	if panelH < 300 {
		panelH = 300
	}
	panelX := (screenW - panelW) / 2
	panelY := (screenH - panelH) / 2

	// Dimmed background
	rl.DrawRectangle(0, 0, screenW, screenH, rl.NewColor(0, 0, 0, 140))

	// Panel
	rl.DrawRectangle(panelX, panelY, panelW, panelH, color.SidebarBg)
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, color.Active.LinkAccent)

	// Title bar
	titleH := int32(28)
	name := p.FileName
	if len(name) > 50 {
		name = name[:48] + ".."
	}
	icon, _ := FileTypeIcon(name, false)
	badgeW := drawIconBadge(icon, panelX+10, panelY+6)
	DrawTextUI(name, panelX+10+badgeW+8, panelY+7, FontSize, color.TextPrimary)
	rl.DrawRectangle(panelX+8, panelY+titleH, panelW-16, 1, color.BorderColor)

	contentX := panelX + 12
	contentY := panelY + titleH + 6
	contentW := panelW - 24
	contentH := panelH - titleH - 30

	switch p.Kind {
	case PreviewText:
		drawTextPreview(p, contentX, contentY, contentW, contentH)
	case PreviewImage:
		drawImagePreview(p, contentX, contentY, contentW, contentH)
	default:
		for i, line := range p.Lines {
			DrawTextUI(line, contentX, contentY+int32(i)*16, SmallFontSize, color.TextDim)
		}
	}

	// Bottom hint
	hint := "Scroll: wheel/arrows  |  Space/Esc: close  |  O: open in app"
	hintW := MeasureTextUI(hint, SmallFontSize)
	DrawTextUI(hint, panelX+(panelW-hintW)/2, panelY+panelH-16, SmallFontSize, color.TextDim)
}

func drawTextPreview(p *PreviewState, x, y, w, h int32) {
	visibleLines := int(h / 14)
	lineH := int32(14)

	// Line number gutter width
	gutterW := int32(36)

	// Clip region (basic: just skip lines outside)
	for i := 0; i < visibleLines && p.ScrollY+i < p.TotalLines; i++ {
		lineIdx := p.ScrollY + i
		ly := y + int32(i)*lineH

		// Line number
		lnStr := fmt.Sprintf("%4d", lineIdx+1)
		DrawTextUI(lnStr, x, ly, SmallFontSize, color.TextDim)

		// Line content (truncate if too long)
		line := p.Lines[lineIdx]
		// Replace tabs with spaces
		line = strings.ReplaceAll(line, "\t", "    ")
		maxChars := int((float32(w) - float32(gutterW)) / 6.2)
		if len(line) > maxChars {
			line = line[:maxChars]
		}
		DrawTextUI(line, x+gutterW, ly, SmallFontSize, color.TextSecondary)
	}

	// Scrollbar
	if p.TotalLines > visibleLines {
		barH := h
		thumbH := barH * int32(visibleLines) / int32(p.TotalLines)
		if thumbH < 10 {
			thumbH = 10
		}
		thumbY := y + barH*int32(p.ScrollY)/int32(p.TotalLines)
		barX := x + w - 4
		rl.DrawRectangle(barX, y, 4, barH, rl.NewColor(60, 60, 60, 100))
		rl.DrawRectangle(barX, thumbY, 4, thumbH, color.Active.LinkAccent)
	}
}

func drawImagePreview(p *PreviewState, x, y, w, h int32) {
	if !p.ImgLoaded {
		DrawTextUI("Failed to load image", x, y, FontSize, color.TextDim)
		return
	}

	// Scale image to fit content area while preserving aspect ratio
	imgW := float32(p.ImgWidth)
	imgH := float32(p.ImgHeight)
	scaleX := float32(w) / imgW
	scaleY := float32(h) / imgH
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	if scale > 1.0 {
		scale = 1.0 // don't upscale
	}

	drawW := imgW * scale
	drawH := imgH * scale
	drawX := float32(x) + (float32(w)-drawW)/2
	drawY := float32(y) + (float32(h)-drawH)/2

	rl.DrawTextureEx(p.Texture, rl.NewVector2(drawX, drawY), 0, scale, rl.White)

	// Image dimensions label
	dimStr := fmt.Sprintf("%dx%d", p.ImgWidth, p.ImgHeight)
	dimW := MeasureTextUI(dimStr, SmallFontSize)
	DrawTextUI(dimStr, x+w-dimW, y+h-14, SmallFontSize, color.TextDim)
}
