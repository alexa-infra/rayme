package render

import (
	. "github.com/alexa-infra/rayme/math"
	"math"
)

type MeshBuilder struct {
	indexArr    []int
	posArr      []*Point3
	vertexArr   []int
	lastIndex   int
	faceCount   int
	vertexCache map[int]int
}

type Triangle struct {
	v0, v1, v2 *Point3
	normal     *Vec3
	material   Material
}

func MakeMeshBuilder() *MeshBuilder {
	indexArr := []int{}
	posArr := []*Point3{}
	vertexArr := []int{}
	vertexCache := map[int]int{}
	return &MeshBuilder{indexArr, posArr, vertexArr, 0, 0, vertexCache}
}

func (this *MeshBuilder) AddPosition(p *Point3) int {
	idx := len(this.posArr)
	this.posArr = append(this.posArr, p)
	return idx
}

func (this *MeshBuilder) BeginPolygon() {
	this.lastIndex = len(this.indexArr)
	this.indexArr = append(this.indexArr, 0)
}

func (this *MeshBuilder) AddVertex(pos int) {
	idx, ok := this.vertexCache[pos]
	if !ok {
		idx = len(this.vertexArr)
		this.vertexArr = append(this.vertexArr, pos)
		this.vertexCache[pos] = idx
	}
	this.indexArr = append(this.indexArr, idx)
	this.indexArr[this.lastIndex] += 1
}

func (this *MeshBuilder) EndPolygon() {
	count := this.indexArr[this.lastIndex]
	if count <= 2 {
		this.indexArr = this.indexArr[:this.lastIndex]
	} else {
		this.faceCount++
	}
}

func (this *MeshBuilder) GetTriMesh(material Material, rng *RandExt) Hittable {
	faces := []Hittable{}
	for i := 0; i < len(this.indexArr); i++ {
		count := this.indexArr[i]
		v0 := this.indexArr[i+1]
		v1 := this.indexArr[i+2]
		for j := 0; j < count-2; j++ {
			v2 := this.indexArr[i+3+j]
			p0 := this.posArr[this.vertexArr[v0]]
			p1 := this.posArr[this.vertexArr[v1]]
			p2 := this.posArr[this.vertexArr[v2]]
			a := GetDirection(p0, p1)
			b := GetDirection(p0, p2)
			n := Cross(b, a).Normalize()
			//n = &Vec3{1, 0, 0}
			faces = append(faces, &Triangle{p0, p1, p2, n, material})
			v1 = v2
		}
		i += count
	}
	return MakeBvh(faces, 0, 1, rng)
}

func (this *Triangle) hit(r *Ray, tMin, tMax float64) (bool, *HitRecord) {
	edge1 := GetDirection(this.v0, this.v1)
	edge2 := GetDirection(this.v0, this.v2)
	h := Cross(r.Direction, edge2)
	det := Dot(edge1, h)
	if Abs(det) < 1e-8 {
		return false, nil // in triangle plane
	}
	invDet := float64(1) / det
	s := GetDirection(this.v0, r.Origin)
	u := Dot(s, h) * invDet
	if u < 0 || u > 1 {
		return false, nil
	}
	q := Cross(s, edge1)
	v := Dot(r.Direction, q) * invDet
	if v < 0 || u+v > 1 {
		return false, nil
	}
	t := Dot(edge2, q) * invDet
	//if Abs(t) < 1e-8 {
	if t < tMin || t > tMax {
		return false, nil
	}
	return true, MakeHitRecord(r, t, r.At(t), this.normal, this.material, 0, 0)
}

func (this *Triangle) boundingBox(t0, t1 float64) (bool, *Aabb) {
	builder := MakeAabbBuilder()
	builder.AddPoint(this.v0)
	builder.AddPoint(this.v1)
	builder.AddPoint(this.v2)
	return true, builder.GetBox()
}

func (this *Triangle) pdfValue(origin *Point3, v *Vec3) float64 {
	return 0.0
}

func (this *Triangle) random(origin *Point3, rng *RandExt) *Vec3 {
	return &Vec3{1, 0, 0}
}

