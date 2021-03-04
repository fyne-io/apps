package main

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne/commands"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type welcome struct {
	shownID, shownPkg, shownIcon string
	name, summary, date          *widget.Label
	developer, version           *widget.Label
	link                         *widget.Hyperlink
	icon                         *canvas.Image

	screenshots  [5]*canvas.Image
	screenScroll *container.Scroll
}

func (w *welcome) loadAppDetail(app App) {
	w.shownID = app.ID
	w.shownPkg = app.Source.Package
	w.shownIcon = app.Icon

	w.name.SetText(app.Name)
	w.developer.SetText(app.Developer)
	w.version.SetText(app.Version)
	w.date.SetText(app.Date.Format("02 Jan 2006"))
	w.summary.SetText(app.Summary)

	w.icon.Resource = nil
	go setImageFromURL(w.icon, app.Icon)

	for i := 0; i < len(w.screenshots); i++ {
		w.screenshots[i].Resource = nil

		if i < len(app.Screenshots) {
			w.screenshots[i].Show()
			go setImageFromURL(w.screenshots[i], app.Screenshots[i].Image)
		} else {
			w.screenshots[i].Hide()
		}

		w.screenshots[i].Refresh()
	}
	w.screenScroll.ScrollToTop()

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
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
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
	i.icon.Move(fyne.NewPos(size.Width-i.icon.Size().Width, 0))
}

func (i *iconHoverLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return i.content.MinSize()
}

func loadWelcome(apps AppList, win fyne.Window) fyne.CanvasObject {
	w := &welcome{}
	w.name = widget.NewLabel("")
	w.developer = widget.NewLabel("")
	w.link = widget.NewHyperlink("", nil)
	w.summary = widget.NewLabel("")
	w.summary.Wrapping = fyne.TextWrapWord
	w.version = widget.NewLabel("")
	w.date = widget.NewLabel("")
	w.icon = &canvas.Image{}
	w.icon.FillMode = canvas.ImageFillContain
	makeScreenshots(w)

	dateAndVersion := fyne.NewContainerWithLayout(layout.NewGridLayout(2), w.date,
		widget.NewForm(&widget.FormItem{Text: "Version", Widget: w.version}))

	form := widget.NewForm(
		&widget.FormItem{Text: "Name", Widget: w.name},
		&widget.FormItem{Text: "Developer", Widget: w.developer},
		&widget.FormItem{Text: "Website", Widget: w.link},
		&widget.FormItem{Text: "Summary", Widget: w.summary},
		&widget.FormItem{Text: "Date", Widget: dateAndVersion},
	)
	details := fyne.NewContainerWithLayout(&iconHoverLayout{content: form, icon: w.icon}, form, w.icon)

	list := widget.NewList(func() int {
		return len(apps)
	},
		func() fyne.CanvasObject {
			return widget.NewLabel("A longish app name")
		},
		func(id int, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(apps[id].Name)
		})
	list.OnSelected = func(id int) {
		w.loadAppDetail(apps[id])
	}

	buttons := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("Install", func() {
			if w.shownPkg == "fyne.io/apps" {
				dialog.ShowInformation("System app", "Cannot overwrite the installer app", win)
				return
			}

			prog := dialog.NewProgressInfinite("Downloading...", "Please wait while the app is installed", win)
			prog.Show()
			get := commands.NewGetter()
			tmpIcon := downloadIcon(w.shownIcon)
			get.SetIcon(tmpIcon)
			get.SetAppID(w.shownID)
			err := get.Get(w.shownPkg)
			prog.Hide()
			if err != nil {
				dialog.ShowError(err, win)
			} else {
				dialog.ShowInformation("Installed", "App was installed successfully :)", win)
			}
			os.Remove(tmpIcon)
		}),
	)

	w.screenScroll = container.NewHScroll(container.NewHBox(
		w.screenshots[0], w.screenshots[1], w.screenshots[2], w.screenshots[3], w.screenshots[4]))
	if len(apps) > 0 {
		w.loadAppDetail(apps[0])
	}
	content := container.NewBorder(details, nil, nil, nil, w.screenScroll)
	return container.NewBorder(nil, nil, list, nil,
		container.NewBorder(nil, buttons, nil, nil, content))
}

func makeScreenshots(w *welcome) {
	for i := 0; i < len(w.screenshots); i++ {
		img := &canvas.Image{}
		img.SetMinSize(fyne.NewSize(320, 240))
		img.FillMode = canvas.ImageFillContain

		w.screenshots[i] = img
	}
}

func downloadIcon(url string) string {
	req, err := http.Get(url)
	if err != nil {
		fyne.LogError("Failed to access icon url: "+url, err)
		return ""
	}
	tmp := filepath.Join(os.TempDir(), "Fyne-Icon.png")
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fyne.LogError("Failed tread icon data", err)
		return ""
	}

	err = ioutil.WriteFile(tmp, data, 0666)
	if err != nil {
		fyne.LogError("Failed to get write icon to: "+tmp, err)
		return ""
	}

	return tmp
}
