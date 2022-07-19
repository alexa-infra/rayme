package math

type Onb struct {
	U, V, W *Vec3
}

func (this *Onb) localVec(a, b, c float64) *Vec3 {
	vx := this.U.Mul(a)
	vy := this.V.Mul(b)
	vz := this.W.Mul(c)
	return vx.Add(vy).Add(vz)
}

func (this *Onb) Local(v *Vec3) *Vec3 {
	return this.localVec(v.X, v.Y, v.Z)
}

func BuildOnbFromW(n *Vec3) *Onb {
	var u, v, w, a *Vec3
	w = n.Normalize()
	if Abs(w.X) > 0.9 {
		a = &Vec3{0, 1, 0}
	} else {
		a = &Vec3{1, 0, 0}
	}
	v = Cross(w, a).Normalize()
	u = Cross(w, v)
	return &Onb{u, v, w}
}
