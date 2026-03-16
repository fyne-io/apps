package main

import (
	"context"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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
	shownApp            App
	name, summary, date *widget.Label
	developer, version  *widget.Label
	link                *widget.Hyperlink
	icon                *canvas.Image

	screenshots      [5]*canvas.Image
	screenScroll     *container.Scroll
	install          *widget.Button
	iconLoader       imageLoader
	screenshotLoader [5]imageLoader
}

func (w *welcome) loadAppDetail(app App) {
	w.shownApp = app

	w.name.SetText(app.Name)
	w.developer.SetText(app.Developer)
	w.version.SetText(app.Version)
	w.date.SetText(app.Date.Format("02 Jan 2006"))
	w.summary.SetText(app.Summary)

	w.icon.Resource = nil
	w.icon.Image = nil
	w.icon.Refresh()
	for _, scr := range w.screenshots {
		scr.Resource = nil
		scr.Image = nil
		scr.Refresh()
		scr.Hide()
	}

	w.iconLoader.load(w.icon, app.Icon)

	for i := 0; i < len(w.screenshots); i++ {
		w.screenshots[i].Resource = nil
		w.screenshots[i].Image = nil
		w.screenshots[i].Refresh()

		if i < len(app.Screenshots) {
			w.screenshots[i].Show()
			w.screenshotLoader[i].load(w.screenshots[i], app.Screenshots[i].Image)
		} else {
			w.screenshotLoader[i].load(w.screenshots[i], "") // cancel any in-flight
			w.screenshots[i].Hide()
		}
	}
	w.screenScroll.ScrollToTop()

	parsed, err := url.Parse(app.Website)
	if err != nil {
		w.link.SetText("")
		return
	}
	w.link.SetText(parsed.Host)
	w.link.SetURL(parsed)

	installedVer := installedVersion(app)
	installed := installedVer != "" && installedVer == app.Version
	if installed || app.Source.Package == "fyne.io/apps" {
		w.install.SetText("Installed")
		w.install.Disable()
	} else if installedVer == "" {
		w.install.SetText("Install")
		w.install.Enable()
	} else {
		w.install.SetText("Upgrade")
		w.install.Enable()
	}
}

