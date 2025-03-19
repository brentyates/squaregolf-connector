package components

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

type StatusCard struct {
	card              *widget.Card
	batteryLabel      *widget.Label
	deviceLabel       *widget.Label
	ballDetectedLabel *widget.Label
	ballReadyLabel    *widget.Label
	ballPositionLabel *widget.Label
	lastErrorLabel    *widget.Label
}

func NewStatusCard() *StatusCard {
	// Create labels with consistent styling
	createLabel := func(text string) *widget.Label {
		label := widget.NewLabel(text)
		label.TextStyle = fyne.TextStyle{Monospace: true}
		return label
	}

	createHeader := func(text string) *widget.Label {
		header := widget.NewLabel(text)
		header.TextStyle = fyne.TextStyle{Bold: true}
		return header
	}

	// Device information section
	deviceInfo := container.NewVBox(
		createHeader("Device Information"),
		widget.NewSeparator(),
		createLabel("Device: N/A"),
		createLabel("Battery: N/A"),
	)

	// Ball status section
	ballStatus := container.NewVBox(
		createHeader("Ball Status"),
		widget.NewSeparator(),
		createLabel("Ball Detected: No"),
		createLabel("Ball Ready: No"),
		createLabel("Ball Position: N/A"),
	)

	// Error section
	errorSection := container.NewVBox(
		createHeader("System Status"),
		widget.NewSeparator(),
		createLabel("Last Error: None"),
	)

	// Create the card with sections
	card := widget.NewCard("Status", "",
		container.NewVBox(
			deviceInfo,
			widget.NewSeparator(),
			ballStatus,
			widget.NewSeparator(),
			errorSection,
		),
	)

	return &StatusCard{
		card:              card,
		batteryLabel:      deviceInfo.Objects[2].(*widget.Label),
		deviceLabel:       deviceInfo.Objects[3].(*widget.Label),
		ballDetectedLabel: ballStatus.Objects[2].(*widget.Label),
		ballReadyLabel:    ballStatus.Objects[3].(*widget.Label),
		ballPositionLabel: ballStatus.Objects[4].(*widget.Label),
		lastErrorLabel:    errorSection.Objects[2].(*widget.Label),
	}
}

func (sc *StatusCard) RegisterCallbacks(stateManager *core.StateManager, window fyne.Window) {
	stateManager.RegisterBatteryLevelCallback(func(oldValue, newValue *int) {
		if newValue != nil {
			sc.batteryLabel.SetText(fmt.Sprintf("Battery: %d%%", *newValue))
		} else {
			sc.batteryLabel.SetText("Battery: N/A")
		}
		window.Canvas().Refresh(sc.batteryLabel)
	})

	stateManager.RegisterDeviceDisplayNameCallback(func(oldValue, newValue *string) {
		if newValue != nil {
			sc.deviceLabel.SetText(fmt.Sprintf("Device: %s", *newValue))
		} else {
			sc.deviceLabel.SetText("Device: N/A")
		}
		window.Canvas().Refresh(sc.deviceLabel)
	})

	stateManager.RegisterBallDetectedCallback(func(oldValue, newValue bool) {
		sc.ballDetectedLabel.SetText(fmt.Sprintf("Ball Detected: %v", newValue))
		window.Canvas().Refresh(sc.ballDetectedLabel)
	})

	stateManager.RegisterBallReadyCallback(func(oldValue, newValue bool) {
		sc.ballReadyLabel.SetText(fmt.Sprintf("Ball Ready: %v", newValue))
		window.Canvas().Refresh(sc.ballReadyLabel)
	})

	stateManager.RegisterBallPositionCallback(func(oldValue, newValue *core.BallPosition) {
		if newValue != nil {
			sc.ballPositionLabel.SetText(fmt.Sprintf("Ball Position: X=%d, Y=%d, Z=%d", newValue.X, newValue.Y, newValue.Z))
		} else {
			sc.ballPositionLabel.SetText("Ball Position: N/A")
		}
		window.Canvas().Refresh(sc.ballPositionLabel)
	})

	stateManager.RegisterLastErrorCallback(func(oldValue, newValue error) {
		if newValue != nil {
			sc.lastErrorLabel.SetText(fmt.Sprintf("Last Error: %v", newValue))
		} else {
			sc.lastErrorLabel.SetText("Last Error: None")
		}
		window.Canvas().Refresh(sc.lastErrorLabel)
	})
}

func (sc *StatusCard) GetCard() *widget.Card {
	return sc.card
}
