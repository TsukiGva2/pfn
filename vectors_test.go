package main

import "testing"

func TestVectorLiterals(t *testing.T) {
	expected := "a"
	out := run(".avg (|vect| -> (/ (sum vect) (len vect)))\narr:=<1,2,3>\n(avg arr)\n").output

	if expected != out {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}
