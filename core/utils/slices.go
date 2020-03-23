package utils

func StringInSlice(search string, slice []string) bool {
	for _, elem := range slice {
		if search == elem {
			return true
		}
	}
	return false
}
