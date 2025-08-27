package components

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

func ShowThemedDialog(title, message string, window fyne.Window) {
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

func ShowThemedError(err error, window fyne.Window) {
    ShowThemedDialog("Error", err.Error(), window)
}