func MakeCubeMesh(material Material, rng *RandExt) Hittable {
	x := &Vec3{1, 0, 0}
	y := &Vec3{0, 1, 0}
	z := &Vec3{0, 0, 1}
	origin := MakePoint3(-0.5, -0.5, -0.5)
	meshBuilder := MakeMeshBuilder()
	meshBuilder.AddPosition(origin)
	meshBuilder.AddPosition(origin.Move(x))
	meshBuilder.AddPosition(origin.Move(x).Move(z))
	meshBuilder.AddPosition(origin.Move(z))
	meshBuilder.AddPosition(origin.Move(y))
	meshBuilder.AddPosition(origin.Move(y).Move(x))
	meshBuilder.AddPosition(origin.Move(y).Move(x).Move(z))
	meshBuilder.AddPosition(origin.Move(y).Move(z))
	addQuad := func(v0, v1, v2, v3 int) {
		meshBuilder.BeginPolygon()
		meshBuilder.AddVertex(v0)
		meshBuilder.AddVertex(v1)
		meshBuilder.AddVertex(v2)
		meshBuilder.AddVertex(v3)
		meshBuilder.EndPolygon()
	}
	addQuad(0, 1, 2, 3)
	addQuad(0, 1, 5, 4)
	addQuad(2, 3, 7, 6)
	addQuad(1, 2, 6, 5)
	addQuad(0, 4, 7, 3)
	addQuad(7, 6, 5, 4)
	return meshBuilder.GetTriMesh(material, rng)
}

func MakeSphereMesh(material Material, radius float64, numSegments int, rng *RandExt) Hittable {
	meshBuilder := MakeMeshBuilder()
	for i := 1; i < numSegments-1; i++ {
		y := float64(1) - float64(2*i)/float64(numSegments-1)
		r := math.Sin(math.Acos(y)) * radius
		for j := 0; j < numSegments; j++ {
			p := (math.Pi * float64(2*j)) / float64(numSegments)
			x := r * math.Sin(p)
			z := r * math.Cos(p)
			v := MakePoint3(x, y*radius, z)
			meshBuilder.AddPosition(v)
		}
	}
	top := meshBuilder.AddPosition(MakePoint3(0, radius, 0))
	bottom := meshBuilder.AddPosition(MakePoint3(0, -radius, 0))

	for i := 1; i < numSegments-2; i++ {
		for j := 0; j < numSegments-1; j++ {
			nextRow := i*numSegments + j
			prevRow := nextRow - numSegments
			meshBuilder.BeginPolygon()
			meshBuilder.AddVertex(prevRow)
			meshBuilder.AddVertex(nextRow)
			meshBuilder.AddVertex(nextRow + 1)
			meshBuilder.AddVertex(prevRow + 1)
			meshBuilder.EndPolygon()
		}
		nextLast := i*numSegments + (numSegments - 1)
		nextFirst := i * numSegments
		prevLast := nextLast - numSegments
		prevFirst := nextFirst - numSegments
		meshBuilder.BeginPolygon()
		meshBuilder.AddVertex(prevLast)
		meshBuilder.AddVertex(nextLast)
		meshBuilder.AddVertex(nextFirst)
		meshBuilder.AddVertex(prevFirst)
		meshBuilder.EndPolygon()
	}
	for i := 0; i < numSegments-1; i++ {
		meshBuilder.BeginPolygon()
		meshBuilder.AddVertex(i)
		meshBuilder.AddVertex(i + 1)
		meshBuilder.AddVertex(top)
		meshBuilder.EndPolygon()
	}
	meshBuilder.BeginPolygon()
	meshBuilder.AddVertex(numSegments - 1)
	meshBuilder.AddVertex(0)
	meshBuilder.AddVertex(top)
	meshBuilder.EndPolygon()

	lastRow := (numSegments - 3) * numSegments
	for i := 0; i < numSegments-1; i++ {
		meshBuilder.BeginPolygon()
		meshBuilder.AddVertex(lastRow + i)
		meshBuilder.AddVertex(lastRow + i + 1)
		meshBuilder.AddVertex(bottom)
		meshBuilder.EndPolygon()
	}
	meshBuilder.BeginPolygon()
	meshBuilder.AddVertex(lastRow + (numSegments - 1))
	meshBuilder.AddVertex(lastRow)
	meshBuilder.AddVertex(bottom)
	meshBuilder.EndPolygon()

	return meshBuilder.GetTriMesh(material, rng)
}
