package math

type Point3 struct {
	*Vec3
}

func MakePoint3(x, y, z float64) *Point3 {
	return &Point3{&Vec3{x, y, z}}
}

func (p *Point3) Move(v *Vec3) *Point3 {
	return MakePoint3(p.X+v.X, p.Y+v.Y, p.Z+v.Z)
}

func GetDirection(a, b *Point3) *Vec3 {
	return &Vec3{b.X - a.X, b.Y - a.Y, b.Z - a.Z}
}

func Distance(a, b *Point3) float64 {
	return GetDirection(a, b).Length()
}
