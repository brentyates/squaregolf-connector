package gspro

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

// processMessage processes a message from GSPro
func (g *Integration) processMessage(rawMessage string) {
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
		// Acknowledge ball data received message
		log.Printf("Received ball data confirmation from GSPro")
	case "Club & Ball Data received":
		// Acknowledge the club and ball data received message
		log.Printf("Received club and ball data confirmation from GSPro")
	default:
		log.Printf("Unknown GSPro message type: %s", baseMsg.Message)
	}
}

// handleGSProReadyMessage handles the GSPro ready message and activates ball detection if in manual mode
func (g *Integration) handleGSProReadyMessage() {
	// Activate ball detection using the launch monitor
	// This will send the appropriate commands to the device to enter ball detection mode
	// The device will then wait for a ball to be placed, become ready, and be hit
	err := g.launchMonitor.ActivateBallDetection()
	if err != nil {
		log.Printf("Failed to activate ball detection: %v", err)
		return
	}
}

// handlePlayerMessage handles player message from GSPro
func (g *Integration) handlePlayerMessage(playerInfo *PlayerInfo) {
	g.lastPlayerInfo = playerInfo

	// Extract club name directly from the Player object
	if clubName := playerInfo.Player.Club; clubName != "" {
		// Map club to our internal type and update state
		clubType := g.mapGSProClubToInternal(clubName)
		if clubType != nil {
			log.Printf("GSPro selected club: %s (mapped to %v)", clubName, clubType)
			g.stateManager.SetClub(clubType)
		} else {
			log.Printf("Unmapped GSPro club: %s", clubName)
		}

		// Store the human-readable club name for camera metadata
		friendlyName := mapGSProClubToFriendlyName(clubName)
		g.stateManager.SetClubName(&friendlyName)
	}

	// Extract handedness from the Player object
	if handed := playerInfo.Player.Handed; handed != "" {
		// Map handedness to our internal type
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

// sendData sends shot data to GSPro
func (g *Integration) sendData(shotData ShotData) error {
	if !g.connected || g.socket == nil {
		return fmt.Errorf("not connected to GSPro")
	}

	jsonData, err := json.Marshal(shotData)
	if err != nil {
		return fmt.Errorf("error marshaling shot data: %w", err)
	}

	message := string(jsonData) + "\n"
	_, err = g.socket.Write([]byte(message))
	if err != nil {
		g.Disconnect()
		return fmt.Errorf("error sending data to GSPro: %w", err)
	}

	return nil
}
