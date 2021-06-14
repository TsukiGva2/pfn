package main

import "testing"

func VectorLiteralsTest(t *testing.T) {
	expected := "a"
	out := run(".avg (|vect| -> (/ (sum vect), (len vect)))").output

	if expected != out {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}
