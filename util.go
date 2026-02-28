package main

import (
	"strconv"
	"strings"
)

func itoa(v int) string {
	return strconv.Itoa(v)
}

func titleCase(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}
