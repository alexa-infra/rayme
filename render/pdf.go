package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type Pdf interface {
	value(direction *Vec3) float64
	generate(rng *RandExt) *Vec3
}

type MixturePdf struct {
	a, b Pdf
}

func (this *MixturePdf) value(direction *Vec3) float64 {
	return 0.5 * this.a.value(direction) + 0.5 * this.b.value(direction)
}

func (this *MixturePdf) generate(rng *RandExt) *Vec3 {
	if (rng.Float64() < 0.5) {
		return this.a.generate(rng)
	}
	return this.b.generate(rng)
}

func MakeMixturePdf(a, b Pdf) *MixturePdf {
	return &MixturePdf{ a, b }
}

type CosinePdf struct {
	uvw *Onb
}

func (this *CosinePdf) value(direction *Vec3) float64 {
	cosine := Dot(direction.Normalize(), this.uvw.W)
	if cosine <= 0 {
		return 0
	}
	return cosine / math.Pi
}

func (this *CosinePdf) generate(rng *RandExt) *Vec3 {
	return this.uvw.Local(rng.RandomCosineDirection())
}

func MakeCosinePdf(w *Vec3) *CosinePdf {
	return &CosinePdf{ BuildOnbFromW(w) }
}

type HittablePdf struct {
	obj Hittable
	origin *Point3
}

func (this *HittablePdf) value(direction *Vec3) float64 {
	return this.obj.pdfValue(this.origin, direction)
}

func (this *HittablePdf) generate(rng *RandExt) *Vec3 {
	return this.obj.random(this.origin, rng)
}

func MakeHittablePdf(obj Hittable, origin *Point3) *HittablePdf {
	return &HittablePdf{ obj, origin }
}
