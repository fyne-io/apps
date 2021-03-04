package main

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_FilterCompatible(t *testing.T) {
	l := AppList{
		App{Requires: runtime.GOOS},
		App{Requires: runtime.GOOS + ",powerpc"},
		App{Requires: "powerpc"},
		App{},
	}

	assert.Equal(t, 4, len(l))
	assert.Equal(t, 3, len(l.filterCompatible()))
}

func TestApp_IsCompatible(t *testing.T) {
	a := &App{Requires: "linux"}

	if runtime.GOOS == "linux" {
		assert.True(t, a.isCompatible())
	} else {
		assert.False(t, a.isCompatible())
	}

	a.Requires = runtime.GOOS
	assert.True(t, a.isCompatible())
}

func TestApp_IsCompatibleWithOS(t *testing.T) {
	a := &App{Requires: "darwin"}
	assert.True(t, a.isCompatibleWithOS("darwin"))
	assert.False(t, a.isCompatibleWithOS("powerpc"))

	a.Requires = "linux,powerpc"
	assert.True(t, a.isCompatibleWithOS("powerpc"))
	a.Requires = ""
	assert.True(t, a.isCompatibleWithOS("powerpc"))
}
