package main

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path"
	"strings"

	"github.com/esimov/stackblur-go"
)

func getImage(imagePath string) (img image.Image, err error) {
	file, err := os.Open(imagePath)
	if err != nil {
		err = fmt.Errorf("Failure: %s", err)
		return
	}

	defer file.Close()

	mimeType := path.Ext(imagePath)
	mimeType = strings.TrimPrefix(mimeType, ".")

	switch mimeType {
	case "jpeg", "jpg":
		img, err = jpeg.Decode(file)

	case "png":
		img, err = png.Decode(file)
	}

	return
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func randomString() string {
	b := make([]byte, 11)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2:11]
}

func save_image(img image.Image) (cache_path string, err error) {
	cache_path = fmt.Sprintf("/tmp/%s.png", randomString())

	f1, err := os.Create(cache_path)
	if err != nil {
		err = fmt.Errorf("failed to create file: %v", err)
	}
	defer f1.Close()

	if err = png.Encode(f1, img); err != nil {
		err = fmt.Errorf("failed to encode: %v", err)
	}

	return
}

func sliceImage(img image.Image, grid int) (tiles []image.Image) {
	tiles = make([]image.Image, 0, grid*grid)

	if cap(tiles) == 0 {
		return
	}

	shape := img.Bounds()

	fheight := float64(shape.Max.Y / int(grid))
	fwidth := float64(shape.Max.X / int(grid))

	height := int(math.Ceil(fheight))
	width := int(math.Ceil(fwidth))

	for y := shape.Min.Y; y+height <= shape.Max.Y; y += height {

		for x := shape.Min.X; x+width <= shape.Max.X; x += width {

			tile := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, x+width, y+height))

			tiles = append(tiles, tile)
		}
	}

	return
}

func main() {

	imageExists, _ := exists(os.Args[1])
	if imageExists == false {
		println("Path does not exists")
		os.Exit(0)
	}

	img, _ := getImage(os.Args[1])

	blurred_img, _ := stackblur.Process(img, 30)

	tiles := sliceImage(blurred_img, 4)

	for _, tile := range tiles {
		colors := make(map[color.RGBA]int)
		bounds := tile.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				col := tile.At(x, y)
				r, g, b, a := col.RGBA()
				rgbaCol := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
				colors[rgbaCol]++

			}
		}
		var mostFrequentColor color.RGBA
		maxCount := 0
		for col, count := range colors {
			if count > maxCount {
				maxCount = count
				mostFrequentColor = col
			}
		}
		hexColor := fmt.Sprintf("#%02x%02x%02x", mostFrequentColor.R, mostFrequentColor.G, mostFrequentColor.B)
		fmt.Println(mostFrequentColor, hexColor)
	}

}
