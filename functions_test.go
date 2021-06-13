package main

import "testing"

func TestFn(t *testing.T) {
	expected := "def f(x):\n\treturn whatever\n\n"
	out := run(".f(|x|->whatever)").output

	if expected != out {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, out)
	}
}

func TestFnScope(t *testing.T) {
	expected := "def f(x):\n\treturn whatever\n\nouter=2\n"
	out := run(".f(|x|->whatever)outer:=2").output

	if expected != out {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, out)
	}
}
