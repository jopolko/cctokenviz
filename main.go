package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	migrateLegacyData()
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "CCTokenViz",
		Width:     240,
		Height:    620,
		MinWidth:  200,
		MinHeight: 150,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 27, B: 27, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		OnDomReady:       app.domReady,
		Bind: []interface{}{
			app,
		},
		// Always on top floating window
		AlwaysOnTop: true,
		// Frameless for cleaner look (optional - comment out if you want standard window)
		// Frameless: true,
		// Mac: &mac.Options{
		// 	TitleBar: &mac.TitleBar{
		// 		TitlebarAppearsTransparent: true,
		// 		HideTitle:                  true,
		// 		HideTitleBar:               false,
		// 		FullSizeContent:            true,
		// 	},
		// 	WebviewIsTransparent: false,
		// 	WindowIsTranslucent:  false,
		// },
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			// The title bar follows Windows' light/dark setting until a theme is applied
			Theme: windows.SystemDefault,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
