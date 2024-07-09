package main

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"
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

func main() {

	imageExists, _ := exists(os.Args[1])
	if imageExists == false {
		println("Path does not exists")
		os.Exit(0)
	}

	img, _ := getImage(os.Args[1])

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	rgbImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgbImg.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 255})
		}
	}
}
