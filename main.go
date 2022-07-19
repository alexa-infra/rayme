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
	"sync"
)

const (
	focalLength = 1.0
	maxDepth    = 50
	distToFocus = 10
	seed        = 99
)

var (
	sceneID                  = flag.Int("scene", 0, "Scene ID")
	lookFrom        *Point3  = nil
	lookAt          *Point3  = nil
	vfov                     = 20.0
	aperture                 = 0.0
	world           Hittable = nil
	lights          Hittable = nil
	bgColor         *Vec3    = nil
	aspectRatio              = 16.0 / 9.0
	imageWidth               = 400
	samplesPerPixel          = 12
	rng             *RandExt = nil
)

func main() {
	flag.Parse()
	rng = MakeRandExt(seed)

	if *sceneID == 0 {
		world = randomScene()
		lookFrom = &Point3{13, 2, 3}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		aperture = 0.1
		bgColor = &Vec3{0.7, 0.8, 1.0}
		imageWidth = 1200
		samplesPerPixel = 32
	} else if *sceneID == 1 {
		world, lights = twoSpheresScene()
		lookFrom = &Point3{13, 7, 3}
		lookAt = &Point3{0, 1, 0}
		vfov = 30.0
		aperture = 0.0
		samplesPerPixel = 128
		imageWidth = 640
		bgColor = &Vec3{0.3, 0.3, 0.3}
	} else if *sceneID == 2 {
		world = earthSphereScene()
		lookFrom = &Point3{13, 2, 3}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		bgColor = &Vec3{0.7, 0.8, 1.0}
	} else if *sceneID == 3 {
		world, lights = simpleLight()
		lookFrom = &Point3{26, 3, 6}
		lookAt = &Point3{0, 0, 0}
		vfov = 20.0
		aperture = 0.0
		bgColor = &Vec3{0.0, 0.0, 0.0}
	} else if *sceneID == 4 {
		world, lights = cornellBox()
		lookFrom = &Point3{278, 278, -800}
		lookAt = &Point3{278, 278, 0}
		vfov = 40.0
		aperture = 0.0
		bgColor = &Vec3{0.0, 0.0, 0.0}
		aspectRatio = 1.0
		imageWidth = 500
		samplesPerPixel = 10
	} else {
		fmt.Println("unknown sceneID")
		os.Exit(1)
	}
	samples := []*Vec3{&Vec3{0.0, 0.0, 0.0}}
	for i := 0; i < samplesPerPixel; i++ {
		samples = append(samples, rng.RandomInUnitDisk())
	}
	camera := MakeCamera(lookFrom, lookAt, &Vec3{0, 1, 0}, vfov, aspectRatio, aperture, distToFocus, 0.0, 1.0)

	startFull := time.Now()

	imageHeight := int(float64(imageWidth) / aspectRatio)
	myImg := image.NewRGBA64(image.Rect(0, 0, imageWidth, imageHeight))
	scale := 1.0 / float64(len(samples))
	zero := Vec3{0, 0, 0}
	type Pixel struct {
		x, y  int
	}
	in := make(chan Pixel)
	var wg sync.WaitGroup
	renderPixel := func(workerId int) {
		rng1 := MakeRandExt(seed + workerId + 1)
		for {
			p, ok := <- in
			if !ok {
				break
			}
			sumColor := &zero
			for _, s := range samples {
				u := (float64(p.x) + s.X) / float64(imageWidth-1)
				v := (float64(p.y) + s.Y) / float64(imageHeight-1)
				ray := camera.CastRay(u, v, rng1)
				rayColor := GetRayColor(ray, bgColor, world, lights, maxDepth, rng1)
				sumColor = sumColor.Add(rayColor)
			}
			sumColor = &Vec3{
				math.Sqrt(sumColor.X * scale),
				math.Sqrt(sumColor.Y * scale),
				math.Sqrt(sumColor.Z * scale),
			}
			myImg.Set(p.x, imageHeight-p.y-1, sumColor.AsColor())
		}
		wg.Done()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go renderPixel(i)
	}
	for j := 0; j < imageHeight; j++ {
		for i := 0; i < imageWidth; i++ {
			in <- Pixel{i, j}
		}
	}
	close(in)
	wg.Wait()
	fmt.Printf("\nFull time: %9.2f seconds\n", time.Since(startFull).Seconds())
	outImage, err := os.Create("output.png")
	if err != nil {
		fmt.Println("can't open file to write")
		os.Exit(1)
	}
	png.Encode(outImage, myImg)
	outImage.Close()
}

