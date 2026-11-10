//go:build !ios && !android && !darwin && !windows && !linux

// This file lives under build/ios but carries the non-iOS build tag so that
// `go build ./...` on a desktop host does not try to link it as a main
// package.  It is only ever compiled as part of the iOS application build.

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// modifyOptionsForIOS is a no-op on non-iOS platforms
func modifyOptionsForIOS(opts *application.Options) {
	// No modifications needed for non-iOS platforms
}