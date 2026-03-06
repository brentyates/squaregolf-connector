package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/brentyates/squaregolf-connector/internal/config"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/core/camera"
	"github.com/brentyates/squaregolf-connector/internal/core/gspro"
	"github.com/brentyates/squaregolf-connector/internal/core/infinitetees"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	stateManager            *core.StateManager
	bluetoothManager        *core.BluetoothManager
	launchMonitor           *core.LaunchMonitor
	gsproIntegration        *gspro.Integration
	infiniteTeesIntegration *infinitetees.Integration
	cameraManager           *camera.Manager
	enableExternalCamera    bool
	upgrader                websocket.Upgrader
	clients                 map[*websocket.Conn]chan []byte
	clientsMu               sync.Mutex
	broadcast               chan []byte
	httpServer              *http.Server
	httpServerMu            sync.Mutex
	webRoot                 string
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
	LauncherVersion  *string              `json:"launcherVersion"`
	MMIVersion       *string              `json:"mmiVersion"`
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

type InfiniteTeesStatus struct {
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
	DeviceName              string `json:"deviceName"`
	SpinMode                string `json:"spinMode"`
	GSProIP                 string `json:"gsproIP"`
	GSProPort               int    `json:"gsproPort"`
	GSProAutoConnect        bool   `json:"gsproAutoConnect"`
	InfiniteTeesIP          string `json:"infiniteTeesIP"`
	InfiniteTeesPort        int    `json:"infiniteTeesPort"`
	InfiniteTeesAutoConnect bool   `json:"infiniteTeesAutoConnect"`
}

type FeatureFlags struct {
	ExternalCamera bool `json:"externalCamera"`
}

func NewServer(stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor, cameraManager *camera.Manager, gsproIP string, gsproPort int, itIP string, itPort int, enableExternalCamera bool) *Server {
	gsproIntegration := gspro.GetInstance(stateManager, launchMonitor, gsproIP, gsproPort)
	itIntegration := infinitetees.GetInstance(stateManager, launchMonitor, itIP, itPort)

	server := &Server{
		stateManager:            stateManager,
		bluetoothManager:        bluetoothManager,
		launchMonitor:           launchMonitor,
		gsproIntegration:        gsproIntegration,
		infiniteTeesIntegration: itIntegration,
		cameraManager:           cameraManager,
		enableExternalCamera:    enableExternalCamera,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:   make(map[*websocket.Conn]chan []byte),
		broadcast: make(chan []byte, 100),
		webRoot:   resolveWebRoot(),
	}

	server.setupCallbacks()
	go server.handleMessages()

	return server
}

