package menu

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
	"github.com/brentyates/squaregolf-connector/internal/version"
)

// CreateSystemMenu creates the main menu for the application
func CreateSystemMenu(window fyne.Window) {
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Help", func() {
			showHelpDialog(window)
		}),
		fyne.NewMenuItem("About", func() {
			showAboutDialog(window)
		}),
	)

	mainMenu := fyne.NewMainMenu(helpMenu)
	window.SetMainMenu(mainMenu)
}

// showHelpDialog shows a dialog for users to submit bug reports and open the log directory
func showHelpDialog(window fyne.Window) {
	// Create bug report entry
	bugReportEntry := widget.NewMultiLineEntry()
	bugReportEntry.SetPlaceHolder("Describe what happened...")

	// Create submit button
	submitButton := widget.NewButton("Submit Bug Report", func() {
		if bugReportEntry.Text == "" {
			showThemedDialog(
				"Error",
				"Please provide a description of what happened.",
				window,
			)
			return
		}
		submitBugReport(bugReportEntry.Text, window)
	})

	// Create log directory button
	logDirButton := widget.NewButton("Open Log Directory", func() {
		openLogDirectory(window)
	})

	content := container.NewVBox(
		widget.NewLabel("Support"),
		widget.NewSeparator(),
		logDirButton,
		widget.NewLabel("Submit a Bug Report:"),
		bugReportEntry,
		submitButton,
	)

	d := dialog.NewCustom("Help", "Close", content, window)
	d.Show()
}

// showAboutDialog shows application version information
func showAboutDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabelWithStyle(
			core.AppName,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),
		widget.NewLabel(version.GetVersion()),
	)

	d := dialog.NewCustom("About", "Close", content, window)
	d.Show()
}

// openLogDirectory opens the log directory in the system's file explorer
func openLogDirectory(window fyne.Window) {
	logDir := logging.GetLogDirectory()

	// Check if directory exists
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		showThemedDialog(
			"Error",
			fmt.Sprintf("Log directory does not exist: %s", logDir),
			window,
		)
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
		showThemedDialog(
			"Error",
			"Unsupported operating system",
			window,
		)
		return
	}

	if err := cmd.Start(); err != nil {
		showThemedDialog(
			"Error",
			fmt.Sprintf("Failed to open log directory: %v", err),
			window,
		)
	}
}

// submitBugReport creates and opens a bug report file
func submitBugReport(description string, window fyne.Window) {
	// Create a temporary file for the bug report
	logDir := logging.GetLogDirectory()
	bugReportPath := filepath.Join(logDir, "bug_report.txt")

	// Read the log file
	logContent, err := os.ReadFile(logging.LogFile)
	if err != nil {
		showThemedDialog(
			"Error",
			fmt.Sprintf("Failed to read log file: %v", err),
			window,
		)
		return
	}

	// Create the bug report content
	bugReport := fmt.Sprintf("Bug Report for %s\n==========\n\nDescription:\n%s\n\nLogs:\n%s",
		core.AppName, description, string(logContent))

	// Write the bug report
	if err := os.WriteFile(bugReportPath, []byte(bugReport), 0644); err != nil {
		showThemedDialog(
			"Error",
			fmt.Sprintf("Failed to write bug report: %v", err),
			window,
		)
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
		showThemedDialog(
			"Error",
			"Unsupported operating system",
			window,
		)
		return
	}

	if err := cmd.Start(); err != nil {
		showThemedDialog(
			"Error",
			fmt.Sprintf("Failed to open bug report: %v", err),
			window,
		)
		return
	}

	showThemedDialog(
		"Bug Report Created",
		"A bug report has been created and opened in your default text editor.\nPlease review it and send it to the developer.",
		window,
	)
}

// showThemedDialog shows a dialog with proper theming
func showThemedDialog(title, message string, window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabelWithStyle(
			message,
			fyne.TextAlignLeading,
			fyne.TextStyle{},
		),
	)
	d := dialog.NewCustom(title, "OK", content, window)
	d.Show()
}
