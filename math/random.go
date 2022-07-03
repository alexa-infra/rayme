package math

import (
	"math"
	"math/rand"
)

type RandExt struct {
	*rand.Rand
}

func MakeRandExt(seed int) *RandExt {
	randGen := rand.New(rand.NewSource(99))
	return &RandExt{ randGen }
}

func (this *RandExt) Between(a, b float64) float64 {
	return this.Rand.Float64()*(b-a) + a
}

func (this *RandExt) RandomInUnitSphere() *Vec3 {
	theta := this.Between(0, 2 * math.Pi)
	phi := this.Between(0, math.Pi)
	r := this.Between(0, 1)
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
	x := r * sinPhi * cosTheta
	y := r * sinPhi * sinTheta
	z := r * cosPhi
	return &Vec3{ x, y, z }
}

func (this *RandExt) RandomInUnitDisk() *Vec3 {
	angle := this.Between(0, 2 * math.Pi)
	r := this.Between(0, 1)
	x := r * math.Cos(angle)
	y := r * math.Sin(angle)
	return &Vec3{x, y, 0.0}
}

func (this *RandExt) RandomUnitVector() *Vec3 {
	theta := this.Between(0, 2 * math.Pi)
	phi := this.Between(0, math.Pi)
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
	x := sinPhi * cosTheta
	y := sinPhi * sinTheta
	z := cosPhi
	return &Vec3{ x, y, z }
}
