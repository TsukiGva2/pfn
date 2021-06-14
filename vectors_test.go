package main

import (
	"testing"
)

func TestVectorLiterals(t *testing.T) {
	expected := "def avg(vect):\n\treturn ((sum(vect))/(len(vect)))\n\narr=[1,2,3]\n(avg(arr))\n"
	out := run(".avg(|vect|->(/(sum vect)(len vect)))arr:=<1,2,3>(avg arr)").output

	if expected != out {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, out)
	}
}