// imageLoader manages cancellable image loading for a single canvas.Image.
// Calling load cancels any in-flight request for this image before starting a new one.
type imageLoader struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (l *imageLoader) load(img *canvas.Image, location string) {
	l.mu.Lock()
	if l.cancel != nil {
		l.cancel()
	}
	img.Resource = nil
	img.Refresh()
	if location == "" {
		l.cancel = nil
		l.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel
	l.mu.Unlock()

	go func() {
		res, err := fyne.CacheResourceFromURLString(location)
		if ctx.Err() != nil {
			return // cancelled, don't touch the image
		}
		fyne.Do(func() {
			if err != nil {
				img.Resource = theme.WarningIcon()
			} else {
				img.Resource = res
			}
			img.Refresh()
		})
	}()
}

// appCardLayout positions an icon centered above a label
type appCardLayout struct {
	icon, label fyne.CanvasObject
}

func (a *appCardLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	iconSize := float32(48)
	labelH := a.label.MinSize().Height
	totalH := iconSize + 4 + labelH
	topPad := (size.Height - totalH) / 2
	if topPad < 4 {
		topPad = 4
	}

	a.icon.Resize(fyne.NewSize(iconSize, iconSize))
	a.icon.Move(fyne.NewPos((size.Width-iconSize)/2, topPad))

	a.label.Resize(fyne.NewSize(size.Width, labelH))
	a.label.Move(fyne.NewPos(0, topPad+iconSize+4))
}

func (a *appCardLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(110, 100)
}

func makeAppCardTemplate() fyne.CanvasObject {
	icon := &canvas.Image{}
	icon.SetMinSize(fyne.NewSquareSize(48))
	icon.FillMode = canvas.ImageFillContain

	name := widget.NewLabel("Template")
	name.Alignment = fyne.TextAlignCenter
	name.Truncation = fyne.TextTruncateEllipsis

	cardContent := container.New(&appCardLayout{icon: icon, label: name}, icon, name)
	return container.NewPadded(cardContent)
}

// cardLoaders tracks an imageLoader per grid cell icon so recycled cells cancel prior loads.
var cardLoaders sync.Map

func updateAppCard(obj fyne.CanvasObject, app App) {
	padded := obj.(*fyne.Container)
	card := padded.Objects[0].(*fyne.Container) // the appCardLayout container
	icon := card.Objects[0].(*canvas.Image)
	label := card.Objects[1].(*widget.Label)

	label.SetText(app.Name)
	icon.Resource = nil
	icon.Image = nil
	icon.Refresh()

	val, _ := cardLoaders.LoadOrStore(icon, &imageLoader{})
	loader := val.(*imageLoader)
	loader.load(icon, app.Icon)
}

// detailHeaderLayout places the icon to the left with info to its right
type detailHeaderLayout struct {
	icon, info fyne.CanvasObject
}

func (d *detailHeaderLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	iconSize := float32(80)
	d.icon.Resize(fyne.NewSize(iconSize, iconSize))
	d.icon.Move(fyne.NewPos(0, 0))

	infoW := size.Width - iconSize - 12
	if infoW < 0 {
		infoW = 0
	}
	d.info.Resize(fyne.NewSize(infoW, size.Height))
	d.info.Move(fyne.NewPos(iconSize+12, 0))
}

func (d *detailHeaderLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	infoMin := d.info.MinSize()
	h := infoMin.Height
	if h < 80 {
		h = 80
	}
	return fyne.NewSize(80+12+infoMin.Width, h)
}

func loadUI(apps AppList, win fyne.Window) fyne.CanvasObject {
	w := &welcome{}
	w.name = widget.NewLabel("")
	w.name.TextStyle.Bold = true
	w.developer = widget.NewLabel("")
	w.link = widget.NewHyperlink("", nil)
	w.summary = widget.NewLabel("")
	w.summary.Wrapping = fyne.TextWrapWord
	w.version = widget.NewLabel("")
	w.date = widget.NewLabel("")
	w.icon = &canvas.Image{}
	w.icon.SetMinSize(fyne.NewSquareSize(80))
	w.icon.FillMode = canvas.ImageFillContain
	makeScreenshots(w)

	// Build the detail view
	w.install = widget.NewButton("Install", nil) // callback set below
	w.install.Importance = widget.HighImportance

	infoBlock := container.NewVBox(
		w.name,
		w.developer,
		container.NewHBox(w.link, layout.NewSpacer(), w.install),
	)
	w.name.TextStyle.Bold = true

	header := container.New(&detailHeaderLayout{icon: w.icon, info: infoBlock}, w.icon, infoBlock)

	metaLine := container.NewHBox(
		widget.NewLabel("Version:"),
		w.version,
		layout.NewSpacer(),
		w.date,
	)

	w.screenScroll = container.NewHScroll(container.NewHBox(
		w.screenshots[0], w.screenshots[1], w.screenshots[2], w.screenshots[3], w.screenshots[4]))

	detailContent := container.NewVBox(
		header,
		widget.NewSeparator(),
		w.summary,
		widget.NewSeparator(),
		metaLine,
		w.screenScroll,
	)
	detailScroll := container.NewVScroll(detailContent)

	// Build the category tabs and app grid
	categories := sortedCategories(apps)
	currentSort := sortByDate
	currentCategory := ""
	filteredIDs := sortedApps(apps, currentCategory, currentSort)

	var detailView, browseView *fyne.Container

	appGrid := widget.NewGridWrap(
		func() int {
			return len(filteredIDs)
		},
		makeAppCardTemplate,
		func(id widget.GridWrapItemID, obj fyne.CanvasObject) {
			if int(id) < len(filteredIDs) {
				updateAppCard(obj, apps[filteredIDs[id]])
			}
		},
	)
	appGrid.OnSelected = func(id widget.GridWrapItemID) {
		if int(id) < len(filteredIDs) {
			appID := filteredIDs[id]
			w.loadAppDetail(apps[appID])
			detailView.Show()
			browseView.Hide()
		}
		appGrid.UnselectAll()
	}

	filterGrid := func(category string) {
		currentCategory = category
		filteredIDs = sortedApps(apps, currentCategory, currentSort)
		appGrid.Refresh()
		appGrid.ScrollToTop()
	}

	catTabs := container.NewAppTabs(
		container.NewTabItem("All", layout.NewSpacer()),
	)
	for _, cat := range categories {
		title := strings.ToUpper(cat[:1]) + cat[1:]
		catTabs.Append(container.NewTabItem(title, layout.NewSpacer()))
	}
	catTabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "All" {
			filterGrid("")
		} else {
			filterGrid(strings.ToLower(tab.Text))
		}
	}

	// Featured banner
	featuredBanner := makeFeatured(apps, func(id string) {
		w.loadAppDetail(apps[id])
		detailView.Show()
		browseView.Hide()
	})

	// Section headers
	featuredHeader := makeSectionHeader("Featured")
	browseHeader := makeSectionHeader("Browse Apps")

	sortSelect := widget.NewSelect([]string{sortByDate, sortByName}, func(val string) {
		currentSort = val
		filteredIDs = sortedApps(apps, currentCategory, currentSort)
		appGrid.Refresh()
		appGrid.ScrollToTop()
	})
	sortSelect.Selected = currentSort

	browseRow := container.NewHBox(browseHeader, layout.NewSpacer(), sortSelect)

	topSection := container.NewVBox(
		featuredHeader,
		featuredBanner,
		widget.NewSeparator(),
		browseRow,
		catTabs,
	)

	browseView = container.NewBorder(topSection, nil, nil, nil, appGrid)

	backBtn := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		browseView.Show()
		detailView.Hide()
	})
	backBtn.Importance = widget.LowImportance

	detailView = container.NewBorder(
		container.NewHBox(backBtn),
		nil, nil, nil,
		detailScroll,
	)

	// Wire up install button
	w.install.OnTapped = func() {
		bar := widget.NewProgressBarInfinite()
		content := container.NewVBox(widget.NewLabel("Please wait while the app is installed"), bar)
		prog := dialog.NewCustomWithoutButtons("Downloading...", content, win)
		bar.Start()
		prog.Show()

		go func() {
			get := commands.NewGetter()
			tmpIcon := downloadIcon(w.shownApp.Icon)
			get.SetIcon(tmpIcon)
			get.SetAppID(w.shownApp.ID)
			err := get.Get(w.shownApp.Source.Package)

			fyne.Do(func() {
				prog.Hide()
				bar.Stop()
				if err != nil {
					dialog.ShowError(err, win)
				} else {
					dialog.ShowInformation("Installed", "App was installed successfully :)", win)
					markInstalled(w.shownApp)
					w.loadAppDetail(w.shownApp)
				}
				_ = os.Remove(tmpIcon)
			})
		}()
	}

	detailView.Hide()
	return container.NewStack(browseView, detailView)
}

