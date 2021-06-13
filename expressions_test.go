package main

import "testing"

func TestExpr(t *testing.T) {
	expected := "def f(x):\n\treturn f((1*2*(2-3))+2+(3/2))\nx=2\nprint(f(x))"
	out := runFile("test.exc")

	if expected != out {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}
