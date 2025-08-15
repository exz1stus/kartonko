package main

import (
	"fmt"
	"kartonko-tooltip/tooltip"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	hook "github.com/robotn/gohook"
)

func main() {
	tooltip.Show()
	a := app.New()
	w := a.NewWindow("Hello World")

	w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()

	hook.Register(hook.KeyDown, []string{"q", "ctrl", "shift"}, summon)
}

func summon(e hook.Event) {
	fmt.Println("Hotkey pressed!")
}
