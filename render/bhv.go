package render

import (
	. "github.com/alexa-infra/rayme/math"
	"sort"
)

type Bvh struct {
	left, right Hittable
	box         *Aabb
}

func (this *Bvh) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	if !this.box.Hit(ray, tMin, tMax) {
		return false, nil
	}
	hitLeft, r1 := this.left.hit(ray, tMin, tMax)
	if hitLeft {
		tMax = r1.t
	}
	hitRight, r2 := this.right.hit(ray, tMin, tMax)
	if hitRight {
		return true, r2
	}
	return hitLeft, r1
}

func (this *Bvh) boundingBox(t0, t1 float64) (bool, *Aabb) {
	return true, this.box
}

func MakeBvh(objects []Hittable, t0, t1 float64) *Bvh {
	if len(objects) == 0 {
		return nil
	}
	if len(objects) == 1 {
		first := objects[0]
		_, box := first.boundingBox(t0, t1)
		return &Bvh{first, first, box}
	}
	axis := RandomInt(3)
	if axis == 0 {
		sort.Slice(objects, func(i, j int) bool {
			_, box1 := objects[i].boundingBox(t0, t1)
			_, box2 := objects[j].boundingBox(t0, t1)
			return box1.Min.X < box2.Min.X
		})
	} else if axis == 1 {
		sort.Slice(objects, func(i, j int) bool {
			_, box1 := objects[i].boundingBox(t0, t1)
			_, box2 := objects[j].boundingBox(t0, t1)
			return box1.Min.Y < box2.Min.Y
		})
	} else {
		sort.Slice(objects, func(i, j int) bool {
			_, box1 := objects[i].boundingBox(t0, t1)
			_, box2 := objects[j].boundingBox(t0, t1)
			return box1.Min.Z < box2.Min.Z
		})
	}
	mid := len(objects) / 2
	left := MakeBvh(objects[:mid], t0, t1)
	right := MakeBvh(objects[mid:], t0, t1)
	_, box1 := left.boundingBox(t0, t1)
	_, box2 := right.boundingBox(t0, t1)
	box := SurroundingBox(box1, box2)
	return &Bvh{left, right, box}
}

func MakeBvhFromList(list *HittableList, t0, t1 float64) *Bvh {
	return MakeBvh(list.Objects, t0, t1)
}
