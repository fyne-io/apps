package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	ID, Name, Icon     string
	Developer, Summary string
	URL, Website       string

	Date    time.Time
	Version string

	Source AppSource
}

type AppSource struct {
	Git, Tag, Dir string

}
type AppList []App

func parseAppList(reader io.Reader) (AppList, error) {
	decode := json.NewDecoder(reader)

	appList := AppList{}
	err := decode.Decode(&appList)
	if err != nil {
		return nil, err
	}

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
