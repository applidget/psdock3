package main

import (
	"strings"

	"github.com/applidget/psdock/stream"
)

// prefix args have the following format: --prefix some-prefix[:blue]
func parsePrefixArg(prefix string) (string, stream.Color) {
	comps := strings.Split(prefix, ":")
	if len(comps) == 1 {
		return comps[0], stream.NoColor
	}
	return comps[0], stream.MapColor(comps[len(comps)-1])
}
