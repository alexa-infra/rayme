package math

import "math/rand"

var (
	RandGen = rand.New(rand.NewSource(99))
)

func RandomBetween(a, b float64) float64 {
	// [0, 1] -> [0, b - a] -> [a, b]
	return RandGen.Float64()*(b-a) + a
}

func RandomInUnitSphere() *Vec3 {
	for {
		x := RandomBetween(-1.0, 1.0)
		y := RandomBetween(-1.0, 1.0)
		z := RandomBetween(-1.0, 1.0)
		v := &Vec3{x, y, z}
		if v.Length2() >= 1.0 {
			continue
		}
		return v
	}
}

func RandomInUnitDisk() *Vec3 {
	for {
		x := RandomBetween(-1.0, 1.0)
		y := RandomBetween(-1.0, 1.0)
		v := &Vec3{x, y, 0.0}
		if v.Length2() >= 1.0 {
			continue
		}
		return v
	}
}

func RandomUnitVector() *Vec3 {
	return RandomInUnitSphere().Normalize()
}
