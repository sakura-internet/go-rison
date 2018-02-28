package rison_test

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"

	"github.com/sakura-internet/go-rison"
)

func ExampleQuote() {
	s := "~!*()-_.,:@$'/ \"#%&+;<=>?[\\]^`{|}"
	fmt.Println(rison.QuoteString(s))
	// Output: ~!*()-_.,:@$'/+%22%23%25%26%2B%3B%3C%3D%3E%3F%5B%5C%5D%5E%60%7B%7C%7D
}

func TestQuoteString(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	for i := byte(0); i < 128; i++ {
		buf.WriteByte(i)
	}
	s := buf.String()
	qs := rison.QuoteString(s)
	qb := rison.Quote([]byte(s))
	if string(qb) != qs {
		t.Errorf("unescaping %s .. : want %s, got %s", qs, string(qb))
	}
	u, err := url.QueryUnescape(qs)
	if err != nil {
		t.Errorf("unescaping %s .. : want %s, got error `%s`", qs, s, err.Error())
	}
	if u != s {
		t.Errorf("unescaping %s .. : want %s, got %s", qs, s, u)
	}
}
