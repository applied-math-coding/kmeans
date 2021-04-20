package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"main/kmeans"
	"math"
	"os"
	"runtime"
	"sort"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	imageFile, errOpen := os.Open("cloud.jpg")
	if errOpen != nil {
		panic(errOpen)
	}
	defer imageFile.Close()
	image, errDecode := jpeg.Decode(imageFile)
	if errDecode != nil {
		panic(errDecode)
	}
	fmt.Println(image.Bounds())

	means := kmeans.Kmeans(10, image, nil, 0.0, 0, 10, 0.5)
	// means := hist(image, 10)  // this is a smarter alternative to kmeans for this specific case

	transformAndSave(image, means)
	fmt.Println(means)
}

func hist(img image.Image, k int) []kmeans.Mean {
	type RGB = struct {
		r uint32
		g uint32
		b uint32
	}
	rgbs := make([]RGB, 0)
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgbs = append(rgbs, RGB{r, g, b})
		}
	}
	sort.Slice(rgbs, func(i int, j int) bool {
		if rgbs[i].r == rgbs[j].r && rgbs[i].g == rgbs[j].g {
			return rgbs[i].b < rgbs[j].b
		} else if rgbs[i].r == rgbs[j].r {
			return rgbs[i].g < rgbs[j].g
		} else {
			return rgbs[i].r < rgbs[j].r
		}
	})
	means := make([]kmeans.Mean, 0)
	d := int(math.Floor(float64(len(rgbs)) / float64(k)))
	for idx := 0; idx < k; idx++ {
		means = append(means, kmeans.Mean{float64(rgbs[d*idx].r), float64(rgbs[d*idx].g), float64(rgbs[d*idx].b)})
	}
	return means
}

func transformAndSave(img image.Image, means []kmeans.Mean) {
	fmt.Println("transforming image")
	newImg := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y))
	outFile, errCreate := os.Create("cloud_transf.jpg")
	if errCreate != nil {
		panic(errCreate)
	}
	defer outFile.Close()
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			_, segment := kmeans.FindSegment(img.At(x, y), means)
			r := uint8(uint32(math.Round(means[segment][0])) >> 8)
			g := uint8(uint32(math.Round(means[segment][1])) >> 8)
			b := uint8(uint32(math.Round(means[segment][2])) >> 8)
			newImg.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	png.Encode(outFile, newImg)
}
