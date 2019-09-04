package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func loadAppListFromTestData() (io.ReadCloser, error) {
	res, err := os.Open(filepath.Join("testdata", "list.json"))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func TestParseAppList(t *testing.T) {
	res, err := loadAppListFromTestData()
	if err != nil {
		t.Error("Error loading app list", err)
	}
	defer res.Close()
	list, err := parseAppList(res)
	if err != nil {
		t.Error("Error parsing app list", err)
	}

	assert.Equal(t, 10, len(list))

	app := list[0]
	assert.Equal(t, "xyz.andy.beebui", app.ID)
	assert.Equal(t, "beebUI", app.Name)
	assert.Equal(t, "https://github.com/andydotxyz/beebui/blob/master/beebui.png?raw=true", app.Img)
	assert.Equal(t, time.Date(2019, 03, 17, 19, 32, 14, 0, time.Local), app.Date)
	assert.Equal(t, "A BBC Micro Emulator based on Fyne and skx/gobasic", app.Excerpt)
	assert.Equal(t, "https://apps.fyne.io/apps/beebui.html", app.URL)
	assert.Equal(t, "https://github.com/andydotxyz/beebui", app.Homepage)

	assert.Equal(t, "https://github.com/andydotxyz/beebui.git", app.Git)
	assert.Equal(t, "master", app.Tag)
	assert.Equal(t, "cmd/beebui", app.Dir)
	assert.Equal(t, "", app.Version)
}
