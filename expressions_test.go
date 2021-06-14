package main

import "testing"

func TestExpr(t *testing.T) {
	expected := "def f(x):\n\treturn (f(((1*2*(2-3))+2+(3/2))))\n\nx=2\n(print((f(x))))\n"
	out := run(".f(|x|->(f(+(*1,2,(-2,3)),2,(/3,2))))x:=2(print(f x))").output

	if expected != out {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}
