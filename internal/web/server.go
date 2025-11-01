package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/brentyates/squaregolf-connector/internal/config"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/core/camera"
	"github.com/brentyates/squaregolf-connector/internal/core/gspro"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	stateManager         *core.StateManager
	bluetoothManager     *core.BluetoothManager
	launchMonitor        *core.LaunchMonitor
	gsproIntegration     *gspro.Integration
	cameraManager        *camera.Manager
	enableExternalCamera bool
	upgrader             websocket.Upgrader
	clients              map[*websocket.Conn]bool
	broadcast            chan []byte
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type DeviceStatus struct {
	ConnectionStatus string               `json:"connectionStatus"`
	DeviceName       *string              `json:"deviceName"`
	BatteryLevel     *int                 `json:"batteryLevel"`
	FirmwareVersion  *string              `json:"firmwareVersion"`
	BallDetected     bool                 `json:"ballDetected"`
	BallReady        bool                 `json:"ballReady"`
	BallPosition     *core.BallPosition   `json:"ballPosition"`
	Club             *core.ClubType       `json:"club"`
	Handedness       *core.HandednessType `json:"handedness"`
	LastError        string               `json:"lastError"`
	LastBallMetrics  *core.BallMetrics    `json:"lastBallMetrics"`
	LastClubMetrics  *core.ClubMetrics    `json:"lastClubMetrics"`
	IsAligning       bool                 `json:"isAligning"`
	AlignmentAngle   float64              `json:"alignmentAngle"`
	IsAligned        bool                 `json:"isAligned"`
}

type GSProStatus struct {
	ConnectionStatus string `json:"connectionStatus"`
	IP               string `json:"ip"`
	Port             int    `json:"port"`
	AutoConnect      bool   `json:"autoConnect"`
	LastError        string `json:"lastError"`
}

type CameraConfig struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type AppSettings struct {
	DeviceName       string `json:"deviceName"`
	AutoConnect      bool   `json:"autoConnect"`
	SpinMode         string `json:"spinMode"`
	GSProIP          string `json:"gsproIP"`
	GSProPort        int    `json:"gsproPort"`
	GSProAutoConnect bool   `json:"gsproAutoConnect"`
}

type FeatureFlags struct {
	ExternalCamera bool `json:"externalCamera"`
}

func NewServer(stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor, cameraManager *camera.Manager, gsproIP string, gsproPort int, enableExternalCamera bool) *Server {
	// Get GSPro integration
	gsproIntegration := gspro.GetInstance(stateManager, launchMonitor, gsproIP, gsproPort)

	server := &Server{
		stateManager:         stateManager,
		bluetoothManager:     bluetoothManager,
		launchMonitor:        launchMonitor,
		gsproIntegration:     gsproIntegration,
		cameraManager:        cameraManager,
		enableExternalCamera: enableExternalCamera,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}

	server.setupCallbacks()
	go server.handleMessages()

	return server
}

