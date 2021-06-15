package main

import (
	"io/ioutil"
	"testing"
)

func makeTest(t *testing.T, filename string) {
	data, err := ioutil.ReadFile("test-expected/" + filename + ".py")
	if err != nil {
		t.Error("error opening file")
		return
	}

	out := runFile("pfn-code/" + filename + ".pfn")

	if string(data) != out {
		t.Errorf("error, expected:\n%s\ngot:\n%s\n", string(data), out)
	}
}

func TestAvg(t *testing.T) {
	makeTest(t, "average")
}

func TestPM(t *testing.T) {
	makeTest(t, "pattern")
}

func TestSimple(t *testing.T) {
	makeTest(t, "simple")
}

func TestImport(t *testing.T) {
	makeTest(t, "import")
}
