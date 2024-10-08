//go:generate fyne bundle -o bundled.go Icon.png

package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("io.fyne.apps")
	a.SetIcon(resourceIconPng)
	w := a.NewWindow("Fyne Applications")

	data, err := loadAppListFromWeb()
	if err != nil {
		log.Println("Web failed, reading cache")
		data, err = loadAppListFromCache()
		if err != nil {
			fyne.LogError("Load error", err)
			return
		}
	}
	defer data.Close()
	apps, err := parseAppList(data)
	if err != nil {
		fyne.LogError("Parse error", err)
		return
	}
	w.SetContent(loadUI(apps, w))
	w.Resize(fyne.NewSize(680, 520))

	w.ShowAndRun()
}
