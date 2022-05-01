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

func randomUnitVector() *Vec3 {
	return randomInUnitSphere().Normalize()
}

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
	hit, rec := world.hit(r, 0.001, 1000)
	if hit {
		scattered, attenuation, target := (*rec.Material).scatter(r, rec)
		if !scattered {
			return &Vec3{ 0.0, 0.0, 0.0 }
		}
		return getRayColor(target, world, depth - 1).MulVec(attenuation)
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

func MakeCamera(lookFrom, lookAt Point3, vup Vec3, vfov float64, aspectRatio float64) *Camera {
	theta := DegreesToRadians(vfov)
	h := math.Tan(theta / 2.0)
	viewportHeight := 2.0 * h
	viewportWidth := aspectRatio * viewportHeight
	w := GetDirection(&lookAt, &lookFrom).Normalize() // note, negative
	u := Cross(&vup, w).Normalize()
	v := Cross(w, u)
	origin := lookFrom
	horizontal := u.Mul(viewportWidth)
	vertical := v.Mul(viewportHeight)
	lowerLeftCorner := origin.Move(horizontal.Mul(-0.5)).Move(vertical.Mul(-0.5)).Move(w.Mul(-1.0))
	return &Camera{origin, *horizontal, *vertical, *lowerLeftCorner}
}

func (c *Camera) GetRay(u, v float64) *Ray {
	target := c.lowerLeftCorner.Move(c.horizontal.Mul(u)).Move(c.vertical.Mul(v))
	return GetRay(&c.origin, target)
}

type Material interface {
	scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray)
}

type Lambertian struct {
	albedo Vec3
}

func (this *Lambertian) scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray) {
	dir := rec.n.Add(randomUnitVector())
	if dir.NearZero() {
		dir = &rec.n
	}
	scattered := GetRay(&rec.p, rec.p.Move(dir))
	return true, &this.albedo, scattered
}

func reflect(v, n *Vec3) *Vec3 {
	dot := Dot(v, n)
	return v.Add(n.Mul(-2.0 * dot))
}

type Metal struct {
	albedo Vec3
	fuzz   float64
}

func (this *Metal) scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray) {
	reflected := reflect(&r.Direction, &rec.n)
	fuzz := randomInUnitSphere().Mul(this.fuzz)
	scattered := GetRay(&rec.p, rec.p.Move(reflected.Add(fuzz)))
	return Dot(&scattered.Direction, &rec.n) > 0, &this.albedo, scattered
}

func refract(uv *Vec3, n *Vec3, angleFrac float64) *Vec3 {
	cosTheta := Min(Dot(uv.Mul(-1.0), n), 1.0)
	perp := uv.Add(n.Mul(cosTheta)).Mul(angleFrac)
	parallel := n.Mul(-math.Sqrt(1.0 - perp.Length2()))
	return perp.Add(parallel)
}

func reflectance(cosine float64, refIdx float64) float64 {
	r0 := (1.0 - refIdx) / (1.0 + refIdx)
	r0 = r0 * r0
	return r0 + (1 - r0) * math.Pow(1 - cosine, 5)
}

type Dielectric struct {
	ri float64 // Index of refraction
}

func (this *Dielectric) scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray) {
	attenuation := &Vec3{1.0, 1.0, 1.0}
	ratio := this.ri
	if rec.frontFace {
		ratio = 1.0 / this.ri
	}
	unitDirection := r.Direction.Normalize()
	cosTheta := Min(Dot(unitDirection.Mul(-1.0), &rec.n), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)
	cannotRefract := ratio * sinTheta > 1.0
	var dir *Vec3 = nil
	if cannotRefract || reflectance(cosTheta, ratio) > randGen.Float64() {
		dir = reflect(unitDirection, &rec.n)
	} else {
		dir = refract(unitDirection, &rec.n, ratio)
	}
	scattered := GetRay(&rec.p, rec.p.Move(dir))
	return true, attenuation, scattered
}

func main() {
	samples := []Vec2{Vec2{0.0, 0.0}}
	for i := 0; i < samplesPerPixel; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(samplesPerPixel)
		samples = append(samples, Vec2{0.25 * math.Cos(angle), 0.25 * math.Sin(angle)})
	}
	camera := MakeCamera(Point3{-2,2,1}, Point3{0,0,-1}, Vec3{0,1,0}, 20.0, aspectRatio)

	ground := &Lambertian{Vec3{0.8, 0.8, 0.0}}
	center := &Lambertian{Vec3{0.1, 0.2, 0.5}}
	left := &Dielectric{1.5}
	right := &Metal{Vec3{0.8, 0.6, 0.2}, 0.0}

	world := &HittableList{
		[]Hittable{
			&Sphere{Point3{0.0, -100.5, -1.0}, 100.0, ground},
			&Sphere{Point3{0.0, 0.0, -1.0}, 0.5, center},
			&Sphere{Point3{-1.0, 0.0, -1.0}, 0.5, left},
			&Sphere{Point3{-1.0, 0.0, -1.0}, -0.4, left},
			&Sphere{Point3{ 1.0, 0.0, -1.0}, 0.5, right},
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
