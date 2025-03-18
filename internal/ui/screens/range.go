package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/ui/components"
)

type RangeScreen struct {
	window       fyne.Window
	stateManager *core.StateManager
	metricsCard  *components.MetricsCard
}

func NewRangeScreen(w fyne.Window, stateManager *core.StateManager) *RangeScreen {
	metricsCard := components.NewMetricsCard()
	metricsCard.RegisterCallbacks(stateManager, w)

	return &RangeScreen{
		window:       w,
		stateManager: stateManager,
		metricsCard:  metricsCard,
	}
}

func (rs *RangeScreen) Initialize() {
	// Initialize any necessary state
}

func (rs *RangeScreen) Show() {
	rs.window.SetContent(rs.GetContent())
}

func (rs *RangeScreen) GetContent() fyne.CanvasObject {
	return container.NewVBox(
		rs.metricsCard.GetCard(),
	)
}

func (rs *RangeScreen) OnShow() {
	// TODO: Initialize any necessary state or callbacks
}

func (rs *RangeScreen) OnHide() {
	// TODO: Clean up any resources or callbacks
}
