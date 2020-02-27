package features

// containsFeature checks if the given feature is present in the
// list of features.
func containsFeature(features []Feature, f Feature) bool {
	for _, feature := range features {
		if f == feature {
			return true
		}
	}
	return false
}
