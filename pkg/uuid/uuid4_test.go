package uuid

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestUUID4(t *testing.T) {
	tests := []struct {
		bytes   []byte
		uuid4   []byte
		encoded []byte
	}{
		{
			bytes: []byte{
				0x82, 0xf2, 0x24, 0x1b,
				0xc7, 0x77, 0x14, 0x89,
				0x32, 0x72, 0x45, 0x57,
				0x68, 0x7c, 0x5c, 0x7d},
			uuid4: []byte{
				0x82, 0xf2, 0x24, 0x1b,
				0xc7, 0x77, 0x44, 0x89,
				0xb2, 0x72, 0x45, 0x57,
				0x68, 0x7c, 0x5c, 0x7d},
			encoded: []byte("82f2241b-c777-4489-b272-4557687c5c7d"),
		},
		{
			bytes: []byte{
				0x3b, 0x67, 0x83, 0xc6,
				0xec, 0x3b, 0x7f, 0xe8,
				0x68, 0x17, 0x17, 0x8d,
				0x45, 0xf1, 0x79, 0x95},
			uuid4: []byte{
				0x3b, 0x67, 0x83, 0xc6,
				0xec, 0x3b, 0x4f, 0xe8,
				0xa8, 0x17, 0x17, 0x8d,
				0x45, 0xf1, 0x79, 0x95},
			encoded: []byte("3b6783c6-ec3b-4fe8-a817-178d45f17995"),
		},
	}
	for _, test := range tests {
		if got, want := uuid4(test.bytes), test.uuid4; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: uuid4(): %s", test.encoded,
				stringDiff(fmt.Sprintf("%x", got), fmt.Sprintf("%x", want)))
		}
		if got, want := format(test.uuid4), test.encoded; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: format(): %s", test.encoded,
				stringDiff(string(got), string(want)))
		}
	}
}

func stringDiff(a, b string) string {
	max := len(a)
	if len(b) > len(a) {
		max = len(b)
	}
	min := len(a)
	if len(b) < len(a) {
		min = len(b)
	}
	equal := true
	diff := bytes.Repeat([]byte{' '}, max)
	for i := 0; i < min; i++ {
		if a[i] != b[i] {
			equal = false
			diff[i] = '|'
		}
	}
	for i := min; i < max; i++ {
		diff[i] = '|'
	}
	if equal {
		return "equal"
	}
	return fmt.Sprintf("content differ:\ngot:\t%s\n\t%s\nwant:\t%s", a, diff, b)
}
