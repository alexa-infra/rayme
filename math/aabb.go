package math

type Aabb struct {
	Min, Max *Point3
}

func (self *Aabb) Hit(ray *Ray, tMin, tMax float64) bool {
	x1 := (self.Min.X - ray.Origin.X) / ray.Direction.X
	x2 := (self.Max.X - ray.Origin.X) / ray.Direction.X
	xMin := min(x1, x2)
	xMax := max(x1, x2)
	tMin = max(xMin, tMin)
	tMax = min(xMax, tMax)
	if tMax <= tMin {
		return false
	}
	y1 := (self.Min.Y - ray.Origin.Y) / ray.Direction.Y
	y2 := (self.Max.Y - ray.Origin.Y) / ray.Direction.Y
	yMin := min(y1, y2)
	yMax := max(y1, y2)
	tMin = max(yMin, tMin)
	tMax = min(yMax, tMax)
	if tMax <= tMin {
		return false
	}
	z1 := (self.Min.Z - ray.Origin.Z) / ray.Direction.Z
	z2 := (self.Max.Z - ray.Origin.Z) / ray.Direction.Z
	zMin := min(z1, z2)
	zMax := max(z1, z2)
	tMin = max(zMin, tMin)
	tMax = min(zMax, tMax)
	if tMax <= tMin {
		return false
	}
	return true
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a < b {
		return b
	}
	return a
}

func SurroundingBox(box0, box1 *Aabb) *Aabb {
	small := &Point3{min(box0.Min.X, box1.Min.X), min(box0.Min.Y, box1.Min.Y), min(box0.Min.Z, box1.Min.Z)}
	big := &Point3{max(box0.Max.X, box1.Max.X), max(box0.Max.Y, box1.Max.Y), max(box0.Max.Z, box1.Max.Z)}
	return &Aabb{small, big}
}
