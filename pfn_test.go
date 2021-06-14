package main

import (
	"io/ioutil"
	"testing"
)

func TestAvg(t *testing.T) {
	data, err := ioutil.ReadFile("test-expected/average.py")
	if err != nil {
		t.Error("error opening file")
		return
	}

	out := runFile("pfn-code/average.pfn")

	if string(data) != out {
		t.Errorf("error, expected:\n%s\ngot:\n%s\n", string(data), out)
	}
}

func TestPM(t *testing.T) {
	data, err := ioutil.ReadFile("test-expected/pattern.py")
	if err != nil {
		t.Error("error opening file")
		return
	}

	out := runFile("pfn-code/pattern.pfn")

	if string(data) != out {
		t.Errorf("error, expected:\n%s\ngot:\n%s\n", string(data), out)
	}
}

func TestSimple(t *testing.T) {
	data, err := ioutil.ReadFile("test-expected/simple.py")
	if err != nil {
		t.Error("error opening file")
		return
	}

	out := runFile("pfn-code/simple.pfn")

	if string(data) != out {
		t.Errorf("error, expected:\n%s\ngot:\n%s\n", string(data), out)
	}
}
