package read

import "testing"

func TestReadDirGo(t *testing.T) {
	path := "testdata/golang"
	pkgs, err := GoFormat{}.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(pkgs), 1; got != want {
		t.Fatalf("ReadDir(%q) got %d packages, want %d", path, got, want)
	}
	pkg := pkgs[0]
	if got, want := pkg.Name, "gopher"; got != want {
		t.Errorf("pkg.Name = %q, want %q", got, want)
	}
	idents := []string{
		"Age", "fmt", "g", "gopher", "Gopher", "int", "Name",
		"Sprintf", "string", "String",
	}
	if !haveSameElements(pkg.Identifiers, idents) {
		t.Errorf("pkg.Identifiers = %#v, want %#v", pkg.Identifiers, idents)
	}
	if got, want := len(pkg.Documentation), 3; got != want {
		t.Errorf("len(pkg.Documentation) = %d, want %d", got, want)
	}
}

func haveSameElements(a, b []string) bool {
	m := make(map[string]bool)
	for _, k := range a {
		m[k] = true
	}
	n := make(map[string]bool)
	for _, k := range b {
		n[k] = true
	}
	if len(m) != len(n) {
		return false
	}
	for k, v := range m {
		if n[k] != v {
			return false
		}
	}
	return true
}
