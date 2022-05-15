package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type HitRecord struct {
	t         float64
	p         Point3
	n         Vec3
	frontFace bool
	Material  *Material
}

func MakeHitRecord(ray *Ray, root float64, point *Point3, normal *Vec3, material *Material) *HitRecord {
	frontFace := Dot(&ray.Direction, normal) < 0
	if !frontFace {
		normal = normal.Mul(-1.0)
	}
	return &HitRecord{root, *point, *normal, frontFace, material}
}

type Hittable interface {
	hit(r *Ray, tMin, tMax float64) (bool, *HitRecord)
}

type Sphere struct {
	Center   Point3
	Radius   float64
	Material
}

func (this *Sphere) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	oc := GetDirection(&this.Center, &ray.Origin)
	a := Dot(&ray.Direction, &ray.Direction)
	h := Dot(oc, &ray.Direction)
	c := Dot(oc, oc) - this.Radius*this.Radius
	discriminant := h*h - a*c
	if discriminant < 0 {
		return false, nil
	}
	root1 := (-h - math.Sqrt(discriminant)) / a
	root2 := (-h + math.Sqrt(discriminant)) / a
	root := 0.0
	if root1 < tMin || root1 > tMax {
		if root2 >= tMin && root2 <= tMax {
			root = root2
		} else {
			return false, nil
		}
	} else {
		if root2 >= tMin && root2 <= tMax {
			root = Min(root1, root2)
		} else {
			root = root1
		}
	}
	hitPoint := ray.At(root)
	normal := GetDirection(&this.Center, hitPoint).Mul(1.0 / this.Radius)
	return true, MakeHitRecord(ray, root, hitPoint, normal, &this.Material)
}

type MovingSphere struct {
	Center0, Center1 Point3
	Radius           float64
	Time0, Time1     float64
	Material
}

func (this *MovingSphere) center(time float64) *Point3 {
	scale := (time - this.Time0) / (this.Time1 - this.Time0)
	dir := GetDirection(&this.Center0, &this.Center1)
	return this.Center0.Move(dir.Mul(scale))
}

func (this *MovingSphere) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	oc := GetDirection(this.center(ray.Time), &ray.Origin)
	a := Dot(&ray.Direction, &ray.Direction)
	h := Dot(oc, &ray.Direction)
	c := Dot(oc, oc) - this.Radius*this.Radius
	discriminant := h*h - a*c
	if discriminant < 0 {
		return false, nil
	}
	root1 := (-h - math.Sqrt(discriminant)) / a
	root2 := (-h + math.Sqrt(discriminant)) / a
	root := 0.0
	if root1 < tMin || root1 > tMax {
		if root2 >= tMin && root2 <= tMax {
			root = root2
		} else {
			return false, nil
		}
	} else {
		if root2 >= tMin && root2 <= tMax {
			root = Min(root1, root2)
		} else {
			root = root1
		}
	}
	hitPoint := ray.At(root)
	normal := GetDirection(this.center(ray.Time), hitPoint).Mul(1.0 / this.Radius)
	return true, MakeHitRecord(ray, root, hitPoint, normal, &this.Material)
}

type HittableList struct {
	Objects []Hittable
}

func (this *HittableList) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	var closest *HitRecord = nil
	for _, object := range this.Objects {
		hit, rec := object.hit(ray, tMin, tMax)
		if !hit {
			continue
		}
		if closest == nil || rec.t < closest.t {
			closest = rec
		}
	}
	return closest != nil, closest
}

func GetRayColor(r *Ray, world Hittable, depth int) *Vec3 {
	if depth <= 0 {
		return &Vec3{ 0.0, 0.0, 0.0 }
	}
	hit, rec := world.hit(r, 0.001, 1000)
	if hit {
		scattered, attenuation, target := (*rec.Material).Scatter(r, rec)
		if !scattered {
			return &Vec3{ 0.0, 0.0, 0.0 }
		}
		return GetRayColor(target, world, depth - 1).MulVec(attenuation)
	}
	t := 0.5 * (r.Direction.Y + 1.0)
	color1 := Vec3{1.0, 1.0, 1.0}
	color2 := Vec3{0.5, 0.7, 1.0}
	return color1.Mul(1.0 - t).Add(color2.Mul(t))
}
