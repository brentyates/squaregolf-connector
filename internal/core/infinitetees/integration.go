package infinitetees

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/core/simulator"
)

var (
	itInstance *Integration
	itOnce     sync.Once
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
	itOnce.Do(func() {
		itInstance = &Integration{
			stateManager:  stateManager,
			launchMonitor: launchMonitor,
			shotListeners: make([]func(ShotData), 0),
		}
		itInstance.Base = simulator.NewBase(itInstance, host, port)
		itInstance.registerStateListeners()
	})
	return itInstance
}

func (it *Integration) Name() string {
	return "Infinite Tees"
}

func (it *Integration) DefaultPort() int {
	return 999
}

func (it *Integration) GetStateManager() *core.StateManager {
	return it.stateManager
}

func (it *Integration) GetLaunchMonitor() *core.LaunchMonitor {
	return it.launchMonitor
}

func (it *Integration) SetStatus(status simulator.ConnectionStatus) {
	switch status {
	case simulator.StatusDisconnected:
		it.stateManager.SetInfiniteTeesStatus(core.InfiniteTeesStatusDisconnected)
	case simulator.StatusConnecting:
		it.stateManager.SetInfiniteTeesStatus(core.InfiniteTeesStatusConnecting)
	case simulator.StatusConnected:
		it.stateManager.SetInfiniteTeesStatus(core.InfiniteTeesStatusConnected)
	case simulator.StatusError:
		it.stateManager.SetInfiniteTeesStatus(core.InfiniteTeesStatusError)
	}
}

func (it *Integration) SetError(err error) {
	it.stateManager.SetInfiniteTeesError(err)
}

func (it *Integration) OnConnected() {
	log.Printf("[%s] Connected - activating ball detection immediately", it.Name())
	err := it.launchMonitor.ActivateBallDetection()
	if err != nil {
		log.Printf("[%s] Failed to activate ball detection: %v", it.Name(), err)
	}
}

func (it *Integration) OnDisconnected() {
}

func (it *Integration) ProcessMessage(rawMessage string) {
	var baseMsg Message
	if err := json.Unmarshal([]byte(rawMessage), &baseMsg); err != nil {
		log.Printf("[%s] Invalid JSON: %v", it.Name(), err)
		return
	}

	switch baseMsg.Message {
	case "GSPro ready", "IT ready":
		it.handleReadyMessage()
	case "GSPro Player Information", "IT Player Information":
		var playerInfo PlayerInfo
		if err := json.Unmarshal([]byte(rawMessage), &playerInfo); err != nil {
			log.Printf("[%s] Error parsing player info: %v", it.Name(), err)
			return
		}
		it.handlePlayerMessage(&playerInfo)
		it.handleReadyMessage()
	case "Ball Data received", "Club & Ball Data received", "Shot received successfully":
		log.Printf("[%s] Shot data confirmed by server", it.Name())
	default:
		log.Printf("[%s] Unknown message type: %s (full message: %s)", it.Name(), baseMsg.Message, rawMessage)
	}
}

func (it *Integration) handleReadyMessage() {
	err := it.launchMonitor.ActivateBallDetection()
	if err != nil {
		log.Printf("[%s] Failed to activate ball detection: %v", it.Name(), err)
		return
	}
}

func (it *Integration) handlePlayerMessage(playerInfo *PlayerInfo) {
	it.lastPlayerInfo = playerInfo

	if clubName := playerInfo.Player.Club; clubName != "" {
		clubType := it.mapClubToInternal(clubName)
		if clubType != nil {
			log.Printf("[%s] Selected club: %s (mapped to %v)", it.Name(), clubName, clubType)
			it.stateManager.SetClub(clubType)
		} else {
			log.Printf("[%s] Unmapped club: %s", it.Name(), clubName)
		}

		friendlyName := mapClubToFriendlyName(clubName)
		it.stateManager.SetClubName(&friendlyName)
	}

	if handed := playerInfo.Player.Handed; handed != "" {
		var handednessType core.HandednessType
		if handed == "LH" {
			handednessType = core.LeftHanded
			log.Printf("[%s] Selected handedness: Left-handed", it.Name())
		} else {
			handednessType = core.RightHanded
			log.Printf("[%s] Selected handedness: Right-handed", it.Name())
		}
		it.stateManager.SetHandedness(&handednessType)
	}
}

func (it *Integration) sendData(shotData ShotData) error {
	jsonData, err := json.Marshal(shotData)
	if err != nil {
		return err
	}
	return it.Base.SendMessage(jsonData)
}

func (it *Integration) AddShotListener(listener func(ShotData)) {
	it.shotListeners = append(it.shotListeners, listener)
}
