package gspro

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/core/simulator"
)

var (
	gsproInstance *Integration
	gsproOnce     sync.Once
)

type Integration struct {
	*simulator.Base
	stateManager   *core.StateManager
	launchMonitor  *core.LaunchMonitor
	shotNumber     int
	lastShotNumber int
	shotListeners  []func(ShotData)
	lastPlayerInfo *PlayerInfo
}

func GetInstance(stateManager *core.StateManager, launchMonitor *core.LaunchMonitor, host string, port int) *Integration {
	gsproOnce.Do(func() {
		gsproInstance = &Integration{
			stateManager:  stateManager,
			launchMonitor: launchMonitor,
			shotListeners: make([]func(ShotData), 0),
		}
		gsproInstance.Base = simulator.NewBase(gsproInstance, host, port)
		gsproInstance.registerStateListeners()
	})
	return gsproInstance
}

func (g *Integration) Name() string {
	return "GSPro"
}

func (g *Integration) DefaultPort() int {
	return 921
}

func (g *Integration) GetStateManager() *core.StateManager {
	return g.stateManager
}

func (g *Integration) GetLaunchMonitor() *core.LaunchMonitor {
	return g.launchMonitor
}

func (g *Integration) SetStatus(status simulator.ConnectionStatus) {
	switch status {
	case simulator.StatusDisconnected:
		g.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
	case simulator.StatusConnecting:
		g.stateManager.SetGSProStatus(core.GSProStatusConnecting)
	case simulator.StatusConnected:
		g.stateManager.SetGSProStatus(core.GSProStatusConnected)
	case simulator.StatusError:
		g.stateManager.SetGSProStatus(core.GSProStatusError)
	}
}

func (g *Integration) SetError(err error) {
	g.stateManager.SetGSProError(err)
}

func (g *Integration) OnConnected() {
	g.shotNumber = 0
	g.lastShotNumber = 0
}

func (g *Integration) OnDisconnected() {
}

func (g *Integration) ProcessMessage(rawMessage string) {
	var baseMsg Message
	if err := json.Unmarshal([]byte(rawMessage), &baseMsg); err != nil {
		log.Printf("Invalid JSON from GSPro: %v", err)
		return
	}

	switch baseMsg.Message {
	case "GSPro ready":
		g.handleGSProReadyMessage()
	case "GSPro Player Information":
		var playerInfo PlayerInfo
		if err := json.Unmarshal([]byte(rawMessage), &playerInfo); err != nil {
			log.Printf("Error parsing player info: %v", err)
			return
		}
		g.handlePlayerMessage(&playerInfo)
		g.handleGSProReadyMessage()
	case "Ball Data received":
		log.Printf("Received ball data confirmation from GSPro")
	case "Club & Ball Data received":
		log.Printf("Received club and ball data confirmation from GSPro")
	default:
		log.Printf("Unknown GSPro message type: %s", baseMsg.Message)
	}
}

func (g *Integration) handleGSProReadyMessage() {
	err := g.launchMonitor.ActivateBallDetection()
	if err != nil {
		log.Printf("Failed to activate ball detection: %v", err)
	}
}

func (g *Integration) handlePlayerMessage(playerInfo *PlayerInfo) {
	g.lastPlayerInfo = playerInfo

	if clubName := playerInfo.Player.Club; clubName != "" {
		clubType := g.mapGSProClubToInternal(clubName)
		if clubType != nil {
			log.Printf("GSPro selected club: %s (mapped to %v)", clubName, clubType)
			g.stateManager.SetClub(clubType)
		} else {
			log.Printf("Unmapped GSPro club: %s", clubName)
		}

		friendlyName := mapGSProClubToFriendlyName(clubName)
		g.stateManager.SetClubName(&friendlyName)
	}

	if handed := playerInfo.Player.Handed; handed != "" {
		var handednessType core.HandednessType
		if handed == "LH" {
			handednessType = core.LeftHanded
			log.Printf("GSPro selected handedness: Left-handed")
		} else {
			handednessType = core.RightHanded
			log.Printf("GSPro selected handedness: Right-handed")
		}
		g.stateManager.SetHandedness(&handednessType)
	}
}

func (g *Integration) sendData(shotData ShotData) error {
	jsonData, err := json.Marshal(shotData)
	if err != nil {
		return err
	}
	return g.Base.SendMessage(jsonData)
}

func (g *Integration) AddShotListener(listener func(ShotData)) {
	g.shotListeners = append(g.shotListeners, listener)
}
