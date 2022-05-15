package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type Material interface {
	Scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray)
}

type Lambertian struct {
	albedo Vec3
}

func MakeLambertian(albedo Vec3) *Lambertian {
	return &Lambertian{ albedo }
}

func (this *Lambertian) Scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray) {
	dir := rec.n.Add(RandomUnitVector())
	if dir.NearZero() {
		dir = &rec.n
	}
	scattered := MakeRayFromDirection(&rec.p, dir, r.Time)
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

func MakeMetal(albedo Vec3, fuzz float64) *Metal {
	return &Metal{ albedo, fuzz }
}

func (this *Metal) Scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray) {
	fuzz := RandomInUnitSphere().Mul(this.fuzz)
	reflected := reflect(&r.Direction, &rec.n).Add(fuzz)
	scattered := MakeRayFromDirection(&rec.p, reflected, r.Time)
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

func MakeDielectric(ri float64) *Dielectric {
	return &Dielectric{ ri }
}

func (this *Dielectric) Scatter(r *Ray, rec *HitRecord) (bool, *Vec3, *Ray) {
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
	if cannotRefract || reflectance(cosTheta, ratio) > RandGen.Float64() {
		dir = reflect(unitDirection, &rec.n)
	} else {
		dir = refract(unitDirection, &rec.n, ratio)
	}
	scattered := MakeRayFromDirection(&rec.p, dir, r.Time)
	return true, attenuation, scattered
}
