package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Screen represents a page in the application
type Screen interface {
	Initialize()
	GetContent() fyne.CanvasObject
	OnShow()
	OnHide()
}

// NavigationManager handles screen transitions and organization
type NavigationManager struct {
	window       fyne.Window
	screens      map[string]Screen
	activeScreen string
	content      *fyne.Container
	sidebar      *fyne.Container
}

func NewNavigationManager(window fyne.Window) *NavigationManager {
	return &NavigationManager{
		window:  window,
		screens: make(map[string]Screen),
		content: container.NewMax(),
	}
}

func (nm *NavigationManager) AddScreen(name string, screen Screen) {
	nm.screens[name] = screen
}

func (nm *NavigationManager) ShowScreen(name string) {
	if screen, exists := nm.screens[name]; exists {
		if currentScreen, exists := nm.screens[nm.activeScreen]; exists {
			currentScreen.OnHide()
		}
		screen.OnShow()
		nm.content.Objects = []fyne.CanvasObject{screen.GetContent()}
		nm.activeScreen = name
		nm.window.Canvas().Refresh(nm.content)
		nm.UpdateSidebar() // Refresh sidebar to update active state
	}
}

func (nm *NavigationManager) GetContent() fyne.CanvasObject {
	return nm.content
}

func (nm *NavigationManager) UpdateSidebar() {
	// Create navigation buttons with icons
	alignmentBtn := createNavButton("Alignment", theme.SettingsIcon(), "alignment", nm)
	alignmentBtn.Disable() // Disable alignment button

	rangeBtn := createNavButton("Range", theme.StorageIcon(), "range", nm)
	rangeBtn.Disable() // Disable range button

	buttons := container.NewVBox(
		createNavButton("Dashboard", theme.HomeIcon(), "dashboard", nm),
		alignmentBtn,
		createNavButton("GSPro", theme.ComputerIcon(), "gspro", nm),
		rangeBtn,
		createNavButton("Settings", theme.SettingsIcon(), "settings", nm),
	)

	// Add some padding and spacing
	buttons.Resize(fyne.NewSize(200, buttons.MinSize().Height))
	buttons.Move(fyne.NewPos(0, 0))

	// Create the sidebar container with a subtle background
	nm.sidebar = container.NewHBox(
		container.NewVBox(
			buttons,
		),
		widget.NewSeparator(),
	)
}

func createNavButton(text string, icon fyne.Resource, screen string, nm *NavigationManager) *widget.Button {
	btn := widget.NewButtonWithIcon(text, icon, func() {
		nm.ShowScreen(screen)
	})

	// Style the button based on active state
	if nm.activeScreen == screen {
		btn.Importance = widget.HighImportance
	}

	// Set the button's alignment to left
	btn.Alignment = widget.ButtonAlignLeading

	return btn
}

func (nm *NavigationManager) GetSidebar() *fyne.Container {
	return nm.sidebar
}
