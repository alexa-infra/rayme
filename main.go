package main

import (
	"flag"
	"fmt"
	. "github.com/alexa-infra/rayme/math"
	. "github.com/alexa-infra/rayme/render"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"time"
)

const (
	focalLength = 1.0
	maxDepth    = 50
	distToFocus = 10
)

var (
	sceneID                  = flag.Int("scene", 0, "Scene ID")
	lookFrom        *Point3  = nil
	lookAt          *Point3  = nil
	vfov                     = 20.0
	aperture                 = 0.0
	world           Hittable = nil
	bgColor         *Vec3    = nil
	aspectRatio              = 16.0 / 9.0
	imageWidth               = 400
	samplesPerPixel          = 12
)

func main() {
	flag.Parse()

	if *sceneID == 0 {
		world = randomScene()
		lookFrom = &Point3{13, 2, 3}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		aperture = 0.1
		bgColor = &Vec3{0.7, 0.8, 1.0}
	} else if *sceneID == 1 {
		world = twoSpheresScene()
		lookFrom = &Point3{13, 2, 3}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		aperture = 0.0
		bgColor = &Vec3{0.7, 0.8, 1.0}
	} else if *sceneID == 2 {
		world = earthSphereScene()
		lookFrom = &Point3{13, 2, 3}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		bgColor = &Vec3{0.7, 0.8, 1.0}
	} else if *sceneID == 3 {
		world = simpleLight()
		lookFrom = &Point3{26, 3, 6}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		aperture = 0.0
		bgColor = &Vec3{0.0, 0.0, 0.0}
	} else if *sceneID == 4 {
		world = cornelBox()
		lookFrom = &Point3{278, 278, -800}
		lookAt = &Point3{278, 278, 0}
		vfov = 40.0
		aperture = 0.0
		bgColor = &Vec3{0.0, 0.0, 0.0}
		aspectRatio = 1.0
		imageWidth = 600
		samplesPerPixel = 64
	} else {
		fmt.Println("unknown sceneID")
		os.Exit(1)
	}
	samples := []Vec2{Vec2{0.0, 0.0}}
	for i := 0; i < samplesPerPixel; i++ {
		angle := RandomBetween(0, 2 * math.Pi)
		r := RandomBetween(0, 1)
		samples = append(samples, Vec2{r * math.Cos(angle), r * math.Sin(angle)})
	}
	camera := MakeCamera(lookFrom, lookAt, &Vec3{0, 1, 0}, vfov, aspectRatio, aperture, distToFocus, 0.0, 1.0)

	startFull := time.Now()

	render := func(u, v float64, out chan *Vec3) {
		ray := camera.CastRay(u, v)
		rayColor := GetRayColor(ray, bgColor, world, maxDepth)
		out <- rayColor
	}

	imageHeight := int(float64(imageWidth) / aspectRatio)
	myImg := image.NewRGBA64(image.Rect(0, 0, imageWidth, imageHeight))
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
			myImg.Set(i, imageHeight-j-1, sumColor.AsColor())
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

func randomScene() Hittable {
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
	return bvh
}

func twoSpheresScene() Hittable {
	noise := MakeNoiseTexture(4.0)
	material1 := MakeLambertianTexture(noise)
	checker := MakeCheckerTexture(&Vec3{0.2, 0.3, 0.1}, &Vec3{0.9, 0.9, 0.9})
	material2 := MakeLambertianTexture(checker)
	world := &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, material1},
			&Sphere{&Point3{0.0, 2, 0.0}, 2.0, material2},
		},
	}
	return world
}

func earthSphereScene() Hittable {
	tex, err := MakeImageTexture("./earthmap.jpg")
	if err != nil {
		log.Fatal(err)
	}
	material := MakeLambertianTexture(tex)
	world := &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, 0.0, 0.0}, 2.0, material},
		},
	}
	return world
}

func simpleLight() Hittable {
	noise := MakeNoiseTexture(4.0)
	material1 := MakeLambertianTexture(noise)
	difflight := MakeDiffuseLightFromColor(&Vec3{4, 4, 4})
	world := &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, material1},
			&Sphere{&Point3{0.0, 2, 0.0}, 2.0, material1},
			MakeRectXY(3, 1, 5, 3, -2, difflight),
		},
	}
	return world
}

func cornelBox() Hittable {
	red := MakeLambertianSolidColor(&Vec3{0.65, 0.05, 0.05})
	white := MakeLambertianSolidColor(&Vec3{0.73, 0.73, 0.73})
	green := MakeLambertianSolidColor(&Vec3{0.12, 0.45, 0.15})
	light := MakeDiffuseLightFromColor(&Vec3{15, 15, 15})

	world := &HittableList{
		[]Hittable{
			MakeRectYZ(0, 0, 555, 555, 555, green),
			MakeRectYZ(0, 0, 555, 555, 0, red),
			MakeRectXZ(0, 0, 555, 555, 555, white),
			MakeRectXZ(0, 0, 555, 555, 0, white),
			MakeRectXY(0, 0, 555, 555, 555, white),
			MakeRectXZ(213, 227, 343, 332, 554, light),
		},
	}
	return world
}
