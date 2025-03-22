package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// GSProMessage represents the base message structure from GSPro
type GSProMessage struct {
	Message string `json:"Message"`
}

// GSProPlayerInfo represents player information from GSPro
type GSProPlayerInfo struct {
	Message string      `json:"Message"`
	Player  GSProPlayer `json:"Player"`
}

// GSProPlayer represents player details from GSPro
type GSProPlayer struct {
	Club   string `json:"Club"`
	Handed string `json:"Handed"`
}

// GSProShotData represents the shot data sent to GSPro
type GSProShotData struct {
	DeviceID        string           `json:"DeviceID"`
	Units           string           `json:"Units"`
	APIversion      string           `json:"APIversion"`
	ShotNumber      int              `json:"ShotNumber"`
	ShotDataOptions GSProShotOptions `json:"ShotDataOptions"`
	BallData        *GSProBallData   `json:"BallData,omitempty"`
	ClubData        *GSProClubData   `json:"ClubData,omitempty"`
}

// GSProShotOptions represents shot data options
type GSProShotOptions struct {
	ContainsBallData          bool `json:"ContainsBallData"`
	ContainsClubData          bool `json:"ContainsClubData"`
	LaunchMonitorIsReady      bool `json:"LaunchMonitorIsReady,omitempty"`
	LaunchMonitorBallDetected bool `json:"LaunchMonitorBallDetected,omitempty"`
}

// GSProBallData represents ball data sent to GSPro
type GSProBallData struct {
	Speed     float64 `json:"Speed"`
	SpinAxis  float64 `json:"SpinAxis"`
	TotalSpin int16   `json:"TotalSpin"`
	BackSpin  int16   `json:"BackSpin"`
	SideSpin  int16   `json:"SideSpin"`
	HLA       float64 `json:"HLA"`
	VLA       float64 `json:"VLA"`
}

// GSProClubData represents club data sent to GSPro
type GSProClubData struct {
	Speed                float64 `json:"Speed"`
	AngleOfAttack        float64 `json:"AngleOfAttack"`
	FaceToTarget         float64 `json:"FaceToTarget"`
	Lie                  float64 `json:"Lie"`
	Loft                 float64 `json:"Loft"`
	Path                 float64 `json:"Path"`
	SpeedAtImpact        float64 `json:"SpeedAtImpact"`
	VerticalFaceImpact   float64 `json:"VerticalFaceImpact"`
	HorizontalFaceImpact float64 `json:"HorizontalFaceImpact"`
	ClosureRate          float64 `json:"ClosureRate"`
}

// GSProIntegration handles communication with GSPro
type GSProIntegration struct {
	stateManager   *StateManager
	launchMonitor  *LaunchMonitor
	host           string
	port           int
	socket         net.Conn
	connected      bool
	running        bool
	connectMutex   sync.Mutex
	shotNumber     int
	shotListeners  []func(GSProShotData)
	lastPlayerInfo *GSProPlayerInfo
	wg             sync.WaitGroup
}

// NewGSProIntegration creates a new GSProIntegration instance
func NewGSProIntegration(stateManager *StateManager, launchMonitor *LaunchMonitor, host string, port int) *GSProIntegration {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 921
	}

	gspro := &GSProIntegration{
		stateManager:  stateManager,
		launchMonitor: launchMonitor,
		host:          host,
		port:          port,
		shotListeners: make([]func(GSProShotData), 0),
	}

	// Register state listeners
	gspro.registerStateListeners()

	return gspro
}

// Register state listeners
func (g *GSProIntegration) registerStateListeners() {
	g.stateManager.RegisterBallReadyCallback(g.onBallReadyChanged)
	g.stateManager.RegisterLastBallMetricsCallback(g.onLastBallMetricsChanged)
	g.stateManager.RegisterLastClubMetricsCallback(g.onLastClubMetricsChanged)
}

// onBallReadyChanged handles ball ready state changed event from state manager
func (g *GSProIntegration) onBallReadyChanged(oldValue, newValue bool) {
	if oldValue == newValue {
		return
	}

	if !g.connected || g.socket == nil {
		log.Println("GSPro not connected, skipping ball ready send")
		return
	}

	// Send empty shot data
	emptyShotData := GSProShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: g.shotNumber,
		ShotDataOptions: GSProShotOptions{
			ContainsBallData:          false,
			ContainsClubData:          false,
			LaunchMonitorIsReady:      newValue,
			LaunchMonitorBallDetected: newValue,
		},
	}

	try := func() error {
		return g.sendData(emptyShotData)
	}

	if err := try(); err != nil {
		log.Printf("Error sending empty shot data to GSPro: %v", err)
	}
}

// onLastBallMetricsChanged handles last ball metrics changed event from state manager
func (g *GSProIntegration) onLastBallMetricsChanged(oldValue, newValue *BallMetrics) {
	if oldValue == newValue {
		return
	}

	if !g.connected || g.socket == nil {
		log.Println("GSPro not connected, skipping ball data send")
		return
	}

	if newValue == nil {
		return
	}

	try := func() error {
		gsproShotData := g.convertToGSProShotFormat(*newValue)
		return g.sendData(gsproShotData)
	}

	if err := try(); err != nil {
		log.Printf("Error sending shot data to GSPro: %v", err)
	}
}

