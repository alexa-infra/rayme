package math

import "math"

type Aabb struct {
	Min, Max *Point3
}

func (self *Aabb) Hit(ray *Ray, tMin, tMax float64) bool {
	x1 := (self.Min.X - ray.Origin.X) / ray.Direction.X
	x2 := (self.Max.X - ray.Origin.X) / ray.Direction.X
	xMin := Min(x1, x2)
	xMax := Max(x1, x2)
	tMin = Max(xMin, tMin)
	tMax = Min(xMax, tMax)
	if tMax <= tMin {
		return false
	}
	y1 := (self.Min.Y - ray.Origin.Y) / ray.Direction.Y
	y2 := (self.Max.Y - ray.Origin.Y) / ray.Direction.Y
	yMin := Min(y1, y2)
	yMax := Max(y1, y2)
	tMin = Max(yMin, tMin)
	tMax = Min(yMax, tMax)
	if tMax <= tMin {
		return false
	}
	z1 := (self.Min.Z - ray.Origin.Z) / ray.Direction.Z
	z2 := (self.Max.Z - ray.Origin.Z) / ray.Direction.Z
	zMin := Min(z1, z2)
	zMax := Max(z1, z2)
	tMin = Max(zMin, tMin)
	tMax = Min(zMax, tMax)
	if tMax <= tMin {
		return false
	}
	return true
}

func SurroundingBox(box0, box1 *Aabb) *Aabb {
	small := MakePoint3(Min(box0.Min.X, box1.Min.X), Min(box0.Min.Y, box1.Min.Y), Min(box0.Min.Z, box1.Min.Z))
	big := MakePoint3(Max(box0.Max.X, box1.Max.X), Max(box0.Max.Y, box1.Max.Y), Max(box0.Max.Z, box1.Max.Z))
	return &Aabb{small, big}
}

type AabbBuilder struct {
	min, max *Point3
	n        int
}

func MakeAabbBuilder() *AabbBuilder {
	inf := math.Inf(1)
	ninf := math.Inf(-1)
	min := MakePoint3(inf, inf, inf)
	max := MakePoint3(ninf, ninf, ninf)
	return &AabbBuilder{min, max, 0}
}

func (this *AabbBuilder) AddPoint(p *Point3) {
	this.min.X = Min(this.min.X, p.X)
	this.max.X = Max(this.max.X, p.X)
	this.min.Y = Min(this.min.Y, p.Y)
	this.max.Y = Max(this.max.Y, p.Y)
	this.min.Z = Min(this.min.Z, p.Z)
	this.max.Z = Max(this.max.Z, p.Z)
	this.n++
}

func (this *AabbBuilder) GetBox() *Aabb {
	if this.n == 0 {
		p := MakePoint3(0, 0, 0)
		return &Aabb{p, p}
	}
	return &Aabb{this.min, this.max}
}
