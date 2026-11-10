// SuMate (东坡嘴替) — a progressive spiritual mining system built on Su
// Dongpo's ci poetry.
//
// The desktop application is a Wails v3 shell around the SuMate service.
// All the domain logic lives under internal/; this file only wires the
// service into the window.
package main

import (
	"embed"
	"log"

	"github.com/incurl/sumate/internal/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails embeds the built frontend so the binary is self-contained.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc, err := service.New()
	if err != nil {
		log.Fatalf("sumate: cannot start the service: %v", err)
	}

	app := application.New(application.Options{
		Name:        "SuMate 东坡嘴替",
		Description: "东坡嘴替 — 逐层挖掘最懂你此刻的那句东坡词",
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "SuMate 东坡嘴替",
		Width:  1280,
		Height: 860,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(28, 26, 24),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