// onLastClubMetricsChanged handles last club metrics changed event from state manager
func (g *GSProIntegration) onLastClubMetricsChanged(oldValue, newValue *ClubMetrics) {
	if oldValue == newValue {
		return
	}

	if !g.connected || g.socket == nil {
		log.Println("GSPro not connected, skipping club data send")
		return
	}

	if newValue == nil {
		return
	}

	// Update the last shot data with club metrics
	try := func() error {
		gsproShotData := g.convertToGSProShotFormat(BallMetrics{})
		gsproShotData.ShotDataOptions.ContainsClubData = true
		gsproShotData.ClubData = g.convertClubDataToGSPro(*newValue)
		return g.sendData(gsproShotData)
	}

	if err := try(); err != nil {
		log.Printf("Error sending club data to GSPro: %v", err)
	}
}

// convertToGSProShotFormat converts internal shot data format to GSPro format
func (g *GSProIntegration) convertToGSProShotFormat(ballMetrics BallMetrics) GSProShotData {
	// Increment shot number for each shot
	g.shotNumber++

	return GSProShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: g.shotNumber,
		ShotDataOptions: GSProShotOptions{
			ContainsBallData: true,
			ContainsClubData: false,
		},
		BallData: &GSProBallData{
			Speed:     ballMetrics.BallSpeedMPS * 2.23694, // Convert m/s to mph
			SpinAxis:  ballMetrics.SpinAxis * -1,
			TotalSpin: ballMetrics.TotalspinRPM,
			BackSpin:  ballMetrics.BackspinRPM,
			SideSpin:  ballMetrics.SidespinRPM * -1,
			HLA:       ballMetrics.HorizontalAngle,
			VLA:       ballMetrics.VerticalAngle,
		},
		ClubData: &GSProClubData{}, // Empty club data
	}
}

// convertClubDataToGSPro converts internal club data format to GSPro format
func (g *GSProIntegration) convertClubDataToGSPro(clubMetrics ClubMetrics) *GSProClubData {
	return &GSProClubData{
		Speed:                0, // Not provided by our sensor
		AngleOfAttack:        clubMetrics.AttackAngle,
		FaceToTarget:         clubMetrics.FaceAngle,
		Lie:                  0, // Not provided by our sensor
		Loft:                 clubMetrics.DynamicLoftAngle,
		Path:                 clubMetrics.PathAngle,
		SpeedAtImpact:        0, // Not provided by our sensor
		VerticalFaceImpact:   0, // Not provided by our sensor
		HorizontalFaceImpact: 0, // Not provided by our sensor
		ClosureRate:          0, // Not provided by our sensor
	}
}

// sendData sends shot data to GSPro
func (g *GSProIntegration) sendData(shotData GSProShotData) error {
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

// mapGSProClubToInternal maps GSPro club name to internal ClubType
func (g *GSProIntegration) mapGSProClubToInternal(clubName string) *ClubType {
	// Map GSPro club names to our internal ClubType
	clubMap := map[string]ClubType{
		// Drivers and woods
		"DR": ClubDriver,
		"W2": ClubWood3,
		"W3": ClubWood3,
		"W4": ClubWood5,
		"W5": ClubWood5,
		"W6": ClubWood7,
		"W7": ClubWood7,

		// Hybrids
		"H2": ClubWood3,
		"H3": ClubWood3,
		"H4": ClubWood3,
		"H5": ClubWood3,
		"H6": ClubWood5,
		"H7": ClubIron4,

		// Irons
		"I1": ClubWood3,
		"I2": ClubWood3,
		"I3": ClubWood5,
		"I4": ClubIron4,
		"I5": ClubIron5,
		"I6": ClubIron6,
		"I7": ClubIron7,
		"I8": ClubIron8,
		"I9": ClubIron9,

		// Wedges
		"PW": ClubPitchingWedge,
		"AW": ClubApproachWedge,
		"GW": ClubApproachWedge,
		"SW": ClubSandWedge,

		// Putter
		"PT": ClubPutter,
	}

	if club, ok := clubMap[clubName]; ok {
		return &club
	}
	return nil
}

// Start starts the GSPro integration in a separate goroutine
func (g *GSProIntegration) Start() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if g.running {
		log.Println("GSPro integration already running")
		return
	}

	g.running = true
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.connectionThread()
	}()
}

// Stop stops the GSPro integration
func (g *GSProIntegration) Stop() {
	log.Println("Stopping GSPro integration...")
	g.running = false
	g.Disconnect()

	// Wait for all goroutines to complete with a timeout
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("GSPro integration stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for GSPro integration to stop")
		g.stateManager.SetGSProError(fmt.Errorf("timeout waiting for integration to stop"))
		g.stateManager.SetGSProStatus(GSProStatusError)
	}
}

