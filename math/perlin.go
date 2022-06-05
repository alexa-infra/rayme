package math

import "math"

const (
	pointCount = 256
	mask       = 255
)

type Perlin struct {
	randVec             []*Vec3
	permX, permY, permZ []int
}

func MakePerlin() *Perlin {
	f := make([]*Vec3, pointCount)
	for i := 0; i < pointCount; i++ {
		f[i] = RandomInUnitSphere()
	}
	mutex.Lock()
	defer mutex.Unlock()
	x := randGen.Perm(pointCount)
	y := randGen.Perm(pointCount)
	z := randGen.Perm(pointCount)
	return &Perlin{f, x, y, z}
}

func (this *Perlin) Noise(p *Point3) float64 {
	u := p.X - math.Floor(p.X)
	v := p.Y - math.Floor(p.Y)
	w := p.Z - math.Floor(p.Z)
	i := int(math.Floor(p.X))
	j := int(math.Floor(p.Y))
	k := int(math.Floor(p.Z))
	c := make([][][]*Vec3, 2)
	for di := 0; di < 2; di++ {
		c[di] = make([][]*Vec3, 2)
		for dj := 0; dj < 2; dj++ {
			c[di][dj] = make([]*Vec3, 2)
			for dk := 0; dk < 2; dk++ {
				c[di][dj][dk] = this.randVec[this.permX[(i+di)&mask]^this.permY[(j+dj)&mask]^this.permZ[(k+dk)&mask]]
			}
		}
	}
	return trilinearInterp(c, u, v, w)
}

func trilinearInterp(c [][][]*Vec3, u, v, w float64) float64 {
	u = u * u * (3 - 2*u)
	v = v * v * (3 - 2*v)
	w = w * w * (3 - 2*w)
	acc := 0.0
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				weight := &Vec3{u - float64(i), v - float64(j), w - float64(k)}
				acc += (float64(i)*u + (1.0-float64(i))*(1-u)) * (float64(j)*v + (1.0-float64(j))*(1-v)) * (float64(k)*w + (1.0-float64(k))*(1-w)) * Dot(c[i][j][k], weight)
			}
		}
	}
	return acc
}

func (this *Perlin) Turb(p *Point3, depth int) float64 {
	acc := 0.0
	weight := 1.0
	ps := &Point3{p.X, p.Y, p.Z}
	for i := 0; i < depth; i++ {
		acc += weight * this.Noise(ps)
		weight *= 0.5
		ps = &Point3{p.X * 2.0, p.Y * 2.0, p.Z * 2.0}
	}
	return Abs(acc)
}
