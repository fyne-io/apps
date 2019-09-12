package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

type welcome struct {
	shownID             string
	name, excerpt, date *widget.Label
	link                *widget.Hyperlink
	img                 *canvas.Image
}

func (w *welcome) loadAppDetail(app App) {
	w.shownID = app.ID

	w.name.SetText(app.Name)
	w.date.SetText(app.Date.Format("02 Jan 2006"))
	w.excerpt.SetText(app.Summary)

	if app.Icon != "" {
		res, err := loadResourceFromURL(app.Icon)
		if err == nil {
			w.img.Resource = res
			canvas.Refresh(w.img)
		}
	}

	parsed, err := url.Parse(app.Website)
	if err != nil {
		w.link.SetText("")
		return
	}
	w.link.SetText(parsed.Host)
	w.link.SetURL(parsed)
}

func loadResourceFromURL(urlStr string) (fyne.Resource, error) {
	res, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	parsed, _ := url.Parse(urlStr)
	name := filepath.Base(parsed.Path)
	return fyne.NewStaticResource(name, bytes), nil
}

func loadWelcome(apps AppList) fyne.CanvasObject {
	w := &welcome{}
	w.name = widget.NewLabel("")
	w.link = widget.NewHyperlink("", nil)
	w.date = widget.NewLabel("")
	w.excerpt = widget.NewLabel("")
	w.img = &canvas.Image{}
	w.img.SetMinSize(fyne.NewSize(320, 240))
	w.img.FillMode = canvas.ImageFillContain

	details := widget.NewForm(
		&widget.FormItem{Text: "Name", Widget: w.name},
		&widget.FormItem{Text: "HomePage", Widget: w.link},
		&widget.FormItem{Text: "Date", Widget: w.date},
		&widget.FormItem{Text: "Excerpt", Widget: w.excerpt},
		&widget.FormItem{Text: "Image", Widget: w.img},
	)

	list := widget.NewVBox()
	for _, app := range apps {
		capture := app
		list.Append(widget.NewButton(app.Name, func() {
			w.loadAppDetail(capture)
		}))
	}

	buttons := widget.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("Install", func() {}),
	)

	if len(apps) > 0 {
		w.loadAppDetail(apps[0])
	}
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, buttons, list, nil), buttons, list, details)
}
