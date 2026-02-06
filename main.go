package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Crank-Git/FSNRedux/internal/app"
)

var version = "dev"

func main() {
	rootPath := flag.String("path", "/", "Root directory to visualize")
	width := flag.Int("width", 1280, "Window width")
	height := flag.Int("height", 800, "Window height")
	depth := flag.Int("depth", 5, "Maximum scan depth (0 = unlimited)")
	theme := flag.String("theme", "", "Color theme: dark, light, or auto (default: auto-detect)")
	showHidden := flag.Bool("hidden", false, "Show hidden files and directories (dotfiles)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("FSNRedux", version)
		return
	}

	// Resolve path
	absPath, err := filepath.Abs(*rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Invalid directory: %s\n", absPath)
		os.Exit(1)
	}

	application := app.New(app.Config{
		RootPath:   absPath,
		Width:      *width,
		Height:     *height,
		MaxDepth:   *depth,
		Theme:      *theme,
		ShowHidden: *showHidden,
	})
	application.Run()
}
