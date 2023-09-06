package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

const keyInstallPrefix = "installed."

type App struct {
	ID, Name, Icon     string
	Developer, Summary string
	URL, Website       string
	Screenshots        []AppScreenshot

	Date    time.Time
	Version string

	Source   AppSource
	Requires string
}

type AppScreenshot struct {
	Image, Type string
}

type AppSource struct {
	Git, Package string
}

type AppList []App

func installedVersion(a App) string {
	return fyne.CurrentApp().Preferences().String(keyInstallPrefix + a.ID)
}

func markInstalled(a App) {
	ver := a.Version
	if ver == "" {
		ver = "latest"
	}
	fyne.CurrentApp().Preferences().SetString(keyInstallPrefix+a.ID, ver)
}

func parseAppList(reader io.Reader) (AppList, error) {
	decode := json.NewDecoder(reader)

	appList := AppList{}
	err := decode.Decode(&appList)
	if err != nil {
		return nil, err
	}

	appList = appList.filterCompatible()
	sort.Slice(appList, func(a, b int) bool {
		return appList[a].Name < appList[b].Name
	})

	sort.Slice(appList, func(i, j int) bool {
		return strings.Compare(strings.ToLower(appList[i].Name), strings.ToLower(appList[j].Name)) < 0
	})
	return appList, nil
}

func loadAppListFromWeb() (io.ReadCloser, error) {
	res, err := http.Get("https://apps.fyne.io/api/v1/list.json")
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// TODO make actual cache read()
func loadAppListFromCache() (io.ReadCloser, error) {
	res, err := os.Open(filepath.Join("testdata", "list.json"))
	if err != nil {
		return nil, err
	}

	return res, nil
}
