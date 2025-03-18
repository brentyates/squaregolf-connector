package components

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

type ConnectionStatus struct {
	container *fyne.Container
	icon      *widget.Icon
	status    *widget.Label
	window    fyne.Window
}

func NewConnectionStatus() *ConnectionStatus {
	icon := widget.NewIcon(theme.RadioButtonIcon())
	status := widget.NewLabel("Disconnected")
	status.Alignment = fyne.TextAlignCenter
	status.TextStyle = fyne.TextStyle{Bold: true}

	container := container.NewHBox(
		container.NewCenter(icon),
		status,
	)

	return &ConnectionStatus{
		container: container,
		icon:      icon,
		status:    status,
	}
}

func (cs *ConnectionStatus) SetWindow(window fyne.Window) {
	cs.window = window
}

func (cs *ConnectionStatus) RegisterCallback(stateManager *core.StateManager) {
	stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue core.ConnectionStatus) {
		if cs.window == nil {
			return
		}

		go func() {
			cs.window.Canvas().Refresh(cs.status)
			cs.window.Canvas().Refresh(cs.icon)

			switch newValue {
			case core.ConnectionStatusConnected:
				time.Sleep(10 * time.Millisecond) // Small delay to ensure Refresh has processed
				cs.status.SetText("Connected")
				cs.icon.SetResource(theme.RadioButtonCheckedIcon())
			case core.ConnectionStatusDisconnected:
				time.Sleep(10 * time.Millisecond) // Small delay to ensure Refresh has processed
				cs.status.SetText("Disconnected")
				cs.icon.SetResource(theme.RadioButtonIcon())
			case core.ConnectionStatusError:
				cs.status.SetText("Connection Error")
				cs.icon.SetResource(theme.ErrorIcon())
			}
		}()
	})
}

func (cs *ConnectionStatus) GetContainer() *fyne.Container {
	return cs.container
}
