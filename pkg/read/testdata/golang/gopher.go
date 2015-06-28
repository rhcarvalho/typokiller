package gopher

import "fmt"

// A Gopher is a cute animal.
type Gopher struct {
	Name string
	Age  int
}

// String implements the fmt.Stringer interfeice.
func (g *Gopher) String() string {
	return fmt.Sprintf("%v (%v)", g.Name, g.Age)
}
