package typokiller

import "testing"

func TestReadDirGo(t *testing.T) {
	path := "testdata/gopher"
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

func TestReadDirAsciiDoc(t *testing.T) {
	path := "testdata/asciidoc"
	pkgs, err := AsciiDocFormat{}.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(pkgs), 1; got != want {
		t.Fatalf("ReadDir(%q) got %d packages, want %d", path, got, want)
	}
	pkg := pkgs[0]
	if got, want := pkg.Name, "README.adoc"; got != want {
		t.Errorf("pkg.Name = %q, want %q", got, want)
	}
	if pkg.Identifiers != nil {
		t.Errorf("pkg.Identifiers = %#v, want nil", pkg.Identifiers)
	}
	if got, want := len(pkg.Documentation), 6; got != want {
		t.Errorf("len(pkg.Documentation) = %d, want %d", got, want)
	}
	for i, offsetLineColumn := range [][3]int{
		{0, 1, 1},
		{76, 5, 1},
		{156, 8, 1},
		{326, 13, 1},
		{345, 15, 1},
		{373, 19, 1},
	} {
		if got, want := pkg.Documentation[i].Position.Offset, offsetLineColumn[0]; got != want {
			t.Errorf("pkg.Documentation[%d].Offset = %d, want %d", i, got, want)
		}
		if got, want := pkg.Documentation[i].Position.Line, offsetLineColumn[1]; got != want {
			t.Errorf("pkg.Documentation[%d].Line = %d, want %d", i, got, want)
		}
		if got, want := pkg.Documentation[i].Position.Column, offsetLineColumn[2]; got != want {
			t.Errorf("pkg.Documentation[%d].Column = %d, want %d", i, got, want)
		}
	}
}

func TestBadPath(t *testing.T) {
	path := "testdata/bad/path"
	for _, readDirer := range []ReadDirer{GoFormat{}, AsciiDocFormat{}} {
		_, err := readDirer.ReadDir(path)
		if err == nil {
			t.Errorf("%#v.ReadDir(%q) returned err=%v, want nil", readDirer, path, err)
		}
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