// connectionThread manages connection to GSPro
func (g *GSProIntegration) connectionThread() {
	for g.running {
		if !g.connected {
			g.Connect(g.host, g.port)
		}

		// Check connection every 5 seconds
		time.Sleep(5 * time.Second)
	}
}

// Connect connects to GSPro server
func (g *GSProIntegration) Connect(host string, port int) {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if g.connected {
		return
	}

	// Update host and port
	g.host = host
	g.port = port

	// Set connecting state
	g.stateManager.SetGSProStatus(GSProStatusConnecting)

	addr := net.JoinHostPort(g.host, fmt.Sprintf("%d", g.port))
	log.Printf("Connecting to GSPro server at %s", addr)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Printf("Error connecting to GSPro server: %v", err)
		g.stateManager.SetGSProError(fmt.Errorf("failed to connect: %v", err))
		g.stateManager.SetGSProStatus(GSProStatusError)
		return
	}

	g.socket = conn
	g.connected = true
	log.Printf("Connected to GSPro server at %s", addr)

	// Start receiving messages in a goroutine
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.receiveMessages()
	}()

	// Add a small delay before setting connected state to ensure UI shows "Connecting..."
	time.Sleep(500 * time.Millisecond)
	g.stateManager.SetGSProStatus(GSProStatusConnected)
}

// receiveMessages receives and processes messages from GSPro
func (g *GSProIntegration) receiveMessages() {
	if g.socket == nil {
		return
	}

	buffer := make([]byte, 4096)

	for g.running && g.connected {
		g.socket.SetReadDeadline(time.Now().Add(10 * time.Second))

		n, err := g.socket.Read(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("Error reading from GSPro: %v", err)
			g.stateManager.SetGSProError(fmt.Errorf("error reading from GSPro: %v", err))
			g.stateManager.SetGSProStatus(GSProStatusError)
			break
		}

		if n == 0 {
			log.Println("GSPro server closed connection")
			g.stateManager.SetGSProError(fmt.Errorf("server closed connection"))
			g.stateManager.SetGSProStatus(GSProStatusError)
			break
		}

		message := string(buffer[:n])
		log.Printf("Received message from GSPro: %s", message)
		g.processMessage(message)
	}

	g.Disconnect()
}

// processMessage processes a message from GSPro
func (g *GSProIntegration) processMessage(rawMessage string) {
	var baseMsg GSProMessage
	if err := json.Unmarshal([]byte(rawMessage), &baseMsg); err != nil {
		log.Printf("Invalid JSON from GSPro: %v", err)
		return
	}

	switch baseMsg.Message {
	case "GSPro ready":
		g.handleGSProReadyMessage()
	case "GSPro Player Information":
		var playerInfo GSProPlayerInfo
		if err := json.Unmarshal([]byte(rawMessage), &playerInfo); err != nil {
			log.Printf("Error parsing player info: %v", err)
			return
		}
		g.handlePlayerMessage(&playerInfo)
		g.handleGSProReadyMessage()
	case "Ball Data received":
	default:
		log.Printf("Unknown GSPro message type: %s", baseMsg.Message)
	}
}

// handleGSProReadyMessage handles the GSPro ready message and activates ball detection if in manual mode
func (g *GSProIntegration) handleGSProReadyMessage() {
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
func (g *GSProIntegration) handlePlayerMessage(playerInfo *GSProPlayerInfo) {
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
	}

	// Extract handedness from the Player object
	if handed := playerInfo.Player.Handed; handed != "" {
		// Map handedness to our internal type
		var handednessType HandednessType
		if handed == "LH" {
			handednessType = LeftHanded
			log.Printf("GSPro selected handedness: Left-handed")
		} else {
			handednessType = RightHanded
			log.Printf("GSPro selected handedness: Right-handed")
		}
		g.stateManager.SetHandedness(&handednessType)
	}
}

// Disconnect disconnects from GSPro server
func (g *GSProIntegration) Disconnect() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if !g.connected || g.socket == nil {
		g.connected = false
		g.stateManager.SetGSProStatus(GSProStatusDisconnected)
		return
	}

	log.Println("Disconnecting from GSPro server...")

	// Set a deadline for graceful disconnect
	_ = g.socket.SetDeadline(time.Now().Add(2 * time.Second))

	// Close the socket with proper error handling
	if g.socket != nil {
		err := g.socket.Close()
		if err != nil {
			log.Printf("Error closing GSPro connection: %v", err)
			g.stateManager.SetGSProError(fmt.Errorf("error closing connection: %v", err))
			g.stateManager.SetGSProStatus(GSProStatusError)
		}
		g.socket = nil
	}

	g.connected = false
	g.stateManager.SetGSProStatus(GSProStatusDisconnected)
	log.Println("Disconnected from GSPro server")
}

// AddShotListener adds a listener for shot events
func (g *GSProIntegration) AddShotListener(listener func(GSProShotData)) {
	g.shotListeners = append(g.shotListeners, listener)
}
