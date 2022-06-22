package util

func MergeMaps(dst map[string]string, src map[string]string) map[string]string {
	if dst == nil {
		return src
	}
	if src == nil {
		return dst
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