func randomScene() Hittable {
	ground := MakeLambertianTexture(MakeCheckerTexture3d(1.0, &Vec3{0.2, 0.3, 0.1}, &Vec3{0.9, 0.9, 0.9}))
	glass := MakeDielectric(1.5)

	world := &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, ground},
		},
	}
	ref := &Point3{4, 0.2, 0}
	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := rng.Between(0.0, 1.0)
			center := &Point3{float64(a) + rng.Between(0.0, 0.9), 0.2, float64(b) + rng.Between(0.0, 0.9)}
			if Distance(center, ref) > 0.9 {
				if chooseMat < 0.8 {
					albedo := rng.RandomInUnitSphere()
					mat := MakeLambertianSolidColor(albedo)
					if chooseMat > 0.7 {
						center2 := center.Move(&Vec3{0, rng.Between(0.0, 0.5), 0.0})
						sphere := &MovingSphere{center, center2, 0.2, 0.0, 1.0, mat}
						world.Objects = append(world.Objects, sphere)
					} else {
						sphere := &Sphere{center, 0.2, mat}
						world.Objects = append(world.Objects, sphere)
					}
				} else if chooseMat > 0.95 {
					albedo := &Vec3{rng.Between(0.5, 1.0), rng.Between(0.5, 1.0), rng.Between(0.5, 1.0)}
					fuzz := rng.Between(0.0, 0.5)
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

	bvh := MakeBvhFromList(world, 0.0, 1.0, rng)
	return bvh
}

func twoSpheresScene() (world, light Hittable) {
	red := MakeMetal(&Vec3{0.9, 0.1, 0.1}, 0.1)
	blue := MakeMetal(&Vec3{0.1, 0.1, 0.9}, 0.1)
	checker := MakeCheckerTexture3d(1.0, &Vec3{0.2, 0.3, 0.1}, &Vec3{0.9, 0.9, 0.9})
	material2 := MakeLambertianTexture(checker)
	lightMat := MakeDiffuseLightFromColor(&Vec3{15, 15, 15})
	light = &Sphere{&Point3{3.0, 25, 1.0}, 3.0, lightMat}
	world = &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, material2},
			&Sphere{&Point3{0.0, 2, 0.0}, 2.0, red},
			&Sphere{&Point3{3.0, 1, -1.0}, 1.0, blue},
			light,
		},
	}
	return
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

func simpleLight() (world, lights Hittable) {
	noise := MakeNoiseTexture(4.0, rng)
	material1 := MakeLambertianTexture(noise)
	difflight := MakeDiffuseLightFromColor(&Vec3{4, 4, 4})
	lights = MakeRectXY(5, 1, 5, 3, -2, difflight)
	world = &HittableList{
		[]Hittable{
			&Sphere{&Point3{0.0, -1000, 0.0}, 1000.0, material1},
			&Sphere{&Point3{0.0, 2, 0.0}, 2.0, material1},
			lights,
		},
	}
	return
}

func cornellBox() (world, lights Hittable) {
	red := MakeLambertianSolidColor(&Vec3{0.65, 0.05, 0.05})
	white := MakeLambertianSolidColor(&Vec3{0.73, 0.73, 0.73})
	green := MakeLambertianSolidColor(&Vec3{0.12, 0.45, 0.15})
	light := MakeDiffuseLightFromColor(&Vec3{15, 15, 15})

	world = &HittableList{
		[]Hittable{
			MakeRectYZ(0, 0, 555, 555, 555, green),
			MakeRectYZ(0, 0, 555, 555, 0, red),
			MakeRectXZ(0, 0, 555, 555, 555, white),
			MakeRectXZ(0, 0, 555, 555, 0, white),
			MakeRectXY(0, 0, 555, 555, 555, white),
			MakeFlipFace(MakeRectXZ(213, 227, 343, 332, 554, light)),
			MakeTranslate(
				MakeRotateY(
					MakeBox(&Point3{0, 0, 0}, &Point3{165, 330, 165}, white),
					15,
				),
				&Vec3{265,0,295},
			),
			MakeTranslate(
				MakeRotateY(
					MakeBox(&Point3{0, 0, 0}, &Point3{165,165,165}, white),
					-18,
				),
				&Vec3{130,0,65},
			),
		},
	}
	lights = &HittableList{
		[]Hittable{
			MakeRectXZ(213, 227, 343, 332, 554, nil),
		},
	}
	return
}
