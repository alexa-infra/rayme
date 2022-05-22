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
	camera := MakeCamera(&Point3{13, 2, 3}, &Point3{0, 0, 0}, &Vec3{0, 1, 0}, vfov, aspectRatio, aperture, distToFocus, 0.0, 1.0)

	ground := MakeLambertianTexture(MakeCheckerTexture(&Vec3{0.2, 0.3, 0.1}, &Vec3{0.9, 0.9, 0.9}))
	glass := MakeDielectric(1.5)

	world := &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, ground},
		},
	}
	ref := &Point3{4, 0.2, 0}
	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := RandomBetween(0.0, 1.0)
			center := &Point3{float64(a) + RandomBetween(0.0, 0.9), 0.2, float64(b) + RandomBetween(0.0, 0.9)}
			if Distance(center, ref) > 0.9 {
				if chooseMat < 0.8 {
					albedo := RandomInUnitSphere()
					mat := MakeLambertianSolidColor(albedo)
					if chooseMat > 0.7 {
						center2 := center.Move(&Vec3{0, RandomBetween(0.0, 0.5), 0.0})
						sphere := &MovingSphere{center, center2, 0.2, 0.0, 1.0, mat}
						world.Objects = append(world.Objects, sphere)
					} else {
						sphere := &Sphere{center, 0.2, mat}
						world.Objects = append(world.Objects, sphere)
					}
				} else if chooseMat > 0.95 {
					albedo := &Vec3{RandomBetween(0.5, 1.0), RandomBetween(0.5, 1.0), RandomBetween(0.5, 1.0)}
					fuzz := RandomBetween(0.0, 0.5)
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

	material2 := MakeLambertianSolidColor(&Vec3{0.4, 0.2, 0.1})
	material3 := MakeMetal(&Vec3{0.7, 0.6, 0.5}, 0.0)
	sphere1 := &Sphere{&Point3{0, 1, 0}, 1.0, glass}
	sphere2 := &Sphere{&Point3{-4, 1, 0}, 1.0, material2}
	sphere3 := &Sphere{&Point3{4, 1, 0}, 1.0, material3}
	world.Objects = append(world.Objects, sphere1)
	world.Objects = append(world.Objects, sphere2)
	world.Objects = append(world.Objects, sphere3)

	bvh := MakeBvhFromList(world, 0.0, 1.0)

	startFull := time.Now()

	render := func(u, v float64, out chan *Vec3) {
		ray := camera.CastRay(u, v)
		rayColor := GetRayColor(ray, bvh, maxDepth)
		out <- rayColor
	}

	myImg := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	for j := 0; j < imageHeight; j++ {
		for i := 0; i < imageWidth; i++ {
			channel := make(chan *Vec3)
			for _, s := range samples {
				u := (float64(i) + s.X) / float64(imageWidth-1)
				v := (float64(j) + s.Y) / float64(imageHeight-1)
				go render(u, v, channel)
			}
			sumColor := &Vec3{0, 0, 0}
			for k := 0; k < len(samples); k++ {
				rayColor := <-channel
				sumColor = sumColor.Add(rayColor)
			}
			close(channel)
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
	out, err := os.Create("output.png")
	if err != nil {
		fmt.Println("can't open file to write")
		os.Exit(1)
	}
	png.Encode(out, myImg)
	out.Close()
}
