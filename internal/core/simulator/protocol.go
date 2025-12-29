package simulator

import "github.com/brentyates/squaregolf-connector/internal/core"

type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

type Protocol interface {
	Name() string
	DefaultPort() int
	ProcessMessage(rawMessage string)
	OnConnected()
	OnDisconnected()
	SetStatus(status ConnectionStatus)
	SetError(err error)
	GetStateManager() *core.StateManager
	GetLaunchMonitor() *core.LaunchMonitor
}
