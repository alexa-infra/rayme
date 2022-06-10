package math

import (
	"math"
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

func RandomInt(n int) int {
	mutex.Lock()
	defer mutex.Unlock()
	return randGen.Intn(n)
}

func RandomInUnitSphere() *Vec3 {
	theta := RandomBetween(0, 2 * math.Pi)
	phi := RandomBetween(0, math.Pi)
	r := RandomBetween(0, 1)
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
	x := r * sinPhi * cosTheta
	y := r * sinPhi * sinTheta
	z := r * cosPhi
	return &Vec3{ x, y, z }
}

func RandomInUnitDisk() *Vec3 {
	angle := RandomBetween(0, 2 * math.Pi)
	r := RandomBetween(0, 1)
	x := r * math.Cos(angle)
	y := r * math.Sin(angle)
	return &Vec3{x, y, 0.0}
}

func RandomUnitVector() *Vec3 {
	theta := RandomBetween(0, 2 * math.Pi)
	phi := RandomBetween(0, math.Pi)
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
	x := sinPhi * cosTheta
	y := sinPhi * sinTheta
	z := cosPhi
	return &Vec3{ x, y, z }
}
