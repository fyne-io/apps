package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Feature struct {
	ID, Description, Image string

	Background, Color, Icon string
}

func makeFeatured(apps AppList, choose func(string)) *fyne.Container {
	res, err := http.Get("https://apps.fyne.io/api/v1/featured.json")
	if err != nil {
		fyne.LogError("Failed to parse featured", err)
		return &fyne.Container{}
	}
	defer res.Body.Close()

	list, err := parseFeatured(res.Body)
	if err != nil {
		fyne.LogError("Parse error", err)
		return &fyne.Container{}
	}

	if len(list) == 0 {
		return &fyne.Container{}
	}

	// Build all slide widgets up front
	slides := make([]fyne.CanvasObject, len(list))
	for i, item := range list {
		app := apps[item.ID]
		slides[i] = makeFeaturedSlide(app, item, choose)
	}

	// Show first slide
	current := 0
	slideContainer := container.NewStack(slides[0])

	// Dot indicators
	dots := make([]*canvas.Rectangle, len(list))
	dotObjects := make([]fyne.CanvasObject, len(list))
	for i := range list {
		dot := canvas.NewRectangle(color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		dot.SetMinSize(fyne.NewSquareSize(8))
		dot.CornerRadius = 4
		dots[i] = dot
		dotObjects[i] = dot
	}
	dots[0].FillColor = color.NRGBA{R: 100, G: 149, B: 237, A: 255}

	showSlide := func(index int) {
		if index < 0 || index >= len(slides) {
			return
		}
		dots[current].FillColor = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
		dots[current].Refresh()
		current = index
		dots[current].FillColor = color.NRGBA{R: 100, G: 149, B: 237, A: 255}
		dots[current].Refresh()

		slideContainer.RemoveAll()
		slideContainer.Add(slides[current])
		slideContainer.Refresh()
	}

	// Auto-advance timer, resets on manual interaction
	autoTimer := time.NewTicker(5 * time.Second)
	go func() {
		for range autoTimer.C {
			fyne.Do(func() {
				n := current + 1
				if n >= len(slides) {
					n = 0
				}
				showSlide(n)
			})
		}
	}()

	resetTimer := func() {
		autoTimer.Reset(5 * time.Second)
	}

	// Wrap dots in tappable buttons
	dotContainer := container.NewHBox(layout.NewSpacer())
	for i, dotObj := range dotObjects {
		idx := i
		tap := widget.NewButton("", func() {
			resetTimer()
			showSlide(idx)
		})
		tap.Importance = widget.LowImportance
		dotContainer.Add(container.NewStack(tap, container.NewCenter(dotObj)))
	}
	dotContainer.Add(layout.NewSpacer())

	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		resetTimer()
		next := current - 1
		if next < 0 {
			next = len(slides) - 1
		}
		showSlide(next)
	})
	prev.Importance = widget.LowImportance

	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		resetTimer()
		n := current + 1
		if n >= len(slides) {
			n = 0
		}
		showSlide(n)
	})
	next.Importance = widget.LowImportance

	carousel := container.NewBorder(nil, nil, prev, next, slideContainer)

	return container.NewVBox(carousel, dotContainer)
}

func makeFeaturedSlide(app App, item Feature, choose func(string)) fyne.CanvasObject {
	var obj fyne.CanvasObject
	if item.Image != "" {
		path := item.Image
		if path[0] == '/' {
			path = "https://apps.fyne.io" + path
		}

		u, _ := storage.ParseURI(path)
		img := canvas.NewImageFromURI(u)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(0, 150))

		obj = img
	} else {
		bg, fg := &color.NRGBA{}, &color.NRGBA{}
		bg.A = 0xff
		_, _ = fmt.Sscanf(item.Background, "#%02x%02x%02x", &bg.R, &bg.G, &bg.B)
		fg.A = 0xff
		_, _ = fmt.Sscanf(item.Color, "#%02x%02x%02x", &fg.R, &fg.G, &fg.B)

		path := item.Icon
		if path[0] == '/' {
			path = "https://apps.fyne.io" + path
		}

		u, _ := storage.ParseURI(path)
		icon := canvas.NewImageFromURI(u)
		icon.SetMinSize(fyne.NewSquareSize(56))

		description := canvas.NewText(item.Description, fg)
		description.TextSize = 13
		name := canvas.NewText(app.Name, fg)
		name.TextSize = 18
		name.TextStyle.Bold = true

		bgRect := canvas.NewRectangle(bg)
		bgRect.SetMinSize(fyne.NewSize(0, 150))
		bgRect.CornerRadius = 8

		content := container.NewVBox(
			container.NewCenter(container.NewHBox(icon, name)),
			container.NewCenter(description),
		)
		obj = container.NewStack(bgRect, container.NewCenter(content))
	}

	tap := widget.NewButton("", func() {
		choose(app.ID)
	})
	tap.Importance = widget.LowImportance

	return container.NewStack(tap, obj)
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
