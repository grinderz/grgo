package utils

func MapKeysAsStrings(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	var i uint
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
