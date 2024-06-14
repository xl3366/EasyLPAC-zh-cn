package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MyTheme struct{}

func (MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x2e, G: 0x77, B: 0xe6, A: 0xff}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 0x2e, G: 0x77, B: 0xe6, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x08, G: 0x65, B: 0xf5, A: 0x2a}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x08, G: 0x65, B: 0xf5, A: 0x2a}
	default:
		return theme.DefaultTheme().Color(n, v)
	}
}

func (MyTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Italic || s.Symbol {
		return theme.DefaultTheme().Font(s)
	}
	if s.Monospace {
		return resourceDroidSansMono
	}
	return resourceDroidSansFallback
}

func (MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (MyTheme) Size(s fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(s)
}
