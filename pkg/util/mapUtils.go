package util

func MergeMaps(dst *map[string]string, src *map[string]string) *map[string]string {
	if dst == nil || *dst == nil {
		return src
	}
	if src == nil || *src == nil {
		return dst
	}
	for k, v := range *src {
		(*dst)[k] = v
	}
	return dst
}
