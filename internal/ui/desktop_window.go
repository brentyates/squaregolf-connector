package ui

import webview "github.com/webview/webview_go"

type DesktopWindow struct {
	webview webview.WebView
}

func NewDesktopWindow(url string) *DesktopWindow {
	w := webview.New(true)
	w.SetTitle("SquareGolf Connector")
	w.SetSize(1440, 980, webview.HintNone)
	w.Navigate(url)

	return &DesktopWindow{webview: w}
}

func (w *DesktopWindow) Run() {
	defer w.webview.Destroy()
	w.webview.Run()
}

func (w *DesktopWindow) Terminate() {
	w.webview.Terminate()
}
