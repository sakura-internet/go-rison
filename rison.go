// Copyright 2018 SAKURA Internet.

// Package rison implements encoding and decoding of Rison.
// https://github.com/Nanonid/rison
//
// Rison is a data serialization format optimized for compactness
// in URIs. Rison is a slight variation of JSON that looks vastly
// superior after URI encoding. Rison still expresses exactly the
// same set of data structures as JSON, so data can be translated
// back and forth without loss or guesswork.
package rison

const (
	notIdChar        = ` '!:(),*@$`
	notIdStart       = notIdChar + `0123456789-`
	parserWhitespace = " \t\n\r\f"
)

type Mode int

const (
	Mode_Rison Mode = iota
	Mode_ORison
	Mode_ARison
)
