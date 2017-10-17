package main

import (
	"testing"
)

func TestIndexOfContains(t *testing.T) {
	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "wibble")
	if actual != 3 {
		t.Errorf("Failed to get index of item. Expected: 3 Actual: %v", actual)
	}
}

func TestIndexOfCaseInsensitivity(t *testing.T) {

	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "WIBBLE")
	if actual != 3 {
		t.Errorf("Failed to get index of item. Expected: 3 Actual: %v", actual)
	}
}
func TestIndexOfUnicode(t *testing.T) {
	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "ø∂˚¨˙ƒ´˚¨˙")
	if actual != 4 {
		t.Errorf("Failed to get index of unicode item. Expected: 3 Actual: %v", actual)
	}
}

func TestIndexOfNotFound(t *testing.T) {
	thingToTest := []string{"hello", "there", "awesomesauce", "wibble", "ø∂˚¨˙ƒ´˚¨˙"}
	actual := indexOfCaseInsensitive(thingToTest, "NotFound")
	if actual != -1 {
		t.Errorf("Failed to detet item not found. Expected: -1 Actual: %v", actual)
	}
}

func TestSmartTruncateWithHighlight(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	actual := SmartTruncateWithHighlight(inputString, "sit", 2, "...", "<b>%s</b>")
	expected := "...ipsum dolor <b>Sit</b> amet, consectetur..."
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}

func TestSmartTruncateWithHighlightNotFound(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	expected := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in..."
	actual := SmartTruncateWithHighlight(inputString, "wibble", 5, "...", "%s")
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}

func TestSmartTruncateWithHighlightStringStart(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	expected := "Lorem ipsum dolor Sit amet,..."
	actual := SmartTruncateWithHighlight(inputString, "ipsum", 2, "...", "%s")
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}

func TestSmartTruncateWithHighlightStringEnd(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	expected := "...placerat. Quisque at odio nunc."
	actual := SmartTruncateWithHighlight(inputString, "odio", 2, "...", "%s")
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}

func TestSmartTruncateWithHighlightStringLastWordWithPunctuation(t *testing.T) {
	inputString := "Lorem ipsum dolor Sit amet, consectetur adipiscing elit. In in erat pretium nisi ornare tempor. Phasellus molestie lectus tellus, a facilisis enim commodo at. Ut vel dui eu libero lacinia congue pretium et ex. Etiam commodo accumsan scelerisque. Suspendisse augue lorem, sodales id ex vel, scelerisque porta neque. Sed posuere sed ligula a accumsan. Nam pellentesque sodales nisl eu placerat. Quisque at odio nunc."
	expected := "...placerat. Quisque at odio nunc."
	actual := SmartTruncateWithHighlight(inputString, "nunc", 2, "...", "%s")
	if actual != expected {
		t.Errorf("Incorrect truncating. Expected \"%s\". Actual: \"%s\"", expected, actual)
	}
}
