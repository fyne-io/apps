package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"log"
)

func main() {
	a := app.New()
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
	w.SetContent(loadWelcome(apps))

	w.ShowAndRun()
}
