package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type HitRecord struct {
	t         float64
	p         *Point3
	n         *Vec3
	frontFace bool
	Material
	u, v float64
}

func MakeHitRecord(ray *Ray, root float64, point *Point3, normal *Vec3, material Material, u, v float64) *HitRecord {
	frontFace := Dot(ray.Direction, normal) < 0
	if !frontFace {
		normal = normal.Mul(-1.0)
	}
	return &HitRecord{root, point, normal, frontFace, material, u, v}
}

type Hittable interface {
	hit(r *Ray, tMin, tMax float64) (bool, *HitRecord)
	boundingBox(t0, t1 float64) (bool, *Aabb)
	pdfValue(origin *Point3, v *Vec3) float64
	random(origin *Point3, rng *RandExt) *Vec3
}

type hittableNoPdf struct {}

func (this *hittableNoPdf) pdfValue(origin *Point3, v *Vec3) float64 {
	return 0.0
}

func (this *hittableNoPdf) random(origin *Point3, rng *RandExt) *Vec3 {
	return &Vec3{ 1, 0, 0 }
}

type Sphere struct {
	Center *Point3
	Radius float64
	Material
}

func getSphereUv(p *Vec3) (u, v float64) {
	theta := math.Acos(-p.Y)
	phi := math.Atan2(-p.Z, p.X) + math.Pi
	u = phi / (2.0 * math.Pi)
	v = theta / math.Pi
	return
}

func (this *Sphere) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	oc := GetDirection(this.Center, ray.Origin)
	a := Dot(ray.Direction, ray.Direction)
	h := Dot(oc, ray.Direction)
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
	normal := GetDirection(this.Center, hitPoint).Mul(1.0 / this.Radius)
	u, v := getSphereUv(normal)
	return true, MakeHitRecord(ray, root, hitPoint, normal, this.Material, u, v)
}

func (this *Sphere) boundingBox(t0, t1 float64) (bool, *Aabb) {
	r := this.Radius * math.Sqrt2
	radius := &Vec3{r, r, r}
	aabb := &Aabb{
		this.Center.Move(radius.Mul(-1)),
		this.Center.Move(radius),
	}
	return true, aabb
}

func (this *Sphere) pdfValue(origin *Point3, v *Vec3) float64 {
	hit, _ := this.hit(MakeRayFromDirection(origin, v, 0), 0.001, 10000)
	if !hit {
		return 0.0
	}
	distance2 := GetDirection(origin, this.Center).Length2()
	cosThetaMax := math.Sqrt(1 - this.Radius * this.Radius / distance2)
	solidAngle := 2 * math.Pi * (1 - cosThetaMax)
	return 1 / solidAngle
}

func (this *Sphere) random(origin *Point3, rng *RandExt) *Vec3 {
	dir := GetDirection(origin, this.Center)
	distance2 := dir.Length2()
	uvw := BuildOnbFromW(dir)
	return uvw.Local(rng.RandomToSphere(this.Radius, distance2))
}

type MovingSphere struct {
	Center0, Center1 *Point3
	Radius           float64
	Time0, Time1     float64
	Material
}

func (this *MovingSphere) center(time float64) *Point3 {
	scale := (time - this.Time0) / (this.Time1 - this.Time0)
	dir := GetDirection(this.Center0, this.Center1)
	return this.Center0.Move(dir.Mul(scale))
}

func (this *MovingSphere) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	c := this.center(ray.Time)
	sphere := Sphere{c, this.Radius, this.Material}
	return sphere.hit(ray, tMin, tMax)
}

func (this *MovingSphere) boundingBox(t0, t1 float64) (bool, *Aabb) {
	center0 := this.center(t0)
	sphere0 := Sphere{center0, this.Radius, this.Material}
	_, aabb0 := sphere0.boundingBox(t0, t1)

	center1 := this.center(t1)
	sphere1 := Sphere{center1, this.Radius, this.Material}
	_, aabb1 := sphere1.boundingBox(t0, t1)
	return true, SurroundingBox(aabb0, aabb1)
}

func (this *MovingSphere) pdfValue(origin *Point3, v *Vec3) float64 {
	return 0.0
}

func (this *MovingSphere) random(origin *Point3, rng *RandExt) *Vec3 {
	return &Vec3{ 1, 0, 0 }
}

type HittableList struct {
	Objects []Hittable
}

