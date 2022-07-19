package render

import (
	. "github.com/alexa-infra/rayme/math"
	"image"
	_ "image/jpeg"
	"math"
	"os"
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

type CheckerTexture3d struct {
	scale     float64
	odd, even *Vec3
}

func MakeCheckerTexture3d(scale float64, odd, even *Vec3) *CheckerTexture3d {
	return &CheckerTexture3d{scale, odd, even}
}

func (this *CheckerTexture3d) GetValue(u, v float64, p *Point3) *Vec3 {
	a := math.Floor(p.X * this.scale)
	b := math.Floor(p.Y * this.scale)
	c := math.Floor(p.Z * this.scale)
	if math.Mod(math.Abs(a+b+c), 2.0) > 0.5 {
		return this.odd
	}
	return this.even
}

type CheckerTexture2d struct {
	scale     float64
	odd, even *Vec3
}

func MakeCheckerTexture2d(scale float64, odd, even *Vec3) *CheckerTexture2d {
	return &CheckerTexture2d{scale, odd, even}
}

func (this *CheckerTexture2d) GetValue(u, v float64, p *Point3) *Vec3 {
	a := math.Floor(u * this.scale)
	b := math.Floor(v * this.scale)
	if math.Mod(math.Abs(a+b), 2.0) > 0.5 {
		return this.odd
	}
	return this.even
}

type NoiseTexture struct {
	perlin *Perlin
	scale  float64
}

func MakeNoiseTexture(scale float64, r *RandExt) *NoiseTexture {
	return &NoiseTexture{MakePerlin(r), scale}
}

func (this *NoiseTexture) GetValue(u, v float64, p *Point3) *Vec3 {
	ps := &Point3{p.X * this.scale, p.Y * this.scale, p.Z * this.scale}
	noise := this.perlin.Turb(ps, 7)
	return &Vec3{noise, noise, noise}
}

type ImageTexture struct {
	img image.Image
}

func MakeImageTexture(path string) (*ImageTexture, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return &ImageTexture{img}, nil
}

func (this *ImageTexture) GetValue(u, v float64, p *Point3) *Vec3 {
	u = Clamp(u, 0.0, 1.0)
	v = 1.0 - Clamp(v, 0.0, 1.0)
	size := this.img.Bounds().Size()
	x := int(u * float64(size.X))
	y := int(v * float64(size.Y))
	r, g, b, _ := this.img.At(x, y).RGBA()
	scale := 1.0 / float64(0xffff)
	return &Vec3{float64(r) * scale, float64(g) * scale, float64(b) * scale}
}
