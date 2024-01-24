package main

import (
	"testing"
)

func TestLoadController(t *testing.T) {
	if err := execToy(); err != nil {
		t.Fatal(err.Error())
	}
}
