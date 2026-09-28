package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ModernTheme is a polished dark theme with refined colors, spacing, and corner radii.
type ModernTheme struct{}

var _ fyne.Theme = (*ModernTheme)(nil)

// palette – all colors are defined once here
var (
	// Backgrounds
	colBackground     = color.NRGBA{R: 18, G: 18, B: 22, A: 255} // very dark base
	colSurface        = color.NRGBA{R: 28, G: 28, B: 35, A: 255} // card / panel surface
	colSurfaceVariant = color.NRGBA{R: 38, G: 38, B: 48, A: 255} // slightly lighter surface
	colOverlay        = color.NRGBA{R: 22, G: 22, B: 30, A: 255} // dialog / menu overlay
	colInputBg        = color.NRGBA{R: 32, G: 32, B: 42, A: 255} // entry / select fill

	// Accent colors
	colPrimary      = color.NRGBA{R: 99, G: 102, B: 241, A: 255} // indigo-500
	colPrimaryHover = color.NRGBA{R: 79, G: 82, B: 221, A: 255}  // slightly deeper
	colDanger       = color.NRGBA{R: 239, G: 68, B: 68, A: 255}  // red-500
	colWarning      = color.NRGBA{R: 245, G: 158, B: 11, A: 255} // amber-500
	colSuccess      = color.NRGBA{R: 34, G: 197, B: 94, A: 255}  // green-500

	// Text
	colTextPrimary     = color.NRGBA{R: 237, G: 237, B: 243, A: 255} // near-white
	colTextSecondary   = color.NRGBA{R: 148, G: 148, B: 165, A: 255} // muted
	colTextDisabled    = color.NRGBA{R: 80, G: 80, B: 95, A: 255}    // very muted
	colTextPlaceholder = color.NRGBA{R: 100, G: 100, B: 118, A: 255}

	// Borders / dividers
	colSeparator = color.NRGBA{R: 48, G: 48, B: 62, A: 255}

	// Selection / hover
	colHover     = color.NRGBA{R: 99, G: 102, B: 241, A: 28} // transparent indigo
	colSelection = color.NRGBA{R: 99, G: 102, B: 241, A: 55} // slightly more opaque
	colFocus     = color.NRGBA{R: 99, G: 102, B: 241, A: 180}
)

func (m *ModernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	// Core backgrounds
	case theme.ColorNameBackground:
		return colBackground
	case theme.ColorNameMenuBackground:
		return colOverlay
	case theme.ColorNameOverlayBackground:
		return colOverlay
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 60, B: 80, A: 160}

	// Surfaces / cards
	case theme.ColorNameButton:
		return colSurface
	case theme.ColorNameInputBackground:
		return colInputBg
	case theme.ColorNameHeaderBackground:
		return colSurface

	// Primary accent
	case theme.ColorNamePrimary:
		return colPrimary
	case theme.ColorNameFocus:
		return colFocus
	case theme.ColorNameHover:
		return colHover
	case theme.ColorNameSelection:
		return colSelection

	// Status colors
	case theme.ColorNameSuccess:
		return colSuccess
	case theme.ColorNameWarning:
		return colWarning
	case theme.ColorNameError:
		return colDanger

	// Text
	case theme.ColorNameForeground:
		return colTextPrimary
	case theme.ColorNameDisabled:
		return colTextDisabled
	case theme.ColorNamePlaceHolder:
		return colTextPlaceholder
	case theme.ColorNameSeparator:
		return colSeparator

	// Shadow
	case theme.ColorNameShadow:
		return color.NRGBA{A: 120}
	}
	return theme.DarkTheme().Color(name, variant)
}

func (m *ModernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}

func (m *ModernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}

func (m *ModernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameLineSpacing:
		return 5
	case theme.SizeNameScrollBar:
		return 6
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 6
	}
	return theme.DarkTheme().Size(name)
}
