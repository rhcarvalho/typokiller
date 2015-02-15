// Package hello is just used for testing typokiller.
package hello

import (
	"fmt"

	"github.com/rhcarvalho/typokiller/testdata/gopher"
)

/*
	Hello says helo to a Gophr.
	It can be used with Gophers of any age.
*/
func Hello(gopher *gopher.Gopher) {
	fmt.Printf("Hello, %v!\n", gopher)
}
