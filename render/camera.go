package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type Camera struct {
	origin          Point3
	horizontal      Vec3
	vertical        Vec3
	lowerLeftCorner Point3
	u, v, w         Vec3
	lensRadius      float64
}

func MakeCamera(lookFrom, lookAt Point3, vup Vec3, vfov float64, aspectRatio float64, aperture float64, focusDist float64) *Camera {
	theta := DegreesToRadians(vfov)
	h := math.Tan(theta / 2.0)
	viewportHeight := 2.0 * h
	viewportWidth := aspectRatio * viewportHeight
	w := GetDirection(&lookAt, &lookFrom).Normalize() // note, negative
	u := Cross(&vup, w).Normalize()
	v := Cross(w, u)
	origin := lookFrom
	horizontal := u.Mul(viewportWidth).Mul(focusDist)
	vertical := v.Mul(viewportHeight).Mul(focusDist)
	lowerLeftCorner := origin.Move(horizontal.Mul(-0.5)).Move(vertical.Mul(-0.5)).Move(w.Mul(-1.0).Mul(focusDist))
	lensRadius := aperture / 2.0
	return &Camera{origin, *horizontal, *vertical, *lowerLeftCorner, *u, *v, *w, lensRadius}
}

func (c *Camera) GetRay(s, t float64) *Ray {
	rd := RandomInUnitDisk().Mul(c.lensRadius);
	offset := c.u.Mul(rd.X).Add(c.v.Mul(rd.Y));
	target := c.lowerLeftCorner.Move(c.horizontal.Mul(s)).Move(c.vertical.Mul(t))
	return GetRay(c.origin.Move(offset), target)
}
