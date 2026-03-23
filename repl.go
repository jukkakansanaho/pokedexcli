package main

import "strings"

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	return strings.Fields(lower)
}
