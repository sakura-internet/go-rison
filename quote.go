package rison

import (
	"net/url"
	"regexp"
)

var escapeRx = regexp.MustCompile(`%[\da-fA-F]{2}`)
var escapeTable = map[string]string{
	"%7E": "~",
	"%21": "!",
	"%2A": "*",
	"%28": "(",
	"%29": ")",
	"%2D": "-",
	"%5F": "_",
	"%2E": ".",
	"%2C": ",",
	"%3A": ":",
	"%40": "@",
	"%24": "$",
	"%27": "'",
	"%2F": "/",
	"%20": "+",
}

// QuoteString is like "net/url".QueryEscape but quotes fewer characters.
func QuoteString(s string) string {
	return escapeRx.ReplaceAllStringFunc(url.QueryEscape(s), func(m string) string {
		r, ok := escapeTable[m]
		if !ok {
			r = m
		}
		return r
	})
}

// Quote is like "net/url".QueryEscape but quotes fewer characters.
func Quote(s []byte) []byte {
	return []byte(QuoteString(string(s)))
}
