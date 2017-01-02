package main

import "testing"

func TestStringContainsFunctionShouldFindStringInSlice(t *testing.T) {
	testslice := []string{"this", "that"}
	if !stringSliceContains(testslice, "that") {
		t.Fail()
	}
}

func TestStringContainsFunctionShouldNotFindStringNotInSlice(t *testing.T) {
	testslice := []string{"this", "that"}
	if stringSliceContains(testslice, "somethingelse") {
		t.Fail()
	}
}
