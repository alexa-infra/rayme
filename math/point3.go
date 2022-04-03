package math

type Point3 struct {
	X, Y, Z float64
}

func (p *Point3) Move(v *Vec3) *Point3 {
	return &Point3{p.X + v.X, p.Y + v.Y, p.Z + v.Z}
}

func GetDirection(a, b *Point3) *Vec3 {
	return &Vec3{b.X - a.X, b.Y - a.Y, b.Z - a.Z}
}