func (s *Server) setupCallbacks() {
	// Register all state callbacks to broadcast updates via WebSocket
	s.stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue core.ConnectionStatus) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterDeviceDisplayNameCallback(func(oldValue, newValue *string) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterBatteryLevelCallback(func(oldValue, newValue *int) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterBallDetectedCallback(func(oldValue, newValue bool) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterBallReadyCallback(func(oldValue, newValue bool) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterBallPositionCallback(func(oldValue, newValue *core.BallPosition) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterClubCallback(func(oldValue, newValue *core.ClubType) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterHandednessCallback(func(oldValue, newValue *core.HandednessType) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterLastErrorCallback(func(oldValue, newValue error) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterLastBallMetricsCallback(func(oldValue, newValue *core.BallMetrics) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterLastClubMetricsCallback(func(oldValue, newValue *core.ClubMetrics) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterGSProStatusCallback(func(oldValue, newValue core.GSProConnectionStatus) {
		s.broadcastGSProStatus()
	})

	s.stateManager.RegisterCameraURLCallback(func(oldValue, newValue *string) {
		s.broadcastCameraConfig()
	})

	s.stateManager.RegisterCameraEnabledCallback(func(oldValue, newValue bool) {
		s.broadcastCameraConfig()
	})

	s.stateManager.RegisterIsAligningCallback(func(oldValue, newValue bool) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterAlignmentAngleCallback(func(oldValue, newValue float64) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterIsAlignedCallback(func(oldValue, newValue bool) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterFirmwareVersionCallback(func(oldValue, newValue *string) {
		s.broadcastDeviceStatus()
	})
}

func (s *Server) handleMessages() {
	for {
		message := <-s.broadcast
		for client := range s.clients {
			select {
			case <-time.After(time.Second):
				delete(s.clients, client)
			default:
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(s.clients, client)
					client.Close()
				}
			}
		}
	}
}

func (s *Server) broadcastDeviceStatus() {
	status := s.getDeviceStatus()
	msg := WSMessage{Type: "deviceStatus", Data: status}
	data, _ := json.Marshal(msg)
	s.broadcast <- data
}

func (s *Server) broadcastGSProStatus() {
	status := s.getGSProStatus()
	msg := WSMessage{Type: "gsproStatus", Data: status}
	data, _ := json.Marshal(msg)
	s.broadcast <- data
}

func (s *Server) getDeviceStatus() DeviceStatus {
	var lastErrorStr string
	if err := s.stateManager.GetLastError(); err != nil {
		lastErrorStr = err.Error()
	}

	connectionStatus := "disconnected"
	switch s.stateManager.GetConnectionStatus() {
	case core.ConnectionStatusConnected:
		connectionStatus = "connected"
	case core.ConnectionStatusConnecting:
		connectionStatus = "connecting"
	case core.ConnectionStatusError:
		connectionStatus = "error"
	}

	return DeviceStatus{
		ConnectionStatus: connectionStatus,
		DeviceName:       s.stateManager.GetDeviceDisplayName(),
		BatteryLevel:     s.stateManager.GetBatteryLevel(),
		FirmwareVersion:  s.stateManager.GetFirmwareVersion(),
		BallDetected:     s.stateManager.GetBallDetected(),
		BallReady:        s.stateManager.GetBallReady(),
		BallPosition:     s.stateManager.GetBallPosition(),
		Club:             s.stateManager.GetClub(),
		Handedness:       s.stateManager.GetHandedness(),
		LastError:        lastErrorStr,
		LastBallMetrics:  s.stateManager.GetLastBallMetrics(),
		LastClubMetrics:  s.stateManager.GetLastClubMetrics(),
		IsAligning:       s.stateManager.GetIsAligning(),
		AlignmentAngle:   s.stateManager.GetAlignmentAngle(),
		IsAligned:        s.stateManager.GetIsAligned(),
	}
}

func (s *Server) getGSProStatus() GSProStatus {
	var lastErrorStr string
	if err := s.stateManager.GetGSProError(); err != nil {
		lastErrorStr = err.Error()
	}

	connectionStatus := "disconnected"
	switch s.stateManager.GetGSProStatus() {
	case core.GSProStatusConnected:
		connectionStatus = "connected"
	case core.GSProStatusConnecting:
		connectionStatus = "connecting"
	case core.GSProStatusError:
		connectionStatus = "error"
	}

	// Get current GSPro settings from integration and config
	ip, port := s.gsproIntegration.GetConnectionInfo()
	settings := config.GetInstance().GetSettings()

	return GSProStatus{
		ConnectionStatus: connectionStatus,
		IP:               ip,
		Port:             port,
		AutoConnect:      settings.GSProAutoConnect,
		LastError:        lastErrorStr,
	}
}

func (s *Server) Start(port int) error {
	router := mux.NewRouter()

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Device endpoints
	api.HandleFunc("/device/status", s.handleDeviceStatus).Methods("GET")
	api.HandleFunc("/device/connect", s.handleDeviceConnect).Methods("POST")
	api.HandleFunc("/device/disconnect", s.handleDeviceDisconnect).Methods("POST")

	// GSPro endpoints
	api.HandleFunc("/gspro/status", s.handleGSProStatus).Methods("GET")
	api.HandleFunc("/gspro/connect", s.handleGSProConnect).Methods("POST")
	api.HandleFunc("/gspro/disconnect", s.handleGSProDisconnect).Methods("POST")
	api.HandleFunc("/gspro/config", s.handleGSProConfig).Methods("GET", "POST")

	// Camera endpoints
	api.HandleFunc("/camera/config", s.handleCameraConfig).Methods("GET", "POST")

	// Settings endpoints
	api.HandleFunc("/settings", s.handleSettings).Methods("GET", "POST")

	// Feature flags endpoint
	api.HandleFunc("/features", s.handleFeatures).Methods("GET")

	// Alignment endpoints
	api.HandleFunc("/alignment/start", s.handleAlignmentStart).Methods("POST")
	api.HandleFunc("/alignment/stop", s.handleAlignmentStop).Methods("POST")
	api.HandleFunc("/alignment/cancel", s.handleAlignmentCancel).Methods("POST")
	api.HandleFunc("/alignment/handedness", s.handleAlignmentHandedness).Methods("POST")

	// WebSocket endpoint
	router.HandleFunc("/ws", s.handleWebSocket)

	// Serve index.html for all non-API routes (SPA support)
	router.PathPrefix("/").HandlerFunc(s.handleIndex)

	log.Printf("Web server starting on port %d", port)
	log.Printf("Access via: http://localhost:%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join("web", "index.html")
	http.ServeFile(w, r, indexPath)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true

	// Send initial status
	s.sendInitialStatus(conn)

	// Keep connection alive and handle client messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(s.clients, conn)
			break
		}
	}
}

func (s *Server) sendInitialStatus(conn *websocket.Conn) {
	// Send device status
	deviceStatus := s.getDeviceStatus()
	msg := WSMessage{Type: "deviceStatus", Data: deviceStatus}
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)

	// Send GSPro status
	gsproStatus := s.getGSProStatus()
	msg = WSMessage{Type: "gsproStatus", Data: gsproStatus}
	data, _ = json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)

	// Send camera config
	cameraConfig := s.getCameraConfig()
	msg = WSMessage{Type: "cameraConfig", Data: cameraConfig}
	data, _ = json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func (s *Server) handleDeviceStatus(w http.ResponseWriter, r *http.Request) {
	status := s.getDeviceStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleDeviceConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceName string `json:"deviceName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	go s.bluetoothManager.StartBluetoothConnection(req.DeviceName, "")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeviceDisconnect(w http.ResponseWriter, r *http.Request) {
	go s.bluetoothManager.DisconnectBluetooth()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGSProStatus(w http.ResponseWriter, r *http.Request) {
	status := s.getGSProStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleGSProConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	go func() {
		s.gsproIntegration.ResetReconnectionState()
		s.gsproIntegration.EnableAutoReconnect()
		s.gsproIntegration.Start()
		s.gsproIntegration.Connect(req.IP, req.Port)
	}()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGSProDisconnect(w http.ResponseWriter, r *http.Request) {
	go func() {
		s.gsproIntegration.DisableAutoReconnect()
		s.gsproIntegration.Disconnect()
	}()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGSProConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ip, port := s.gsproIntegration.GetConnectionInfo()
		settings := config.GetInstance().GetSettings()
		configData := struct {
			IP          string `json:"ip"`
			Port        int    `json:"port"`
			AutoConnect bool   `json:"autoConnect"`
		}{
			IP:          ip,
			Port:        port,
			AutoConnect: settings.GSProAutoConnect,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configData)
	} else {
		var configData struct {
			IP          string `json:"ip"`
			Port        int    `json:"port"`
			AutoConnect bool   `json:"autoConnect"`
		}
		if err := json.NewDecoder(r.Body).Decode(&configData); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Save GSPro settings
		cfg := config.GetInstance()
		cfg.SetGSProIP(configData.IP)
		cfg.SetGSProPort(configData.Port)
		cfg.SetGSProAutoConnect(configData.AutoConnect)

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		settings := config.GetInstance().GetSettings()

		appSettings := AppSettings{
			DeviceName:       settings.DeviceName,
			AutoConnect:      settings.AutoConnect,
			SpinMode:         settings.SpinMode,
			GSProIP:          settings.GSProIP,
			GSProPort:        settings.GSProPort,
			GSProAutoConnect: settings.GSProAutoConnect,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(appSettings)
	} else {
		var appSettings AppSettings
		if err := json.NewDecoder(r.Body).Decode(&appSettings); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Update settings in config manager
		cfg := config.GetInstance()
		cfg.SetDeviceName(appSettings.DeviceName)
		cfg.SetAutoConnect(appSettings.AutoConnect)
		cfg.SetSpinMode(appSettings.SpinMode)
		cfg.SetGSProIP(appSettings.GSProIP)
		cfg.SetGSProPort(appSettings.GSProPort)
		cfg.SetGSProAutoConnect(appSettings.GSProAutoConnect)

		// Update spin mode in state manager
		var spinMode core.SpinMode
		if appSettings.SpinMode == "standard" {
			spinMode = core.Standard
		} else {
			spinMode = core.Advanced
		}
		s.stateManager.SetSpinMode(&spinMode)

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) getCameraConfig() CameraConfig {
	url := "http://localhost:5000"
	if cameraURL := s.stateManager.GetCameraURL(); cameraURL != nil {
		url = *cameraURL
	}

	enabled := s.stateManager.GetCameraEnabled()

	return CameraConfig{
		URL:     url,
		Enabled: enabled,
	}
}

func (s *Server) handleCameraConfig(w http.ResponseWriter, r *http.Request) {
	// Return 404 if external camera feature is disabled
	if !s.enableExternalCamera {
		http.Error(w, "External camera feature not enabled", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		cameraConfig := s.getCameraConfig()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cameraConfig)
	} else {
		var cameraConfig CameraConfig
		if err := json.NewDecoder(r.Body).Decode(&cameraConfig); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Save camera settings to config
		cfg := config.GetInstance()
		cfg.SetCameraURL(cameraConfig.URL)
		cfg.SetCameraEnabled(cameraConfig.Enabled)

		// Update camera URL and enabled state in state manager
		s.stateManager.SetCameraURL(&cameraConfig.URL)
		s.stateManager.SetCameraEnabled(cameraConfig.Enabled)

		// Update camera manager
		s.cameraManager.SetBaseURL(cameraConfig.URL)
		s.cameraManager.SetEnabled(cameraConfig.Enabled)

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) broadcastCameraConfig() {
	config := s.getCameraConfig()
	msg := WSMessage{Type: "cameraConfig", Data: config}
	data, _ := json.Marshal(msg)
	s.broadcast <- data
}

func (s *Server) handleFeatures(w http.ResponseWriter, r *http.Request) {
	features := FeatureFlags{
		ExternalCamera: s.enableExternalCamera,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(features)
}

func (s *Server) handleAlignmentStart(w http.ResponseWriter, r *http.Request) {
	err := s.launchMonitor.StartAlignment()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAlignmentStop(w http.ResponseWriter, r *http.Request) {
	err := s.launchMonitor.StopAlignment()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAlignmentCancel(w http.ResponseWriter, r *http.Request) {
	err := s.launchMonitor.CancelAlignment()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAlignmentHandedness(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Handedness string `json:"handedness"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert string to HandednessType
	var handedness core.HandednessType
	if req.Handedness == "left" {
		handedness = core.LeftHanded
	} else if req.Handedness == "right" {
		handedness = core.RightHanded
	} else {
		http.Error(w, "Invalid handedness value (must be 'left' or 'right')", http.StatusBadRequest)
		return
	}

	// Update state manager
	s.stateManager.SetHandedness(&handedness)

	w.WriteHeader(http.StatusOK)
}
