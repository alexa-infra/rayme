package math

import (
	"math/rand"
	"sync"
)

var (
	randGen = rand.New(rand.NewSource(99))
	mutex   = new(sync.Mutex)
)

func RandomBetween(a, b float64) float64 {
	mutex.Lock()
	defer mutex.Unlock()
	return randGen.Float64()*(b-a) + a
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
