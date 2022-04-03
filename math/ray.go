package math

type Ray struct {
	Origin    Point3
	Direction Vec3
}

func GetRay(origin, target *Point3) *Ray {
	dir := GetDirection(origin, target).Normalize()
	return &Ray{*origin, *dir}
}

func (r *Ray) At(t float64) *Point3 {
	return r.Origin.Move(r.Direction.Mul(t))
}