func resolveWebRoot() string {
	candidates := []string{}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "..", "Resources", "web"),
			filepath.Join(exeDir, "web"),
		)
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "web"))
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
		candidates = append(candidates, filepath.Join(repoRoot, "web"))
	}

	for _, candidate := range candidates {
		indexPath := filepath.Join(candidate, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return candidate
		}
	}

	return "web"
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

	s.stateManager.RegisterInfiniteTeesStatusCallback(func(oldValue, newValue core.InfiniteTeesConnectionStatus) {
		s.broadcastInfiniteTeesStatus()
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

	s.stateManager.RegisterLauncherVersionCallback(func(oldValue, newValue *string) {
		s.broadcastDeviceStatus()
	})

	s.stateManager.RegisterMMIVersionCallback(func(oldValue, newValue *string) {
		s.broadcastDeviceStatus()
	})
}

func (s *Server) handleMessages() {
	for message := range s.broadcast {
		s.clientsMu.Lock()
		clientCount := len(s.clients)
		activeChannels := make([]chan []byte, 0, clientCount)
		for _, clientChan := range s.clients {
			activeChannels = append(activeChannels, clientChan)
		}
		s.clientsMu.Unlock()

		for _, clientChan := range activeChannels {
			func(ch chan []byte) {
				defer func() {
					if r := recover(); r != nil {
						// Channel was closed by client disconnect, ignore panic
					}
				}()
				select {
				case ch <- message:
				default:
					log.Printf("WebSocket client buffer full, dropping message")
				}
			}(clientChan)
		}
	}
}

func (s *Server) broadcastDeviceStatus() {
	status := s.getDeviceStatus()
	log.Printf("Broadcasting device status - BallDetected: %v, BallPosition: %+v", status.BallDetected, status.BallPosition)
	msg := WSMessage{Type: "deviceStatus", Data: status}
	data, _ := json.Marshal(msg)
	select {
	case s.broadcast <- data:
	default:
	}
}

func (s *Server) broadcastGSProStatus() {
	status := s.getGSProStatus()
	msg := WSMessage{Type: "gsproStatus", Data: status}
	data, _ := json.Marshal(msg)
	select {
	case s.broadcast <- data:
	default:
	}
}

func (s *Server) broadcastInfiniteTeesStatus() {
	status := s.getInfiniteTeesStatus()
	msg := WSMessage{Type: "infiniteTeesStatus", Data: status}
	data, _ := json.Marshal(msg)
	select {
	case s.broadcast <- data:
	default:
	}
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
	case core.ConnectionStatusScanning:
		connectionStatus = "scanning"
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
		LauncherVersion:  s.stateManager.GetLauncherVersion(),
		MMIVersion:       s.stateManager.GetMMIVersion(),
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

func (s *Server) getInfiniteTeesStatus() InfiniteTeesStatus {
	var lastErrorStr string
	if err := s.stateManager.GetInfiniteTeesError(); err != nil {
		lastErrorStr = err.Error()
	}

	connectionStatus := "disconnected"
	switch s.stateManager.GetInfiniteTeesStatus() {
	case core.InfiniteTeesStatusConnected:
		connectionStatus = "connected"
	case core.InfiniteTeesStatusConnecting:
		connectionStatus = "connecting"
	case core.InfiniteTeesStatusError:
		connectionStatus = "error"
	}

	ip, port := s.infiniteTeesIntegration.GetConnectionInfo()
	settings := config.GetInstance().GetSettings()

	return InfiniteTeesStatus{
		ConnectionStatus: connectionStatus,
		IP:               ip,
		Port:             port,
		AutoConnect:      settings.InfiniteTeesAutoConnect,
		LastError:        lastErrorStr,
	}
}

func (s *Server) Start(port int) error {
	router := mux.NewRouter()

	// Serve static files with no-cache headers for development
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(s.webRoot, "static"))))
	router.PathPrefix("/static/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		staticHandler.ServeHTTP(w, r)
	}))

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Device endpoints
	api.HandleFunc("/device/status", s.handleDeviceStatus).Methods("GET")
	api.HandleFunc("/device/connect", s.handleDeviceConnect).Methods("POST")
	api.HandleFunc("/device/disconnect", s.handleDeviceDisconnect).Methods("POST")
	api.HandleFunc("/device/practice", s.handlePracticeMode).Methods("POST")

	// GSPro endpoints
	api.HandleFunc("/gspro/status", s.handleGSProStatus).Methods("GET")
	api.HandleFunc("/gspro/connect", s.handleGSProConnect).Methods("POST")
	api.HandleFunc("/gspro/disconnect", s.handleGSProDisconnect).Methods("POST")
	api.HandleFunc("/gspro/config", s.handleGSProConfig).Methods("GET", "POST")

	// Infinite Tees endpoints
	api.HandleFunc("/infinitetees/status", s.handleInfiniteTeesStatus).Methods("GET")
	api.HandleFunc("/infinitetees/connect", s.handleInfiniteTeesConnect).Methods("POST")
	api.HandleFunc("/infinitetees/disconnect", s.handleInfiniteTeesDisconnect).Methods("POST")
	api.HandleFunc("/infinitetees/config", s.handleInfiniteTeesConfig).Methods("GET", "POST")

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

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	s.httpServerMu.Lock()
	s.httpServer = httpServer
	s.httpServerMu.Unlock()

	log.Printf("Web server starting on port %d", port)
	log.Printf("Access via: http://localhost:%d", port)
	err := httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Stop(ctx context.Context) error {
	s.httpServerMu.Lock()
	httpServer := s.httpServer
	s.httpServerMu.Unlock()

	if httpServer == nil {
		return nil
	}

	return httpServer.Shutdown(ctx)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(s.webRoot, "index.html")
	http.ServeFile(w, r, indexPath)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	clientChan := make(chan []byte, 100)

	s.clientsMu.Lock()
	s.clients[conn] = clientChan
	s.clientsMu.Unlock()

	go func() {
		defer conn.Close()
		for msg := range clientChan {
			conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("WebSocket send error: %v, closing client", err)
				return
			}
		}
	}()

	s.sendInitialStatus(clientChan)

	defer func() {
		s.clientsMu.Lock()
		if ch, exists := s.clients[conn]; exists {
			delete(s.clients, conn)
			close(ch)
		}
		s.clientsMu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) sendInitialStatus(clientChan chan []byte) {
	// Send device status
	deviceStatus := s.getDeviceStatus()
	msg := WSMessage{Type: "deviceStatus", Data: deviceStatus}
	data, _ := json.Marshal(msg)
	clientChan <- data

	// Send GSPro status
	gsproStatus := s.getGSProStatus()
	msg = WSMessage{Type: "gsproStatus", Data: gsproStatus}
	data, _ = json.Marshal(msg)
	clientChan <- data

	// Send Infinite Tees status
	itStatus := s.getInfiniteTeesStatus()
	msg = WSMessage{Type: "infiniteTeesStatus", Data: itStatus}
	data, _ = json.Marshal(msg)
	clientChan <- data

	// Send camera config
	cameraConfig := s.getCameraConfig()
	msg = WSMessage{Type: "cameraConfig", Data: cameraConfig}
	data, _ = json.Marshal(msg)
	clientChan <- data
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
		settings := config.GetInstance().GetSettings()
		configData := struct {
			IP          string `json:"ip"`
			Port        int    `json:"port"`
			AutoConnect bool   `json:"autoConnect"`
		}{
			IP:          settings.GSProIP,
			Port:        settings.GSProPort,
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

		cfg := config.GetInstance()
		cfg.SetGSProIP(configData.IP)
		cfg.SetGSProPort(configData.Port)
		cfg.SetGSProAutoConnect(configData.AutoConnect)

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleInfiniteTeesStatus(w http.ResponseWriter, r *http.Request) {
	status := s.getInfiniteTeesStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleInfiniteTeesConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	go func() {
		s.infiniteTeesIntegration.ResetReconnectionState()
		s.infiniteTeesIntegration.EnableAutoReconnect()
		s.infiniteTeesIntegration.Start()
		s.infiniteTeesIntegration.Connect(req.IP, req.Port)
	}()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleInfiniteTeesDisconnect(w http.ResponseWriter, r *http.Request) {
	go func() {
		s.infiniteTeesIntegration.DisableAutoReconnect()
		s.infiniteTeesIntegration.Disconnect()
	}()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleInfiniteTeesConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		settings := config.GetInstance().GetSettings()
		configData := struct {
			IP          string `json:"ip"`
			Port        int    `json:"port"`
			AutoConnect bool   `json:"autoConnect"`
		}{
			IP:          settings.InfiniteTeesIP,
			Port:        settings.InfiniteTeesPort,
			AutoConnect: settings.InfiniteTeesAutoConnect,
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

		cfg := config.GetInstance()
		cfg.SetInfiniteTeesIP(configData.IP)
		cfg.SetInfiniteTeesPort(configData.Port)
		cfg.SetInfiniteTeesAutoConnect(configData.AutoConnect)

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		settings := config.GetInstance().GetSettings()

		appSettings := AppSettings{
			DeviceName:              settings.DeviceName,
			SpinMode:                settings.SpinMode,
			GSProIP:                 settings.GSProIP,
			GSProPort:               settings.GSProPort,
			GSProAutoConnect:        settings.GSProAutoConnect,
			InfiniteTeesIP:          settings.InfiniteTeesIP,
			InfiniteTeesPort:        settings.InfiniteTeesPort,
			InfiniteTeesAutoConnect: settings.InfiniteTeesAutoConnect,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(appSettings)
	} else {
		var rawSettings map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&rawSettings); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		cfg := config.GetInstance()

		if rawValue, ok := rawSettings["deviceName"]; ok {
			var value string
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid deviceName", http.StatusBadRequest)
				return
			}
			cfg.SetDeviceName(value)
		}

		if rawValue, ok := rawSettings["spinMode"]; ok {
			var value string
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid spinMode", http.StatusBadRequest)
				return
			}
			cfg.SetSpinMode(value)

			var spinMode core.SpinMode
			if value == "standard" {
				spinMode = core.Standard
			} else {
				spinMode = core.Advanced
			}
			s.stateManager.SetSpinMode(&spinMode)
		}

		if rawValue, ok := rawSettings["gsproIP"]; ok {
			var value string
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid gsproIP", http.StatusBadRequest)
				return
			}
			cfg.SetGSProIP(value)
		}

		if rawValue, ok := rawSettings["gsproPort"]; ok {
			var value int
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid gsproPort", http.StatusBadRequest)
				return
			}
			cfg.SetGSProPort(value)
		}

		if rawValue, ok := rawSettings["gsproAutoConnect"]; ok {
			var value bool
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid gsproAutoConnect", http.StatusBadRequest)
				return
			}
			cfg.SetGSProAutoConnect(value)
		}

		if rawValue, ok := rawSettings["infiniteTeesIP"]; ok {
			var value string
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid infiniteTeesIP", http.StatusBadRequest)
				return
			}
			cfg.SetInfiniteTeesIP(value)
		}

		if rawValue, ok := rawSettings["infiniteTeesPort"]; ok {
			var value int
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid infiniteTeesPort", http.StatusBadRequest)
				return
			}
			cfg.SetInfiniteTeesPort(value)
		}

		if rawValue, ok := rawSettings["infiniteTeesAutoConnect"]; ok {
			var value bool
			if err := json.Unmarshal(rawValue, &value); err != nil {
				http.Error(w, "Invalid infiniteTeesAutoConnect", http.StatusBadRequest)
				return
			}
			cfg.SetInfiniteTeesAutoConnect(value)
		}

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
	select {
	case s.broadcast <- data:
	default:
	}
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

func (s *Server) GetInfiniteTeesIntegration() *infinitetees.Integration {
	return s.infiniteTeesIntegration
}

func (s *Server) handlePracticeMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	if req.Enabled {
		err = s.launchMonitor.ActivateBallDetection()
	} else {
		err = s.launchMonitor.DeactivateBallDetection()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
