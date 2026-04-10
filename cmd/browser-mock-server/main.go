package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type wsMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type app struct {
	mu       sync.Mutex
	status   map[string]interface{}
	gspro    map[string]interface{}
	it       map[string]interface{}
	camera   map[string]interface{}
	settings map[string]interface{}
	features map[string]interface{}
	clients  map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
	webRoot  string
}

func main() {
	port := flag.Int("port", 8091, "Port for the browser mock server")
	flag.Parse()

	a := &app{
		status: map[string]interface{}{
			"connectionStatus":    "disconnected",
			"deviceName":          nil,
			"batteryLevel":        nil,
			"firmwareVersion":     "1.9.27",
			"launcherVersion":     "1.0.0",
			"mmiVersion":          "1.2.0",
			"launchMonitorStatus": "none",
			"ballDetected":        false,
			"ballReady":           false,
			"ballPosition":        nil,
			"club":                map[string]interface{}{"regularCode": "0204", "swingStickCode": "0202"},
			"handedness":          0,
			"lastError":           "",
			"lastBallMetrics":     nil,
			"lastClubMetrics":     nil,
			"isAligning":          false,
			"alignmentAngle":      0,
			"isAligned":           false,
		},
		gspro: map[string]interface{}{
			"connectionStatus": "disconnected",
			"ip":               "127.0.0.1",
			"port":             921,
			"autoConnect":      false,
			"lastError":        "",
		},
		it: map[string]interface{}{
			"connectionStatus": "disconnected",
			"ip":               "127.0.0.1",
			"port":             999,
			"autoConnect":      false,
			"lastError":        "",
		},
		camera: map[string]interface{}{
			"url":     "http://localhost:5000",
			"enabled": false,
		},
		settings: map[string]interface{}{
			"deviceName":              "",
			"spinMode":                "advanced",
			"gsproIP":                 "127.0.0.1",
			"gsproPort":               921,
			"gsproAutoConnect":        false,
			"infiniteTeesIP":          "127.0.0.1",
			"infiniteTeesPort":        999,
			"infiniteTeesAutoConnect": false,
		},
		features: map[string]interface{}{
			"externalCamera": false,
		},
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		webRoot: "/Users/byates/projects/squaregolf-connector/web",
	}

	router := mux.NewRouter()
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(a.webRoot, "static"))))
	router.PathPrefix("/static/").Handler(staticHandler)
	router.HandleFunc("/ws", a.handleWS)
	router.HandleFunc("/api/features", a.handleFeatures).Methods("GET")
	router.HandleFunc("/api/settings", a.handleSettings).Methods("GET", "POST")
	router.HandleFunc("/api/device/connect", a.handleConnect).Methods("POST")
	router.HandleFunc("/api/device/disconnect", a.handleDisconnect).Methods("POST")
	router.HandleFunc("/api/device/practice", a.handlePractice).Methods("POST")
	router.HandleFunc("/api/camera/config", a.handleCamera).Methods("GET", "POST")
	router.HandleFunc("/", a.handleIndex)
	router.PathPrefix("/").HandlerFunc(a.handleIndex)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Browser mock server listening on http://localhost:%d", *port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(a.webRoot, "index.html"))
}

func (a *app) handleFeatures(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, a.features)
}

func (a *app) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var incoming map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&incoming)
		for k, v := range incoming {
			a.settings[k] = v
		}
	}
	a.writeJSON(w, a.settings)
}

func (a *app) handleCamera(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var incoming map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&incoming)
		for k, v := range incoming {
			a.camera[k] = v
		}
		a.broadcast("cameraConfig", a.camera)
	}
	a.writeJSON(w, a.camera)
}

func (a *app) handleConnect(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.status["connectionStatus"] = "connected"
	a.status["deviceName"] = "SquareGolf Simulator"
	a.status["batteryLevel"] = 80
	a.status["launchMonitorStatus"] = "idle"
	a.mu.Unlock()
	a.broadcast("deviceStatus", a.snapshotStatus())
	w.WriteHeader(http.StatusOK)
}

