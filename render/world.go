package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type HitRecord struct {
	t         float64
	p         *Point3
	n         *Vec3
	frontFace bool
	Material
	u, v float64
}

func MakeHitRecord(ray *Ray, root float64, point *Point3, normal *Vec3, material Material, u, v float64) *HitRecord {
	frontFace := Dot(ray.Direction, normal) < 0
	if !frontFace {
		normal = normal.Mul(-1.0)
	}
	return &HitRecord{root, point, normal, frontFace, material, u, v}
}

type Hittable interface {
	hit(r *Ray, tMin, tMax float64) (bool, *HitRecord)
	boundingBox(t0, t1 float64) (bool, *Aabb)
}

type Sphere struct {
	Center *Point3
	Radius float64
	Material
}

func getSphereUv(p *Vec3) (u, v float64) {
	theta := math.Acos(-p.Y)
	phi := math.Atan2(-p.Z, p.X) + math.Pi
	u = phi / (2.0 * math.Pi)
	v = theta / math.Pi
	return
}

func (this *Sphere) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	oc := GetDirection(this.Center, ray.Origin)
	a := Dot(ray.Direction, ray.Direction)
	h := Dot(oc, ray.Direction)
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
	normal := GetDirection(this.Center, hitPoint).Mul(1.0 / this.Radius)
	u, v := getSphereUv(normal)
	return true, MakeHitRecord(ray, root, hitPoint, normal, this.Material, u, v)
}

func (this *Sphere) boundingBox(t0, t1 float64) (bool, *Aabb) {
	r := this.Radius * math.Sqrt2
	radius := &Vec3{r, r, r}
	aabb := &Aabb{
		this.Center.Move(radius.Mul(-1)),
		this.Center.Move(radius),
	}
	return true, aabb
}

type MovingSphere struct {
	Center0, Center1 *Point3
	Radius           float64
	Time0, Time1     float64
	Material
}

func (this *MovingSphere) center(time float64) *Point3 {
	scale := (time - this.Time0) / (this.Time1 - this.Time0)
	dir := GetDirection(this.Center0, this.Center1)
	return this.Center0.Move(dir.Mul(scale))
}

func (this *MovingSphere) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	c := this.center(ray.Time)
	sphere := Sphere{c, this.Radius, this.Material}
	return sphere.hit(ray, tMin, tMax)
}

func (this *MovingSphere) boundingBox(t0, t1 float64) (bool, *Aabb) {
	center0 := this.center(t0)
	sphere0 := Sphere{center0, this.Radius, this.Material}
	_, aabb0 := sphere0.boundingBox(t0, t1)

	center1 := this.center(t1)
	sphere1 := Sphere{center1, this.Radius, this.Material}
	_, aabb1 := sphere1.boundingBox(t0, t1)
	return true, SurroundingBox(aabb0, aabb1)
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

func (this *HittableList) boundingBox(t0, t1 float64) (bool, *Aabb) {
	found := false
	var surrounding *Aabb
	for _, obj := range this.Objects {
		ok, aabb := obj.boundingBox(t0, t1)
		if ok {
			if !found {
				surrounding = aabb
				found = true
			} else {
				surrounding = SurroundingBox(surrounding, aabb)
			}
		}
	}
	return found, surrounding
}

func GetRayColor(r *Ray, bgColor *Vec3, world Hittable, depth int) *Vec3 {
	if depth <= 0 {
		return noColor
	}
	hit, rec := world.hit(r, 0.001, 10000)
	if !hit {
		return bgColor
	}

	emitted := rec.Material.Emitted(rec.u, rec.v, rec.p)
	scattered, attenuation, target := rec.Material.Scatter(r, rec)
	if !scattered {
		return emitted
	}
	return GetRayColor(target, bgColor, world, depth-1).MulVec(attenuation).Add(emitted)
}

type RectXY struct {
	x0, y0 float64
	x1, y1 float64
	k      float64
	Material
}

func MakeRectXY(x0, y0, x1, y1, k float64, m Material) *RectXY {
	return &RectXY{x0, y0, x1, y1, k, m}
}

func (this *RectXY) boundingBox(t0, t1 float64) (bool, *Aabb) {
	a := &Point3{this.x0, this.y0, this.k - 0.0001}
	b := &Point3{this.x1, this.y1, this.k + 0.0001}
	return true, &Aabb{a, b}
}

func (this *RectXY) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	t := (this.k - ray.Origin.Z) / ray.Direction.Z
	if t < tMin || t > tMax {
		return false, nil
	}
	x := ray.Origin.X + t*ray.Direction.X
	y := ray.Origin.Y + t*ray.Direction.Y
	if x < this.x0 || x > this.x1 || y < this.y0 || y > this.y1 {
		return false, nil
	}
	hitPoint := ray.At(t)
	u := (x - this.x0) / (this.x1 - this.x0)
	v := (y - this.y0) / (this.y1 - this.y0)
	normal := &Vec3{0, 0, 1}
	return true, MakeHitRecord(ray, t, hitPoint, normal, this.Material, u, v)
}

type RectXZ struct {
	x0, z0 float64
	x1, z1 float64
	k      float64
	Material
}

func MakeRectXZ(x0, z0, x1, z1, k float64, m Material) *RectXZ {
	return &RectXZ{x0, z0, x1, z1, k, m}
}

func (this *RectXZ) boundingBox(t0, t1 float64) (bool, *Aabb) {
	a := &Point3{this.x0, this.z0, this.k - 0.0001}
	b := &Point3{this.x1, this.z1, this.k + 0.0001}
	return true, &Aabb{a, b}
}

func (this *RectXZ) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	t := (this.k - ray.Origin.Y) / ray.Direction.Y
	if t < tMin || t > tMax {
		return false, nil
	}
	x := ray.Origin.X + t*ray.Direction.X
	z := ray.Origin.Z + t*ray.Direction.Z
	if x < this.x0 || x > this.x1 || z < this.z0 || z > this.z1 {
		return false, nil
	}
	hitPoint := ray.At(t)
	u := (x - this.x0) / (this.x1 - this.x0)
	v := (z - this.z0) / (this.z1 - this.z0)
	normal := &Vec3{0, 1, 0}
	return true, MakeHitRecord(ray, t, hitPoint, normal, this.Material, u, v)
}

type RectYZ struct {
	y0, z0 float64
	y1, z1 float64
	k      float64
	Material
}

func MakeRectYZ(y0, z0, y1, z1, k float64, m Material) *RectYZ {
	return &RectYZ{y0, z0, y1, z1, k, m}
}

func (this *RectYZ) boundingBox(t0, t1 float64) (bool, *Aabb) {
	a := &Point3{this.y0, this.z0, this.k - 0.0001}
	b := &Point3{this.y1, this.z1, this.k + 0.0001}
	return true, &Aabb{a, b}
}

func (this *RectYZ) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	t := (this.k - ray.Origin.X) / ray.Direction.X
	if t < tMin || t > tMax {
		return false, nil
	}
	y := ray.Origin.Y + t*ray.Direction.Y
	z := ray.Origin.Z + t*ray.Direction.Z
	if y < this.y0 || y > this.y1 || z < this.z0 || z > this.z1 {
		return false, nil
	}
	hitPoint := ray.At(t)
	u := (y - this.y0) / (this.y1 - this.y0)
	v := (z - this.z0) / (this.z1 - this.z0)
	normal := &Vec3{1, 0, 0}
	return true, MakeHitRecord(ray, t, hitPoint, normal, this.Material, u, v)
}