func (this *HittableList) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	var closest *HitRecord = nil
	for _, object := range this.Objects {
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

func (this *HittableList) boundingBox(t0, t1 float64) (bool, *Aabb) {
	found := false
	var surrounding *Aabb
	for _, obj := range this.Objects {
		ok, aabb := obj.boundingBox(t0, t1)
		if ok {
			if !found {
				surrounding = aabb
				found = true
			} else {
				surrounding = SurroundingBox(surrounding, aabb)
			}
		}
	}
	return found, surrounding
}

func (this *HittableList) pdfValue(origin *Point3, v *Vec3) float64 {
	weight := float64(1) / float64(len(this.Objects))
	sum := float64(0)
	for _, obj := range this.Objects {
		sum += weight * obj.pdfValue(origin, v)
	}
	return sum
}

func (this *HittableList) random(origin *Point3, rng *RandExt) *Vec3 {
	idx := rng.Intn(len(this.Objects))
	return this.Objects[idx].random(origin, rng)
}

func GetRayColor(r *Ray, bgColor *Vec3, world Hittable, lights Hittable, depth int, rng *RandExt) *Vec3 {
	if depth <= 0 {
		return noColor
	}
	hit, rec := world.hit(r, 0.001, 10000)
	if !hit {
		return bgColor
	}

	var emitted *Vec3
	if rec.frontFace {
		emitted = rec.Material.Emitted(rec.u, rec.v, rec.p)
	} else {
		emitted = &Vec3{0, 0, 0}
	}
	scattered, srec := rec.Material.Scatter(r, rec, rng)
	if !scattered {
		return emitted
	}
	if !srec.isSpecular && lights != nil {
		lightPdf := MakeHittablePdf(lights, rec.p)
		cosinePdf := MakeCosinePdf(rec.n)
		mixPdf := MakeMixturePdf(lightPdf, cosinePdf)

		srec.specular = MakeRayFromDirection(rec.p, mixPdf.generate(rng), r.Time)
		pdfVal := mixPdf.value(srec.specular.Direction)
		srec.attenuation = srec.attenuation.Mul(rec.Material.ScatteringPDF(r, rec, srec.specular)).Mul(1 / pdfVal)
	}
	return GetRayColor(srec.specular, bgColor, world, lights, depth-1, rng).MulVec(srec.attenuation).Add(emitted)
}

type RectXY struct {
	x0, y0 float64
	x1, y1 float64
	k      float64
	Material
	hittableNoPdf
}

func MakeRectXY(x0, y0, x1, y1, k float64, m Material) *RectXY {
	return &RectXY{x0, y0, x1, y1, k, m, hittableNoPdf{}}
}

func (this *RectXY) boundingBox(t0, t1 float64) (bool, *Aabb) {
	a := &Point3{this.x0, this.y0, this.k - 0.0001}
	b := &Point3{this.x1, this.y1, this.k + 0.0001}
	return true, &Aabb{a, b}
}

func (this *RectXY) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	t := (this.k - ray.Origin.Z) / ray.Direction.Z
	if t < tMin || t > tMax {
		return false, nil
	}
	x := ray.Origin.X + t*ray.Direction.X
	y := ray.Origin.Y + t*ray.Direction.Y
	if x < this.x0 || x > this.x1 || y < this.y0 || y > this.y1 {
		return false, nil
	}
	hitPoint := ray.At(t)
	u := (x - this.x0) / (this.x1 - this.x0)
	v := (y - this.y0) / (this.y1 - this.y0)
	normal := &Vec3{0, 0, 1}
	return true, MakeHitRecord(ray, t, hitPoint, normal, this.Material, u, v)
}

type RectXZ struct {
	x0, z0 float64
	x1, z1 float64
	k      float64
	Material
}

func MakeRectXZ(x0, z0, x1, z1, k float64, m Material) *RectXZ {
	return &RectXZ{x0, z0, x1, z1, k, m}
}

func (this *RectXZ) boundingBox(t0, t1 float64) (bool, *Aabb) {
	a := &Point3{this.x0, this.z0, this.k - 0.0001}
	b := &Point3{this.x1, this.z1, this.k + 0.0001}
	return true, &Aabb{a, b}
}

