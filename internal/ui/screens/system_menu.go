package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/logging"
	"github.com/brentyates/squaregolf-connector/internal/ui/components"
	"github.com/brentyates/squaregolf-connector/internal/version"
)

// SystemMenu represents the system menu that combines Help and About functionality
type SystemMenu struct {
	window fyne.Window
}

// NewSystemMenu creates a new system menu instance
func NewSystemMenu(w fyne.Window) *SystemMenu {
	return &SystemMenu{
		window: w,
	}
}

// OpenLogDirectory opens the log directory in the system's file explorer
func (sm *SystemMenu) OpenLogDirectory() {
	logDir := logging.GetLogDirectory()

	// Check if directory exists
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		components.ShowThemedError(fmt.Errorf("log directory does not exist: %s", logDir), sm.window)
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", logDir)
	case "windows":
		cmd = exec.Command("explorer", logDir)
	case "linux":
		cmd = exec.Command("xdg-open", logDir)
	default:
		components.ShowThemedError(fmt.Errorf("unsupported operating system"), sm.window)
		return
	}

	if err := cmd.Start(); err != nil {
		components.ShowThemedError(fmt.Errorf("failed to open log directory: %v", err), sm.window)
	}
}

// submitBugReport creates and opens a bug report file
func (sm *SystemMenu) submitBugReport(description string) {
	// Create a temporary file for the bug report
	logDir := logging.GetLogDirectory()
	bugReportPath := filepath.Join(logDir, "bug_report.txt")

	// Read the log file
	logContent, err := os.ReadFile(logging.LogFile)
	if err != nil {
		components.ShowThemedError(fmt.Errorf("failed to read log file: %v", err), sm.window)
		return
	}

	// Create the bug report content
	bugReport := fmt.Sprintf("Bug Report for %s\n==========\n\nDescription:\n%s\n\nLogs:\n%s",
		core.AppName, description, string(logContent))

	// Write the bug report
	if err := os.WriteFile(bugReportPath, []byte(bugReport), 0644); err != nil {
		components.ShowThemedError(fmt.Errorf("failed to write bug report: %v", err), sm.window)
		return
	}

	// Open the bug report in the default text editor
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-e", bugReportPath)
	case "windows":
		cmd = exec.Command("notepad", bugReportPath)
	case "linux":
		cmd = exec.Command("xdg-open", bugReportPath)
	default:
		components.ShowThemedError(fmt.Errorf("unsupported operating system"), sm.window)
		return
	}

	if err := cmd.Start(); err != nil {
		components.ShowThemedError(fmt.Errorf("failed to open bug report: %v", err), sm.window)
		return
	}

	components.ShowThemedDialog(
		"Bug Report Created",
		"A bug report has been created and opened in your default text editor.\nPlease review it and send it to the developer.",
		sm.window,
	)
}

// ShowBugReport shows the bug report dialog
func (sm *SystemMenu) ShowBugReport() {
	// Create bug report entry
	bugReportEntry := widget.NewMultiLineEntry()
	bugReportEntry.SetPlaceHolder("Describe what happened...")

	// Create submit button
	submitButton := widget.NewButton("Submit", func() {
		if bugReportEntry.Text == "" {
			components.ShowThemedDialog(
				"Error",
				"Please provide a description of what happened.",
				sm.window,
			)
			return
		}
		sm.submitBugReport(bugReportEntry.Text)
	})

	content := container.NewVBox(
		widget.NewLabel("Please describe the issue you encountered:"),
		bugReportEntry,
		submitButton,
	)

	d := dialog.NewCustom("Submit Bug Report", "Close", content, sm.window)
	d.Resize(fyne.NewSize(500, 300))
	d.Show()
}

// ShowAbout shows the about dialog
func (sm *SystemMenu) ShowAbout() {
	content := container.NewVBox(
		widget.NewLabel("About "+core.AppName),
		widget.NewSeparator(),
		widget.NewLabel("Version: "+version.GetVersion()),
		widget.NewLabel(""),
		widget.NewLabel("IMPORTANT: This is an UNOFFICIAL connector application"),
		widget.NewLabel("for the Square Golf launch monitor."),
		widget.NewLabel(""),
		widget.NewLabel("This application is not affiliated with, maintained, authorized,"),
		widget.NewLabel("or endorsed by Square Golf or any of its affiliates."),
		widget.NewLabel(""),
		widget.NewLabel("Created by the open source community"),
		widget.NewLabel("for the golf simulator community."),
	)

	d := dialog.NewCustom("About", "Close", content, sm.window)
	d.Show()
}
