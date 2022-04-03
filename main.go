package main

import (
	"fmt"
	. "github.com/alexa-infra/rayme/math"
	"image"
	"image/png"
	"math"
	"math/rand"
	"os"
)

const (
	aspectRatio     = 16.0 / 9.0
	imageWidth      = 400
	imageHeight     = int(imageWidth / aspectRatio)
	viewportHeight  = 2.0
	viewportWidth   = aspectRatio * viewportHeight
	focalLength     = 1.0
	samplesPerPixel = 12
	maxDepth        = 50
)

var (
	randGen = rand.New(rand.NewSource(99))
)

func randomInUnitSphere() *Vec3 {
	for {
		// [0, 1] -> [-0.5, 0.5] -> [-1, 1]
		x := (randGen.Float64() - 0.5) * 2.0
		y := (randGen.Float64() - 0.5) * 2.0
		z := (randGen.Float64() - 0.5) * 2.0
		v := &Vec3{ x, y, z }
		if v.Length2() >= 1.0 {
			continue
		}
		return v
	}
}

type HitRecord struct {
	t         float64
	p         Point3
	n         Vec3
	frontFace bool
}

func MakeHitRecord(ray *Ray, root float64, point *Point3, normal *Vec3) *HitRecord {
	frontFace := Dot(&ray.Direction, normal) < 0
	n := normal
	if !frontFace {
		n = normal.Mul(-1.0)
	}
	return &HitRecord{root, *point, *n, frontFace}
}

type Hittable interface {
	hit(r *Ray, tMin, tMax float64) (bool, *HitRecord)
}

type Sphere struct {
	Center Point3
	Radius float64
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
	return true, MakeHitRecord(ray, root, hitPoint, normal)
}

type HittableList struct {
	objects []Hittable
}

func (this *HittableList) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	var closest *HitRecord = nil
	for _, object := range this.objects {
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

func getRayColor(r *Ray, world Hittable, depth int) *Vec3 {
	if depth <= 0 {
		return &Vec3{ 0.0, 0.0, 0.0 }
	}
	hit, point := world.hit(r, 0.001, 1000)
	if hit {
		target := point.p.Move(&point.n).Move(randomInUnitSphere())
		return getRayColor(GetRay(&point.p, target), world, depth - 1).Mul(0.5)
	}
	t := 0.5 * (r.Direction.Y + 1.0)
	color1 := Vec3{1.0, 1.0, 1.0}
	color2 := Vec3{0.5, 0.7, 1.0}
	return color1.Mul(1.0 - t).Add(color2.Mul(t))
}

type Camera struct {
	origin          Point3
	horizontal      Vec3
	vertical        Vec3
	lowerLeftCorner Point3
}

func MakeCamera() *Camera {
	origin := Point3{0.0, 0.0, 0.0}
	horizontal := Vec3{viewportWidth, 0.0, 0.0}
	vertical := Vec3{0.0, viewportHeight, 0.0}
	lowerLeftCorner := origin.Move(horizontal.Mul(-0.5)).Move(vertical.Mul(-0.5)).Move(&Vec3{0.0, 0.0, -focalLength})
	return &Camera{origin, horizontal, vertical, *lowerLeftCorner}
}

func (c *Camera) GetRay(u, v float64) *Ray {
	target := c.lowerLeftCorner.Move(c.horizontal.Mul(u)).Move(c.vertical.Mul(v))
	return GetRay(&c.origin, target)
}

func main() {
	samples := []Vec2{Vec2{0.0, 0.0}}
	for i := 0; i < samplesPerPixel; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(samplesPerPixel)
		samples = append(samples, Vec2{0.25 * math.Cos(angle), 0.25 * math.Sin(angle)})
	}
	camera := MakeCamera()

	world := &HittableList{
		[]Hittable{
			&Sphere{Point3{0.0, 0.0, -1.0}, 0.5},
			&Sphere{Point3{0.0, -100.5, -1.0}, 100.0},
		},
	}

	myImg := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	for j := 0; j < imageHeight; j++ {
		for i := 0; i < imageWidth; i++ {
			sumColor := &Vec3{0, 0, 0}
			for _, s := range samples {
				u := (float64(i) + s.X) / float64(imageWidth-1)
				v := (float64(j) + s.Y) / float64(imageHeight-1)
				ray := camera.GetRay(u, v)
				rayColor := getRayColor(ray, world, maxDepth)
				sumColor = sumColor.Add(rayColor)
			}
			scale := 1.0 / float64(len(samples))
			sumColor = &Vec3{
				math.Sqrt(sumColor.X * scale),
				math.Sqrt(sumColor.Y * scale),
				math.Sqrt(sumColor.Z * scale),
			}
			myImg.SetRGBA(i, imageHeight - j - 1, sumColor.AsColor())
		}
	}
	out, err := os.Create("output.png")
	if err != nil {
		fmt.Println("can't open file to write")
		os.Exit(1)
	}
	png.Encode(out, myImg)
	out.Close()
}
