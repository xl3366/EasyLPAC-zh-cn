package main

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed assets/DroidSansFallback.ttf
var droidSansFallback []byte

//go:embed assets/DroidSansMono.ttf
var droidSansMono []byte

var resourceDroidSansFallback = &fyne.StaticResource{
	StaticName:    "DroidSansFallback.ttf",
	StaticContent: droidSansFallback,
}

var resourceDroidSansMono = &fyne.StaticResource{
	StaticName:    "DroidSansMono.ttf",
	StaticContent: droidSansMono,
}

