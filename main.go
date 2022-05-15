package main

import (
	"fmt"
	. "github.com/alexa-infra/rayme/math"
	. "github.com/alexa-infra/rayme/render"
	"image"
	"image/png"
	"math"
	"os"
)

const (
	aspectRatio     = 16.0 / 9.0
	imageWidth      = 400
	imageHeight     = int(imageWidth / aspectRatio)
	focalLength     = 1.0
	samplesPerPixel = 12
	maxDepth        = 50
	vfov            = 20.0
	aperture        = 0.1
	distToFocus     = 10
)

func main() {
	samples := []Vec2{Vec2{0.0, 0.0}}
	for i := 0; i < samplesPerPixel; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(samplesPerPixel)
		samples = append(samples, Vec2{0.25 * math.Cos(angle), 0.25 * math.Sin(angle)})
	}
	camera := MakeCamera(Point3{-2,2,1}, Point3{0,0,-1}, Vec3{0,1,0}, vfov, aspectRatio, aperture, distToFocus)

	ground := MakeLambertian(Vec3{0.8, 0.8, 0.0})
	center := MakeLambertian(Vec3{0.1, 0.2, 0.5})
	left := MakeDielectric(1.5)
	right := MakeMetal(Vec3{0.8, 0.6, 0.2}, 0.0)

	world := &HittableList{
		[]Hittable{
			&Sphere{Point3{0.0, -100.5, -1.0}, 100.0, ground},
			&Sphere{Point3{0.0, 0.0, -1.0}, 0.5, center},
			&Sphere{Point3{-1.0, 0.0, -1.0}, 0.5, left},
			&Sphere{Point3{-1.0, 0.0, -1.0}, -0.4, left},
			&Sphere{Point3{ 1.0, 0.0, -1.0}, 0.5, right},
		},
	}

	myImg := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	for j := 0; j < imageHeight; j++ {
		for i := 0; i < imageWidth; i++ {
			sumColor := &Vec3{0, 0, 0}
			for _, s := range samples {
				u := (float64(i) + s.X) / float64(imageWidth-1)
				v := (float64(j) + s.Y) / float64(imageHeight-1)
				ray := camera.GetRay(u, v)
				rayColor := GetRayColor(ray, world, maxDepth)
				sumColor = sumColor.Add(rayColor)
			}
			scale := 1.0 / float64(len(samples))
			sumColor = &Vec3{
				math.Sqrt(sumColor.X * scale),
				math.Sqrt(sumColor.Y * scale),
				math.Sqrt(sumColor.Z * scale),
			}
			myImg.SetRGBA(i, imageHeight - j - 1, sumColor.AsColor())
		}
	}
	out, err := os.Create("output.png")
	if err != nil {
		fmt.Println("can't open file to write")
		os.Exit(1)
	}
	png.Encode(out, myImg)
	out.Close()
}
