package controller

import "testing"

func TestVersionCompare(t *testing.T) {
	v001 := MustParseVersion("0.0.1")
	v010 := MustParseVersion("0.1.0")
	v100 := MustParseVersion("1.0.0")

	if !v010.GT(v001) {
		t.Fatal("0.1.0 should be greater than 0.0.1")
	}
	if !v100.GT(v010) {
		t.Fatal("1.0.0 should be greater than 0.1.0")
	}
	if !v100.GTE(v100) {
		t.Fatal("version should be greater than or equal to itself")
	}
	if v001.String() != "0.0.1" {
		t.Fatalf("version string mismatch: %s", v001.String())
	}
}
