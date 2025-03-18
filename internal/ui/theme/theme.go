package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type SquareTheme struct {
	defaultTheme fyne.Theme
}

func NewSquareTheme() fyne.Theme {
	return &SquareTheme{defaultTheme: theme.DefaultTheme()}
}

func (t *SquareTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return t.PrimaryColor()
	case theme.ColorNameBackground:
		return t.BackgroundColor()
	case theme.ColorNameForeground:
		return t.ForegroundColor()
	case theme.ColorNameButton:
		return t.ButtonColor()
	case theme.ColorNameHover:
		return t.HoverColor()
	case theme.ColorNamePressed:
		return t.PressedColor()
	case theme.ColorNameDisabled:
		return t.DisabledColor()
	case theme.ColorNameDisabledButton:
		return t.DisabledButtonColor()
	case theme.ColorNamePlaceHolder:
		return t.PlaceholderColor()
	case theme.ColorNameScrollBar:
		return t.ScrollBarColor()
	case theme.ColorNameShadow:
		return t.ShadowColor()
	case theme.ColorNameInputBackground:
		return t.InputBackgroundColor()
	case theme.ColorNameOverlayBackground:
		return t.OverlayBackgroundColor()
	case theme.ColorNameMenuBackground:
		return t.MenuBackgroundColor()
	default:
		return t.defaultTheme.Color(name, variant)
	}
}

func (t *SquareTheme) PrimaryColor() color.Color {
	return color.NRGBA{R: 0, G: 180, B: 0, A: 255} // Golf green
}

func (t *SquareTheme) SecondaryColor() color.Color {
	return color.NRGBA{R: 40, G: 40, B: 40, A: 255} // Dark gray
}

func (t *SquareTheme) BackgroundColor() color.Color {
	return color.NRGBA{R: 30, G: 30, B: 30, A: 255} // Dark background
}

func (t *SquareTheme) ForegroundColor() color.Color {
	return color.NRGBA{R: 240, G: 240, B: 240, A: 255} // Light text for better readability
}

func (t *SquareTheme) ButtonColor() color.Color {
	return color.NRGBA{R: 60, G: 60, B: 60, A: 255} // Darker gray for buttons
}

func (t *SquareTheme) HoverColor() color.Color {
	return color.NRGBA{R: 80, G: 80, B: 80, A: 255} // Slightly lighter for hover
}

func (t *SquareTheme) PressedColor() color.Color {
	return color.NRGBA{R: 100, G: 100, B: 100, A: 255} // Even lighter for pressed state
}

func (t *SquareTheme) DisabledColor() color.Color {
	return color.NRGBA{R: 100, G: 100, B: 100, A: 128} // Semi-transparent gray
}

func (t *SquareTheme) DisabledButtonColor() color.Color {
	return color.NRGBA{R: 60, G: 60, B: 60, A: 128} // Semi-transparent button color
}

func (t *SquareTheme) PlaceholderColor() color.Color {
	return color.NRGBA{R: 150, G: 150, B: 150, A: 255} // Medium gray for placeholders
}

func (t *SquareTheme) ScrollBarColor() color.Color {
	return color.NRGBA{R: 60, G: 60, B: 60, A: 255} // Dark gray for scrollbars
}

func (t *SquareTheme) ShadowColor() color.Color {
	return color.NRGBA{R: 0, G: 0, B: 0, A: 50} // Semi-transparent black for shadows
}

func (t *SquareTheme) InputBackgroundColor() color.Color {
	return color.NRGBA{R: 45, G: 45, B: 45, A: 255} // Slightly lighter than background for input fields
}

func (t *SquareTheme) OverlayBackgroundColor() color.Color {
	return color.NRGBA{R: 40, G: 40, B: 40, A: 255} // Dark gray for overlays and dialogs
}

func (t *SquareTheme) MenuBackgroundColor() color.Color {
	return color.NRGBA{R: 45, G: 45, B: 45, A: 255} // Dark background for menus
}

func (t *SquareTheme) MenuTextColor() color.Color {
	return color.NRGBA{R: 240, G: 240, B: 240, A: 255} // Light text for menus
}

func (t *SquareTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.defaultTheme.Font(style)
}

func (t *SquareTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.defaultTheme.Icon(name)
}

func (t *SquareTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.defaultTheme.Size(name)
}
