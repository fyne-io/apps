package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Feature struct {
	ID, Description, Image string

	Background, Color, Icon string
}

func makeFeatured(apps AppList, choose func(string)) *fyne.Container {
	featured := canvas.NewRectangle(theme.ErrorColor())
	featured.SetMinSize(fyne.NewSquareSize(64))
	res, err := http.Get("https://apps.fyne.io/api/v1/featured.json")
	if err != nil {
		// TODO handle this!
		fyne.LogError("Failed to parse featured", err)
		return &fyne.Container{}
	}
	defer res.Body.Close()

	list, err := parseFeatured(res.Body)
	if err != nil {
		fyne.LogError("Parse error", err)
		return &fyne.Container{}
	}

	items := make([]fyne.CanvasObject, len(list))
	for i, item := range list {
		app := apps[item.ID]

		var obj fyne.CanvasObject
		if item.Image != "" {
			path := item.Image
			if path[0] == '/' {
				path = "https://apps.fyne.io" + path
			}

			u, _ := storage.ParseURI(path)
			img := canvas.NewImageFromURI(u)
			img.FillMode = canvas.ImageFillContain

			obj = img
		} else {
			bg, fg := &color.NRGBA{}, &color.NRGBA{}
			bg.A = 0xff
			_, err = fmt.Sscanf(item.Background, "#%02x%02x%02x", &bg.R, &bg.G, &bg.B)
			fg.A = 0xff
			_, err = fmt.Sscanf(item.Color, "#%02x%02x%02x", &fg.R, &fg.G, &fg.B)

			path := item.Icon
			if path[0] == '/' {
				path = "https://apps.fyne.io" + path
			}

			u, _ := storage.ParseURI(path)
			icon := canvas.NewImageFromURI(u)
			icon.SetMinSize(fyne.NewSquareSize(42))

			desc := item.Description
			if len(desc) > 64 {
				desc = desc[:62] + "…"
			}
			description := canvas.NewText(desc, fg)
			name := canvas.NewText(app.Name+" ", fg)
			name.TextSize += 6
			name.TextStyle.Bold = true

			content := container.NewVBox(container.NewCenter(container.NewHBox(name, icon)), description)
			obj = container.NewStack(canvas.NewRectangle(bg), container.NewCenter(content))
		}

		tap := widget.NewButton("", func() {
			choose(app.ID)
		})
		tap.Importance = widget.LowImportance

		items[i] = container.NewStack(tap, obj)
	}

	return container.NewGridWithColumns(1, items...)
}

func parseFeatured(reader io.Reader) ([]Feature, error) {
	decode := json.NewDecoder(reader)

	var list []Feature
	err := decode.Decode(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}