func (this *RectXZ) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	t := (this.k - ray.Origin.Y) / ray.Direction.Y
	if t < tMin || t > tMax {
		return false, nil
	}
	x := ray.Origin.X + t*ray.Direction.X
	z := ray.Origin.Z + t*ray.Direction.Z
	if x < this.x0 || x > this.x1 || z < this.z0 || z > this.z1 {
		return false, nil
	}
	hitPoint := ray.At(t)
	u := (x - this.x0) / (this.x1 - this.x0)
	v := (z - this.z0) / (this.z1 - this.z0)
	normal := &Vec3{0, 1, 0}
	return true, MakeHitRecord(ray, t, hitPoint, normal, this.Material, u, v)
}

func (this *RectXZ) pdfValue(origin *Point3, v *Vec3) float64 {
	hit, rec := this.hit(MakeRayFromDirection(origin, v, 0), 0.001, 10000)
	if !hit {
		return 0.0
	}
	area := (this.x1 - this.x0) * (this.z1 - this.z0)
	distance2 := rec.t * rec.t
	cosine := Abs(Dot(v, rec.n))
	return distance2 / (cosine * area)
}

func (this *RectXZ) random(origin *Point3, rng *RandExt) *Vec3 {
	randomPoint := &Point3{ rng.Between(this.x0, this.x1), this.k, rng.Between(this.z0, this.z1) }
	return GetDirection(origin, randomPoint)
}

type RectYZ struct {
	y0, z0 float64
	y1, z1 float64
	k      float64
	Material
	hittableNoPdf
}

func MakeRectYZ(y0, z0, y1, z1, k float64, m Material) *RectYZ {
	return &RectYZ{y0, z0, y1, z1, k, m, hittableNoPdf{}}
}

func (this *RectYZ) boundingBox(t0, t1 float64) (bool, *Aabb) {
	a := &Point3{this.y0, this.z0, this.k - 0.0001}
	b := &Point3{this.y1, this.z1, this.k + 0.0001}
	return true, &Aabb{a, b}
}

func (this *RectYZ) hit(ray *Ray, tMin, tMax float64) (bool, *HitRecord) {
	t := (this.k - ray.Origin.X) / ray.Direction.X
	if t < tMin || t > tMax {
		return false, nil
	}
	y := ray.Origin.Y + t*ray.Direction.Y
	z := ray.Origin.Z + t*ray.Direction.Z
	if y < this.y0 || y > this.y1 || z < this.z0 || z > this.z1 {
		return false, nil
	}
	hitPoint := ray.At(t)
	u := (y - this.y0) / (this.y1 - this.y0)
	v := (z - this.z0) / (this.z1 - this.z0)
	normal := &Vec3{1, 0, 0}
	return true, MakeHitRecord(ray, t, hitPoint, normal, this.Material, u, v)
}

type Box struct {
	min, max *Point3
	sides HittableList
	hittableNoPdf
}

func (this *Box) hit(r *Ray, tMin, tMax float64) (bool, *HitRecord) {
	return this.sides.hit(r, tMin, tMax)
}

func (this *Box) boundingBox(t0, t1 float64) (bool, *Aabb) {
	return true, &Aabb{this.min, this.max}
}

func MakeBox(p0, p1 *Point3, m Material) *Box {
	sides := HittableList{
		[]Hittable{
			MakeRectXY(p0.X, p0.Y, p1.X, p1.Y, p1.Z, m),
			MakeRectXY(p0.X, p0.Y, p1.X, p1.Y, p0.Z, m),
			MakeRectXZ(p0.X, p0.Z, p1.X, p1.Z, p1.Y, m),
			MakeRectXZ(p0.X, p0.Z, p1.X, p1.Z, p0.Y, m),
			MakeRectYZ(p0.Y, p0.Z, p1.Y, p1.Z, p1.X, m),
			MakeRectYZ(p0.Y, p0.Z, p1.Y, p1.Z, p0.X, m),
		},
	}
	return &Box{p0, p1, sides, hittableNoPdf{}}
}

type Translate struct {
	obj    Hittable
	offset *Vec3
}

func (this *Translate) hit(r *Ray, tMin, tMax float64) (bool, *HitRecord) {
	ray := MakeRayFromDirection(r.Origin.Move(this.offset.Mul(-1)), r.Direction, r.Time)
	hit, rec := this.obj.hit(ray, tMin, tMax)
	if !hit {
		return false, nil
	}
	return true, MakeHitRecord(ray, rec.t, rec.p.Move(this.offset), rec.n, rec.Material, rec.u, rec.v)
}

