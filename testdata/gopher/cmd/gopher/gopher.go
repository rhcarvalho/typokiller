package main

import (
	"github.com/rhcarvalho/typokiller/testdata/gopher"
	"github.com/rhcarvalho/typokiller/testdata/gopher/hello"
)

func main() {
	hello.Hello(&gopher.Gopher{"Typokiller", 42})
}
