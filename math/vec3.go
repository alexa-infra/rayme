package math

import (
	"image/color"
	"math"
)

const (
	eps = 1e-8
)

type Vec3 struct {
	X, Y, Z float64
}

func (v *Vec3) AsColor() color.Color {
	return color.RGBA64{
		uint16(0xffff * Clamp(v.X, 0.0, 1.0)),
		uint16(0xffff * Clamp(v.Y, 0.0, 1.0)),
		uint16(0xffff * Clamp(v.Z, 0.0, 1.0)),
		uint16(0xffff),
	}
}

func (v *Vec3) AsPoint3() *Point3 {
	return MakePoint3(v.X, v.Y, v.Z)
}

func (v *Vec3) Mul(t float64) *Vec3 {
	return &Vec3{v.X * t, v.Y * t, v.Z * t}
}

func (v *Vec3) MulVec(a *Vec3) *Vec3 {
	return &Vec3{v.X * a.X, v.Y * a.Y, v.Z * a.Z}
}

func (v *Vec3) Add(a *Vec3) *Vec3 {
	return &Vec3{v.X + a.X, v.Y + a.Y, v.Z + a.Z}
}

func (v *Vec3) Sub(a *Vec3) *Vec3 {
	return &Vec3{v.X - a.X, v.Y - a.Y, v.Z - a.Z}
}

func Dot(a, b *Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (v *Vec3) Length2() float64 {
	return Dot(v, v)
}

func (v *Vec3) Length() float64 {
	return math.Sqrt(v.Length2())
}

func (v *Vec3) Normalize() *Vec3 {
	length := v.Length()
	if length == 0.0 {
		return &Vec3{0.0, 0.0, 0.0}
	}
	return v.Mul(1.0 / length)
}

func (v *Vec3) NearZero() bool {
	return Abs(v.X) < eps && Abs(v.Y) < eps && Abs(v.Z) < eps
}

func Cross(u, v *Vec3) *Vec3 {
	return &Vec3{u.Y*v.Z - u.Z*v.Y,
		u.Z*v.X - u.X*v.Z,
		u.X*v.Y - u.Y*v.X}
}
