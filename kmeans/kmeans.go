package kmeans

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

type Mean = []float64
type SegmentData = struct {
	Sum   []uint64
	Count int
}

func Kmeans(k int, img image.Image, means []Mean, prevVariance float64, iter int, maxIter int, tol float64) []Mean {
	if means == nil {
		means = createInitialMeans(k, img)
	}
	availableRoutines := runtime.NumCPU()
	resourcesLock := sync.Mutex{}
	routineReady := make(chan bool)
	variance := 0.0
	segmentMap := make(map[int]*SegmentData)
	for idx := 0; idx < k; idx++ {
		segmentMap[idx] = &SegmentData{Sum: []uint64{0, 0, 0}, Count: 0}
	}
	for x := 0; x < img.Bounds().Max.X; x++ {
		if availableRoutines == 0 {
			<-routineReady
			availableRoutines = availableRoutines + 1
		}
		availableRoutines = availableRoutines - 1
		go func(internalX int) {
			internalSegmentMap := make(map[int]*SegmentData)
			for idx := 0; idx < k; idx++ {
				internalSegmentMap[idx] = &SegmentData{Sum: []uint64{0, 0, 0}, Count: 0}
			}
			internalVariance := 0.0
			for y := 0; y < img.Bounds().Max.Y; y++ {
				dist, segment := FindSegment(img.At(internalX, y), means)
				r, g, b, _ := img.At(internalX, y).RGBA()
				internalSegmentMap[segment].Sum[0] = internalSegmentMap[segment].Sum[0] + uint64(r)
				internalSegmentMap[segment].Sum[1] = internalSegmentMap[segment].Sum[1] + uint64(g)
				internalSegmentMap[segment].Sum[2] = internalSegmentMap[segment].Sum[2] + uint64(b)
				internalSegmentMap[segment].Count = internalSegmentMap[segment].Count + 1
				internalVariance = internalVariance + dist
			}
			resourcesLock.Lock()
			for segmIdx, segmentData := range internalSegmentMap {
				for idx, v := range segmentData.Sum {
					segmentMap[segmIdx].Sum[idx] = segmentMap[segmIdx].Sum[idx] + v
				}
				segmentMap[segmIdx].Count = segmentMap[segmIdx].Count + segmentData.Count
			}
			variance = variance + internalVariance
			routineReady <- true
			resourcesLock.Unlock()
		}(x)
	}
	for availableRoutines < runtime.NumCPU() {
		<-routineReady
		availableRoutines = availableRoutines + 1
	}
	means = computeMeans(segmentMap)
	deviation := math.Abs(prevVariance - variance)
	fmt.Println(variance)
	if deviation < tol || iter >= maxIter {
		return means
	}
	iter++
	return Kmeans(k, img, means, variance, iter, maxIter, tol)
}

func computeMeans(segmentMap map[int]*SegmentData) []Mean {
	res := make([]Mean, len(segmentMap))
	for idx, segmentData := range segmentMap {
		mean := make([]float64, 3)
		for idx := range mean {
			mean[idx] = float64(segmentData.Sum[idx]) / float64(segmentData.Count)
		}
		res[idx] = mean
	}
	return res
}

func FindSegment(color color.Color, means []Mean) (float64, int) {
	var res *float64
	var segment int
	for idx, m := range means {
		dist := computeDistance(color, m)
		if res == nil {
			res = &dist
			segment = idx
		} else if dist < *res {
			segment = idx
			*res = dist
		}
	}
	return *res, segment
}

func computeDistance(color color.Color, mean Mean) float64 {
	r, g, b, _ := color.RGBA()
	return math.Pow(mean[0]-float64(r), 2.0) +
		math.Pow(mean[1]-float64(g), 2.0) +
		math.Pow(mean[2]-float64(b), 2.0)
}

func createInitialMeans(k int, img image.Image) []Mean {
	means := make([]Mean, 0)
	for i := 0; i < k; i++ {
		x := rand.Intn(img.Bounds().Max.X)
		y := rand.Intn(img.Bounds().Max.Y)
		r, g, b, _ := img.At(x, y).RGBA()
		means = append(means, []float64{float64(r), float64(g), float64(b)})
	}
	return means
}
