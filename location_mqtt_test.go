package main

import (
	"encoding/json"
	"testing"
)

func TestMarshallWorks(t *testing.T) {
	testMsg := "{\"_type\":\"location\",\"tid\":\"s5\",\"acc\":20,\"batt\":90,\"conn\":\"m\",\"doze\":false,\"lat\":51.7471862,\"lon\":-0.4734345,\"t\":\"u\",\"tst\":1483358150}"

	var locator MQTTMsg
	err := json.Unmarshal([]byte(testMsg), &locator)
	if err != nil {
		t.Logf("Error unmarshalling: %v", err)
		t.Fail()
	}
	if (locator.Type != "location") {
		t.Logf("Type. Expected: 'location'. Actual: %v", locator.Type)
		t.Fail()
	}
	if (locator.Accuracy != 20) {
		t.Logf("Accuracy. Expected: 20. Actual: %v", locator.Accuracy)
		t.Fail()
	}
}
