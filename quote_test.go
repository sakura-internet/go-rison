package rison_test

import (
	"fmt"

	"github.com/sakura-internet/go-rison/v4"
)

func ExampleQuote() {
	s := "~!*()-_.,:@$'/ \"#%&+;<=>?[\\]^`{|}"
	fmt.Println(rison.QuoteString(s))
	// Output: ~!*()-_.,:@$'/+%22%23%25%26%2B%3B%3C%3D%3E%3F%5B%5C%5D%5E%60%7B%7C%7D
}
