package main

import (
	"html"
	"strings"
)

//EscapeSpecialChars avoid sqlInjection
func EscapeSpecialChars(inp string) string {
	if len(inp) == 0 {
		return ""
	}
	toReplace := []string{"'", "`", "\"", "#", ",", "/", "!", "=", "*", ";", "\\", "//", ":"}
	for _, i := range toReplace {
		inp = strings.ReplaceAll(inp, i, "")
	}
	return html.EscapeString(inp)
}
