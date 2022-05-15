package math

type Ray struct {
	Origin    Point3
	Direction Vec3
	Time	  float64
}

func MakeRayFromPoints(origin, target *Point3, time float64) *Ray {
	dir := GetDirection(origin, target).Normalize()
	return &Ray{*origin, *dir, time}
}

func MakeRayFromDirection(origin *Point3, dir *Vec3, time float64) *Ray {
	return &Ray{*origin, *dir.Normalize(), time}
}

func (r *Ray) At(t float64) *Point3 {
	return r.Origin.Move(r.Direction.Mul(t))
}
