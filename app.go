package main

import (
	"runtime"
	"strings"
)

func (l AppList) filterCompatible() AppList {
	ret := make(AppList, 0)
	for _, v := range l {
		if v.isCompatible() {
			ret[v.ID] = v
		}
	}
	return ret
}

func (a App) isCompatible() bool {
	return a.isCompatibleWithOS(runtime.GOOS)
}

func (a App) isCompatibleWithOS(os string) bool {
	if a.Requires == "" {
		return true
	}

	requires := strings.Split(a.Requires, ",")
	for _, r := range requires {
		if r == os {
			return true
		}
	}

	return false
}
