package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
)

const keyInstallPrefix = "installed."

type App struct {
	ID, Name, Icon         string
	Developer, Summary     string
	URL, Website, Category string
	Screenshots            []AppScreenshot

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

type AppList map[string]App

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

	var list []App
	err := decode.Decode(&list)
	if err != nil {
		return nil, err
	}

	appList := AppList{}
	for _, a := range list {
		appList[a.ID] = a
	}

	return appList.filterCompatible(), nil
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
