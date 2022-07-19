package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

var noColor *Vec3 = &Vec3{0, 0, 0}

type ScatterRecord struct {
	specular *Ray
	isSpecular bool
	attenuation *Vec3
	pdf float64
}

type Material interface {
	Scatter(r *Ray, rec *HitRecord, rng *RandExt) (bool, *ScatterRecord)
	Emitted(u, v float64, p *Point3) *Vec3
	ScatteringPDF(r *Ray, rec *HitRecord, scattered *Ray) float64
}

type materialNoPdf struct {}

func (this *materialNoPdf) ScatteringPDF(r *Ray, rec *HitRecord, scattered *Ray) float64 {
	return 0.0
}

type Lambertian struct {
	albedo Texture
}

func MakeLambertianSolidColor(albedo *Vec3) *Lambertian {
	t := MakeSolidColor(albedo)
	return &Lambertian{t}
}

func MakeLambertianTexture(t Texture) *Lambertian {
	return &Lambertian{t}
}

func (this *Lambertian) Scatter(r *Ray, rec *HitRecord, rng *RandExt) (bool, *ScatterRecord) {
	onb := BuildOnbFromW(rec.n)
	dir := onb.Local(rng.RandomCosineDirection())
	scattered := MakeRayFromDirection(rec.p, dir, r.Time)
	attenuation := this.albedo.GetValue(rec.u, rec.v, rec.p)
	pdf := Dot(onb.W, scattered.Direction) / math.Pi
	return true, &ScatterRecord{ scattered, false, attenuation, pdf }
}

func (this *Lambertian) ScatteringPDF(r *Ray, rec *HitRecord, scattered *Ray) float64 {
	cosine := Dot(rec.n, scattered.Direction)
	if cosine < 0 {
		return 0
	}
	return cosine / math.Pi
}

func (this *Lambertian) Emitted(u, v float64, p *Point3) *Vec3 {
	return noColor
}

func reflect(v, n *Vec3) *Vec3 {
	dot := Dot(v, n)
	return v.Add(n.Mul(-2.0 * dot))
}

type Metal struct {
	albedo *Vec3
	fuzz   float64
	materialNoPdf
}

func MakeMetal(albedo *Vec3, fuzz float64) *Metal {
	return &Metal{albedo, fuzz, materialNoPdf{}}
}

func (this *Metal) Scatter(r *Ray, rec *HitRecord, rng *RandExt) (bool, *ScatterRecord) {
	fuzz := rng.RandomInUnitSphere().Mul(this.fuzz)
	reflected := reflect(r.Direction, rec.n).Add(fuzz)
	scattered := MakeRayFromDirection(rec.p, reflected, r.Time)
	return Dot(scattered.Direction, rec.n) > 0, &ScatterRecord{ scattered, true, this.albedo, 0.0 }
}

func (this *Metal) Emitted(u, v float64, p *Point3) *Vec3 {
	return noColor
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
	return r0 + (1-r0)*math.Pow(1-cosine, 5)
}

type Dielectric struct {
	ri float64 // Index of refraction
	materialNoPdf
}

func MakeDielectric(ri float64) *Dielectric {
	return &Dielectric{ri, materialNoPdf{}}
}

func (this *Dielectric) Scatter(r *Ray, rec *HitRecord, rng *RandExt) (bool, *ScatterRecord) {
	attenuation := &Vec3{1.0, 1.0, 1.0}
	ratio := this.ri
	if rec.frontFace {
		ratio = 1.0 / this.ri
	}
	unitDirection := r.Direction.Normalize()
	cosTheta := Min(Dot(unitDirection.Mul(-1.0), rec.n), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)
	cannotRefract := ratio*sinTheta > 1.0
	var dir *Vec3 = nil
	if cannotRefract || reflectance(cosTheta, ratio) > rng.Between(0.0, 1.0) {
		dir = reflect(unitDirection, rec.n)
	} else {
		dir = refract(unitDirection, rec.n, ratio)
	}
	scattered := MakeRayFromDirection(rec.p, dir, r.Time)
	return true, &ScatterRecord{ scattered, true, attenuation, 0.0 }
}

func (this *Dielectric) Emitted(u, v float64, p *Point3) *Vec3 {
	return noColor
}

type DiffuseLight struct {
	emit Texture
	materialNoPdf
}

func MakeDiffuseLightFromTexture(emit Texture) *DiffuseLight {
	return &DiffuseLight{emit, materialNoPdf{}}
}

func MakeDiffuseLightFromColor(c *Vec3) *DiffuseLight {
	tex := MakeSolidColor(c)
	return &DiffuseLight{tex, materialNoPdf{}}
}

func (this *DiffuseLight) Scatter(r *Ray, rec *HitRecord, rng *RandExt) (bool, *ScatterRecord) {
	return false, &ScatterRecord{ nil, false, nil, 0.0 }
}

func (this *DiffuseLight) Emitted(u, v float64, p *Point3) *Vec3 {
	return this.emit.GetValue(u, v, p)
}
