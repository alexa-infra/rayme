package math

import "math"

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

func Abs(value float64) float64 {
	if value < 0.0 {
		return -value
	}
	return value
}

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
