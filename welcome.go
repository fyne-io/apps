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
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type welcome struct {
	shownID             string
	name, excerpt, date *widget.Label
	developer, version  *widget.Label
	link                *widget.Hyperlink
	icon, screenshot    *canvas.Image
}

func (w *welcome) loadAppDetail(app App) {
	w.shownID = app.ID

	w.name.SetText(app.Name)
	w.developer.SetText(app.Developer)
	w.version.SetText(app.Version)
	w.date.SetText(app.Date.Format("02 Jan 2006"))
	w.excerpt.SetText(app.Summary)

	w.icon.Resource = nil
	go setImageFromURL(w.icon, app.Icon)

	w.screenshot.Resource = nil
	if len(app.Screenshots) > 0 {
		go setImageFromURL(w.screenshot, app.Screenshots[0].Image)
	}

	parsed, err := url.Parse(app.Website)
	if err != nil {
		w.link.SetText("")
		return
	}
	w.link.SetText(parsed.Host)
	w.link.SetURL(parsed)
}

func setImageFromURL(img *canvas.Image, location string) {
	if location == "" {
		return
	}

	res, err := loadResourceFromURL(location)
	if err != nil {
		img.Resource = theme.WarningIcon()
	} else {
		img.Resource = res
	}

	canvas.Refresh(img)
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

// iconHoverLayout specifies a layout that floats an icon image top right over other content
type iconHoverLayout struct {
	content, icon fyne.CanvasObject
}

func (i *iconHoverLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	i.content.Resize(size)

	i.icon.Resize(fyne.NewSize(64, 64))
	i.icon.Move(fyne.NewPos(size.Width - i.icon.Size().Width, 0))
}

func (i *iconHoverLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return i.content.MinSize()
}

func loadWelcome(apps AppList) fyne.CanvasObject {
	w := &welcome{}
	w.name = widget.NewLabel("")
	w.developer = widget.NewLabel("")
	w.link = widget.NewHyperlink("", nil)
	w.excerpt = widget.NewLabel("")
	w.version = widget.NewLabel("")
	w.date = widget.NewLabel("")
	w.icon = &canvas.Image{}
	w.icon.FillMode = canvas.ImageFillContain
	w.screenshot = &canvas.Image{}
	w.screenshot.SetMinSize(fyne.NewSize(320, 240))
	w.screenshot.FillMode = canvas.ImageFillContain

	content := widget.NewForm(
		&widget.FormItem{Text: "Name", Widget: w.name},
		&widget.FormItem{Text: "Developer", Widget: w.developer},
		&widget.FormItem{Text: "Website", Widget: w.link},
		&widget.FormItem{Text: "Excerpt", Widget: w.excerpt},
		&widget.FormItem{Text: "Screenshot", Widget: w.screenshot},
		&widget.FormItem{Text: "Version", Widget: w.version},
		&widget.FormItem{Text: "Date", Widget: w.date},
	)
	details := fyne.NewContainerWithLayout(&iconHoverLayout{content:content, icon:w.icon}, content, w.icon)

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
