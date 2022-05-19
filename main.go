package main

import (
	"fmt"
	. "github.com/alexa-infra/rayme/math"
	. "github.com/alexa-infra/rayme/render"
	"image"
	"image/png"
	"math"
	"os"
	"time"
)

const (
	aspectRatio     = 16.0 / 9.0
	imageWidth      = 400
	imageHeight     = int(imageWidth / aspectRatio)
	focalLength     = 1.0
	samplesPerPixel = 32
	maxDepth        = 50
	vfov            = 20.0
	aperture        = 0.1
	distToFocus     = 10
)

func main() {
	samples := []Vec2{Vec2{0.0, 0.0}}
	for i := 0; i < samplesPerPixel; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(samplesPerPixel)
		samples = append(samples, Vec2{math.Cos(angle), math.Sin(angle)})
	}
	camera := MakeCamera(&Point3{13, 2, 3}, &Point3{0, 0, 0}, &Vec3{0, 1, 0}, vfov, aspectRatio, aperture, distToFocus, 0.0, 0.1)

	ground := MakeLambertian(&Vec3{0.5, 0.5, 0.5})
	glass := MakeDielectric(1.5)

	world := &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, ground},
		},
	}
	ref := &Point3{4, 0.2, 0}
	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := RandGen.Float64()
			center := &Point3{float64(a) + 0.9*RandGen.Float64(), 0.2, float64(b) + 0.9*RandGen.Float64()}
			if Distance(center, ref) > 0.9 {
				if chooseMat < 0.8 {
					albedo := &Vec3{RandGen.Float64(), RandGen.Float64(), RandGen.Float64()}
					mat := MakeLambertian(albedo)
					center2 := center.Move(&Vec3{0, RandGen.Float64() * 0.5, 0.0})
					sphere := &MovingSphere{center, center2, 0.2, 0.0, 0.1, mat}
					world.Objects = append(world.Objects, sphere)
				} else if chooseMat > 0.95 {
					albedo := &Vec3{RandGen.Float64()/2.0 + 0.5, RandGen.Float64()/2.0 + 0.5, RandGen.Float64()/2.0 + 0.5}
					fuzz := RandGen.Float64() / 2.0
					mat := MakeMetal(albedo, fuzz)
					sphere := &Sphere{center, 0.2, mat}
					world.Objects = append(world.Objects, sphere)
				} else {
					sphere := &Sphere{center, 0.2, glass}
					world.Objects = append(world.Objects, sphere)
				}
			}

		}
	}

	material2 := MakeLambertian(&Vec3{0.4, 0.2, 0.1})
	material3 := MakeMetal(&Vec3{0.7, 0.6, 0.5}, 0.0)
	sphere1 := &Sphere{&Point3{0, 1, 0}, 1.0, glass}
	sphere2 := &Sphere{&Point3{-4, 1, 0}, 1.0, material2}
	sphere3 := &Sphere{&Point3{4, 1, 0}, 1.0, material3}
	world.Objects = append(world.Objects, sphere1)
	world.Objects = append(world.Objects, sphere2)
	world.Objects = append(world.Objects, sphere3)

	nRenders := int64(0)
	sumMicroseconds := int64(0)
	startFull := time.Now()
	myImg := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	for j := 0; j < imageHeight; j++ {
		for i := 0; i < imageWidth; i++ {
			sumColor := &Vec3{0, 0, 0}
			for _, s := range samples {
				u := (float64(i) + s.X) / float64(imageWidth-1)
				v := (float64(j) + s.Y) / float64(imageHeight-1)
				ray := camera.CastRay(u, v)
				start := time.Now()
				rayColor := GetRayColor(ray, world, maxDepth)
				sumMicroseconds += time.Since(start).Microseconds()
				nRenders++
				sumColor = sumColor.Add(rayColor)
			}
			scale := 1.0 / float64(len(samples))
			sumColor = &Vec3{
				math.Sqrt(sumColor.X * scale),
				math.Sqrt(sumColor.Y * scale),
				math.Sqrt(sumColor.Z * scale),
			}
			myImg.SetRGBA(i, imageHeight-j-1, sumColor.AsColor())
		}
	}
	fmt.Println("Full time:", time.Since(startFull).Seconds(), "seconds")
	fmt.Println("Number of rays:", nRenders)
	fmt.Println("Avg ray:", sumMicroseconds / nRenders, "microseconds")
	out, err := os.Create("output.png")
	if err != nil {
		fmt.Println("can't open file to write")
		os.Exit(1)
	}
	png.Encode(out, myImg)
	out.Close()
}
