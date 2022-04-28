package util

func MergeMaps(x *map[string]string,  y *map[string]string) *map[string]string {
	if x == nil{
		return y
	}
	if y == nil{
		return x
	}
	for k, v := range *y {
		(*x)[k] = v
	}
	return x
}