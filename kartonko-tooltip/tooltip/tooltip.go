package tooltip

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func Show() {

	myApp := app.New()
	w := myApp.NewWindow("Tooltip")

	w.SetContent(widget.NewLabel("Images here"))

	// Fyne does not support SetPosition directly!
	// You may have to wrap native code to move the window.

	w.Show()
	myApp.Run()
}
