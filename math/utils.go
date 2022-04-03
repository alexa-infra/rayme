package math

func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func Clamp(value, a, b float64) float64 {
	if value < a {
		return a
	}
	if value > b {
		return b
	}
	return value
}