func (this *Translate) boundingBox(t0, t1 float64) (bool, *Aabb) {
	ok, box := this.obj.boundingBox(t0, t1)
	if !ok {
		return false, nil
	}
	return true, &Aabb{box.Min.Move(this.offset), box.Max.Move(this.offset)}
}

func MakeTranslate(obj Hittable, displacement *Vec3) *Translate {
	return &Translate{obj, displacement}
}

func (this *Translate) pdfValue(origin *Point3, v *Vec3) float64 {
	return this.obj.pdfValue(origin.Move(this.offset), v)
}

func (this *Translate) random(origin *Point3, rng *RandExt) *Vec3 {
	return this.obj.random(origin.Move(this.offset), rng)
}

type RotateY struct {
	obj    Hittable
	sinTheta, cosTheta float64
	hasBox bool
	bbox   *Aabb
	hittableNoPdf
}

func MakeRotateY(obj Hittable, angle float64) *RotateY {
	radians := DegreesToRadians(angle)
	sinTheta := math.Sin(radians)
	cosTheta := math.Cos(radians)
	hasBox, box := obj.boundingBox(0, 1)
	inf := math.Inf(1)
	ninf := math.Inf(-1)
	min := Point3{inf, inf, inf}
	max := Point3{ninf, ninf, ninf}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				x := float64(i) * box.Max.X + (1 - float64(i)) * box.Min.X
				y := float64(j) * box.Max.Y + (1 - float64(j)) * box.Min.Y
				z := float64(k) * box.Max.Z + (1 - float64(k)) * box.Min.Z
				nx := cosTheta*x + sinTheta*z
				nz := -sinTheta*x + cosTheta*z
				p := Point3{nx, y, nz}
				min.X = Min(min.X, p.X)
				max.X = Max(max.X, p.X)
				min.Y = Min(min.Y, p.Y)
				max.Y = Max(max.Y, p.Y)
				min.Z = Min(min.Z, p.Z)
				max.Z = Max(max.Z, p.Z)
			}
		}
	}
	bbox := Aabb{&min, &max}
	return &RotateY{obj, sinTheta, cosTheta, hasBox, &bbox, hittableNoPdf{}}
}

func (this *RotateY) boundingBox(t0, t1 float64) (bool, *Aabb) {
	return this.hasBox, this.bbox
}

func (this *RotateY) hit(r *Ray, tMin, tMax float64) (bool, *HitRecord) {
	origin := Point3{0, 0, 0}
	direction := Vec3{0, 0, 0}

	origin.X = this.cosTheta*r.Origin.X - this.sinTheta*r.Origin.Z
	origin.Y = r.Origin.Y
	origin.Z = this.sinTheta*r.Origin.X + this.cosTheta*r.Origin.Z

	direction.X = this.cosTheta*r.Direction.X - this.sinTheta*r.Direction.Z
	direction.Y = r.Direction.Y
	direction.Z = this.sinTheta*r.Direction.X + this.cosTheta*r.Direction.Z

	rotated := MakeRayFromDirection(&origin, &direction, r.Time)
	hit, rec := this.obj.hit(rotated, tMin, tMax)
	if !hit {
		return false, nil
	}
	p := Point3{0, 0, 0}
	normal := Vec3{0, 0, 0}

	p.X = this.cosTheta*rec.p.X + this.sinTheta*rec.p.Z
	p.Y = rec.p.Y
	p.Z = -this.sinTheta*rec.p.X + this.cosTheta*rec.p.Z

	normal.X = this.cosTheta*rec.n.X + this.sinTheta*rec.n.Z
	normal.Y = rec.n.Y
	normal.Z = -this.sinTheta*rec.n.X + this.cosTheta*rec.n.Z

	return true, MakeHitRecord(rotated, rec.t, &p, &normal, rec.Material, rec.u, rec.v)
}

type FlipFace struct {
	obj    Hittable
	hittableNoPdf
}

func (this *FlipFace) boundingBox(t0, t1 float64) (bool, *Aabb) {
	return this.obj.boundingBox(t0, t1)
}

func (this *FlipFace) hit(r *Ray, tMin, tMax float64) (bool, *HitRecord) {
	hit, rec := this.obj.hit(r, tMin, tMax)
	if !hit {
		return false, nil
	}
	rec.frontFace = !rec.frontFace
	return hit, rec
}

func MakeFlipFace(obj Hittable) *FlipFace {
	return &FlipFace{ obj, hittableNoPdf{} }
}