func (a *app) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.status["connectionStatus"] = "disconnected"
	a.status["launchMonitorStatus"] = "none"
	a.status["ballDetected"] = false
	a.status["ballReady"] = false
	a.status["ballPosition"] = nil
	a.status["lastBallMetrics"] = nil
	a.status["lastClubMetrics"] = nil
	a.mu.Unlock()
	a.broadcast("deviceStatus", a.snapshotStatus())
	w.WriteHeader(http.StatusOK)
}

func (a *app) handlePractice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Enabled {
		go a.runShotSequence()
	}
	w.WriteHeader(http.StatusOK)
}

func (a *app) runShotSequence() {
	stages := []struct {
		delay  time.Duration
		update func(map[string]interface{})
	}{
		{
			delay: 400 * time.Millisecond,
			update: func(status map[string]interface{}) {
				status["launchMonitorStatus"] = "detect"
				status["ballDetected"] = true
				status["ballPosition"] = map[string]interface{}{"x": 14, "y": -8, "z": 2}
			},
		},
		{
			delay: 700 * time.Millisecond,
			update: func(status map[string]interface{}) {
				status["launchMonitorStatus"] = "ready"
				status["ballReady"] = true
				status["ballPosition"] = map[string]interface{}{"x": 3, "y": -2, "z": 1}
			},
		},
		{
			delay: 1100 * time.Millisecond,
			update: func(status map[string]interface{}) {
				status["launchMonitorStatus"] = "shot"
				status["lastBallMetrics"] = map[string]interface{}{
					"rawData":          []string{"11", "02", "37", "84", "0B", "F0", "03", "14", "00", "96", "00", "C2", "01", "6E", "00", "00", "80"},
					"speed":            29.48,
					"launchAngle":      10.08,
					"horizontalAngle":  0.20,
					"totalSpin":        150,
					"spinAxis":         4.50,
					"backSpin":         110,
					"sideSpin":         0,
					"isBallSpeedValid": true,
					"isTotalSpinValid": true,
					"isSpinAxisValid":  true,
					"isBackSpinValid":  true,
					"isSideSpinValid":  false,
				}
			},
		},
		{
			delay: 1500 * time.Millisecond,
			update: func(status map[string]interface{}) {
				status["launchMonitorStatus"] = "done"
				status["lastClubMetrics"] = map[string]interface{}{
					"rawData":            []string{"11", "07", "0D", "D4", "FE", "90", "01", "00", "80", "D0", "07"},
					"path":               -3.00,
					"angle":              4.00,
					"attackAngle":        0,
					"dynamicLoft":        20.00,
					"isPathValid":        true,
					"isFaceAngleValid":   true,
					"isAttackAngleValid": false,
					"isDynamicLoftValid": true,
				}
			},
		},
	}

	for _, stage := range stages {
		time.Sleep(stage.delay)
		a.mu.Lock()
		stage.update(a.status)
		snapshot := a.snapshotStatusLocked()
		a.mu.Unlock()
		a.broadcast("deviceStatus", snapshot)
	}
}

func (a *app) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	a.mu.Lock()
	a.clients[conn] = struct{}{}
	device := a.snapshotStatusLocked()
	gspro := cloneMap(a.gspro)
	it := cloneMap(a.it)
	camera := cloneMap(a.camera)
	a.mu.Unlock()

	a.send(conn, "deviceStatus", device)
	a.send(conn, "gsproStatus", gspro)
	a.send(conn, "infiniteTeesStatus", it)
	a.send(conn, "cameraConfig", camera)

	go func() {
		defer func() {
			a.mu.Lock()
			delete(a.clients, conn)
			a.mu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}()
}

func (a *app) snapshotStatus() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshotStatusLocked()
}

func (a *app) snapshotStatusLocked() map[string]interface{} {
	return cloneMap(a.status)
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	raw, _ := json.Marshal(src)
	var out map[string]interface{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func (a *app) broadcast(msgType string, data interface{}) {
	a.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(a.clients))
	for conn := range a.clients {
		conns = append(conns, conn)
	}
	a.mu.Unlock()

	for _, conn := range conns {
		a.send(conn, msgType, data)
	}
}

func (a *app) send(conn *websocket.Conn, msgType string, data interface{}) {
	_ = conn.WriteJSON(wsMessage{Type: msgType, Data: data})
}

func (a *app) writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
