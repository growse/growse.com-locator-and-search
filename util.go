package main

func stringSliceContains(haystack []string, needle string) bool {
	for _, bitOfHay := range haystack {
		if bitOfHay == needle {
			return true
		}
	}
	return false
}
