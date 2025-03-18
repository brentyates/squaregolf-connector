package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

type MetricsCard struct {
	card *widget.Card
}

func NewMetricsCard() *MetricsCard {
	// Create an empty card
	card := widget.NewCard("", "", container.NewVBox())

	return &MetricsCard{
		card: card,
	}
}

func (mc *MetricsCard) RegisterCallbacks(stateManager *core.StateManager, window fyne.Window) {
	// Keep the callbacks registered but don't display anything
	stateManager.RegisterLastBallMetricsCallback(func(oldValue, newValue *core.BallMetrics) {
		// Metrics are still being collected but not displayed
	})

	stateManager.RegisterLastClubMetricsCallback(func(oldValue, newValue *core.ClubMetrics) {
		// Metrics are still being collected but not displayed
	})
}

func (mc *MetricsCard) GetCard() *widget.Card {
	return mc.card
}
