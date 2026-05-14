package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "todo-app",
		Width:     280,
		Height:    420,
		MinWidth:  280,
		MinHeight: 420,
		MaxWidth:  280,
		MaxHeight: 420,

		Frameless:   true,
		AlwaysOnTop: true,

		// Transparent background lets the frontend control rounded corners
		// and widget opacity via CSS.
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},

		// Hide until startup() has positioned the window correctly.
		StartHidden: true,

		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		OnStartup: app.startup,

		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
