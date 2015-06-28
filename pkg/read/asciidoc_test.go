package read

import "testing"

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