func makeSectionHeader(title string) fyne.CanvasObject {
	label := canvas.NewText(title, color.NRGBA{R: 100, G: 149, B: 237, A: 255})
	label.TextSize = 18
	label.TextStyle.Bold = true
	return container.NewPadded(label)
}

func makeScreenshots(w *welcome) {
	for i := 0; i < len(w.screenshots); i++ {
		img := &canvas.Image{}
		img.SetMinSize(fyne.NewSize(320, 240))
		img.FillMode = canvas.ImageFillContain

		w.screenshots[i] = img
	}
}

func sortedCategories(apps AppList) []string {
	catMap := map[string]bool{}
	for _, a := range apps {
		if a.Category != "" {
			catMap[a.Category] = true
		}
	}
	var cats []string
	for c := range catMap {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

const (
	sortByDate = "Newest"
	sortByName = "A-Z"
)

func sortedApps(apps AppList, category, sortMode string) []string {
	var ids []string
	for id, a := range apps {
		if category == "" || a.Category == category {
			ids = append(ids, id)
		}
	}
	switch sortMode {
	case sortByName:
		sort.Slice(ids, func(i, j int) bool {
			return strings.ToLower(apps[ids[i]].Name) < strings.ToLower(apps[ids[j]].Name)
		})
	default: // sortByDate – newest first
		sort.Slice(ids, func(i, j int) bool {
			return apps[ids[i]].Date.After(apps[ids[j]].Date)
		})
	}
	return ids
}

func downloadIcon(url string) string {
	req, err := http.Get(url)
	if err != nil {
		fyne.LogError("Failed to access icon url: "+url, err)
		return ""
	}
	tmp := filepath.Join(os.TempDir(), "Fyne-Icon.png")
	data, err := io.ReadAll(req.Body)
	if err != nil {
		fyne.LogError("Failed tread icon data", err)
		return ""
	}

	err = os.WriteFile(tmp, data, 0666)
	if err != nil {
		fyne.LogError("Failed to get write icon to: "+tmp, err)
		return ""
	}

	return tmp
}
