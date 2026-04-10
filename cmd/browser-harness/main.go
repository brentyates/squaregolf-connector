package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	appcfg "github.com/brentyates/squaregolf-connector/internal/config"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/web"
)

func main() {
	mockMode := flag.String("mock", "simulate", "Mock mode to use for the browser harness")
	port := flag.Int("port", 8091, "Port for the browser harness web server")
	omni := flag.Bool("omni", false, "Simulate an Omni device instead of Home")
	flag.Parse()

	log.Printf("Browser harness starting with mock mode %q on port %d", *mockMode, *port)

	stateManager := core.GetInstance()
	settings := appcfg.GetInstance().GetSettings()
	appcfg.GetInstance().ApplyToStateManager(stateManager)

	var bleClient core.BluetoothClient
	switch core.MockMode(*mockMode) {
	case core.MockModeStub:
		bleClient = core.NewMockBluetoothClient()
	case core.MockModeSimulate:
		bleClient = core.NewSimulatorBluetoothClient(core.SimulatorConfig{
			BatteryDrainRate: 1,
			ResponseDelay:    100 * time.Millisecond,
			SimulateOmni:     *omni,
		})
	default:
		log.Fatalf("unsupported mock mode for browser harness: %q", *mockMode)
	}

	bluetoothManager := core.GetBluetoothInstance(stateManager)
	bluetoothManager.SetClient(bleClient)
	launchMonitor := core.GetLaunchMonitorInstance(stateManager, bluetoothManager)
	launchMonitor.SetupNotifications(bluetoothManager)

	server := web.NewServer(
		stateManager,
		bluetoothManager,
		launchMonitor,
		nil,
		settings.GSProIP,
		settings.GSProPort,
		settings.InfiniteTeesIP,
		settings.InfiniteTeesPort,
		false,
	)

	bluetoothManager.StartBluetoothConnection(settings.DeviceName, "")

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(*port)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Browser harness shutting down on signal: %s", sig)
	case err := <-errCh:
		log.Fatalf("browser harness server error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("browser harness stop warning: %v", err)
	}
	bluetoothManager.DisconnectBluetooth()
}
