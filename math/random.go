package math

import "math/rand"

var (
	RandGen = rand.New(rand.NewSource(99))
)

func RandomInUnitSphere() *Vec3 {
	for {
		// [0, 1] -> [-0.5, 0.5] -> [-1, 1]
		x := (RandGen.Float64() - 0.5) * 2.0
		y := (RandGen.Float64() - 0.5) * 2.0
		z := (RandGen.Float64() - 0.5) * 2.0
		v := &Vec3{ x, y, z }
		if v.Length2() >= 1.0 {
			continue
		}
		return v
	}
}

func RandomInUnitDisk() *Vec3 {
	for {
		x := (RandGen.Float64() - 0.5) * 2.0
		y := (RandGen.Float64() - 0.5) * 2.0
		v := &Vec3{ x, y, 0.0 }
		if v.Length2() >= 1.0 {
			continue
		}
		return v
	}
}

func RandomUnitVector() *Vec3 {
	return RandomInUnitSphere().Normalize()
}
