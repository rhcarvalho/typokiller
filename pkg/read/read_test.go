package read

import "testing"

func TestBadPath(t *testing.T) {
	path := "testdata/bad/path"
	for _, dirReader := range []DirReader{GoFormat{}, AsciiDocFormat{}} {
		_, err := dirReader.ReadDir(path)
		if err == nil {
			t.Errorf("%#v.ReadDir(%q) returned err=%v, want nil", dirReader, path, err)
		}
	}
}
