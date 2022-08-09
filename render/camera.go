package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type Camera struct {
	origin          *Point3
	horizontal      *Vec3
	vertical        *Vec3
	lowerLeftCorner *Point3
	onb             *Onb
	lensRadius      float64
	t1, t2          float64
}

func MakeCamera(lookFrom, lookAt *Point3, vup *Vec3, vfov float64, aspectRatio float64, aperture float64, focusDist float64, t1, t2 float64) *Camera {
	theta := DegreesToRadians(vfov)
	h := math.Tan(theta / 2.0)
	viewportHeight := 2.0 * h
	viewportWidth := aspectRatio * viewportHeight
	dir := GetDirection(lookAt, lookFrom) // note, negative
	onb := MakeOnbFromDirection(dir, vup)
	origin := lookFrom
	horizontal := onb.U.Mul(viewportWidth * focusDist)
	vertical := onb.V.Mul(viewportHeight * focusDist)
	forward := onb.W.Mul(-focusDist)
	lowerLeftCorner := origin.Move(horizontal.Mul(-0.5)).Move(vertical.Mul(-0.5)).Move(forward)
	lensRadius := aperture / 2.0
	return &Camera{origin, horizontal, vertical, lowerLeftCorner, onb, lensRadius, t1, t2}
}

func (c *Camera) CastRay(s, t float64, rng *RandExt) *Ray {
	var origin *Point3 = c.origin
	if c.lensRadius != 0.0 {
		rd := rng.RandomInUnitDisk().Mul(c.lensRadius)
		offset := c.onb.Local(rd)
		origin = origin.Move(offset)
	}
	target := c.lowerLeftCorner.Move(c.horizontal.Mul(s)).Move(c.vertical.Mul(t))
	return MakeRayFromPoints(origin, target, rng.Between(c.t1, c.t2))
}
