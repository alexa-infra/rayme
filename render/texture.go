package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type Texture interface {
	GetValue(u, v float64, p *Point3) *Vec3
}

type SolidColor struct {
	color *Vec3
}

func (this *SolidColor) GetValue(u, v float64, p *Point3) *Vec3 {
	return this.color
}

func MakeSolidColor(color *Vec3) *SolidColor {
	return &SolidColor{color}
}

type CheckerTexture struct {
	odd, even *Vec3
}

func MakeCheckerTexture(odd, even *Vec3) *CheckerTexture {
	return &CheckerTexture{odd, even}
}

func (this *CheckerTexture) GetValue(u, v float64, p *Point3) *Vec3 {
	sines := math.Sin(10*p.X) * math.Sin(10*p.Y) * math.Sin(10*p.Z)
	if sines < 0 {
		return this.odd
	}
	return this.even
}

type NoiseTexture struct {
	perlin *Perlin
	scale float64
}

func MakeNoiseTexture(scale float64) *NoiseTexture {
	return &NoiseTexture{MakePerlin(), scale}
}

func (this *NoiseTexture) GetValue(u, v float64, p *Point3) *Vec3 {
	ps := &Point3{p.X * this.scale, p.Y * this.scale, p.Z * this.scale}
	noise := this.perlin.Turb(ps, 7)
	return &Vec3{noise, noise, noise}
}
