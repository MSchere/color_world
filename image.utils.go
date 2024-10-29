package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

func GenerateImage(pixels []Pixel) ([]byte, error) {
	width, height := MAP_WIDTH*PIXEL_SIZE, MAP_HEIGHT*PIXEL_SIZE
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for _, pixel := range pixels {
		rgba = drawPixel(rgba, pixel)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		fmt.Println("Error encoding image")
		return nil, err
	}
	fmt.Println("Generated map image")
	return buf.Bytes(), nil
}

func UpdateImage(imageToUpdate []byte, newPixel Pixel) ([]byte, error) {
	if imageToUpdate == nil {
		return nil, fmt.Errorf("Image to update is nil")
	}
	img, _, _ := image.Decode(bytes.NewReader(imageToUpdate))
	rgba := drawPixel(img.(*image.RGBA), newPixel)

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		fmt.Println("Error encoding image")
		return nil, err
	}

	return buf.Bytes(), nil
}

func drawPixel(rgba *image.RGBA, newPixel Pixel) *image.RGBA {
	r, g, b := hexToRGB(newPixel.Color)
	for x := 0; x < PIXEL_SIZE; x++ {
		for y := 0; y < PIXEL_SIZE; y++ {
			rgba.Set(newPixel.X*PIXEL_SIZE+x, newPixel.Y*PIXEL_SIZE+y, color.RGBA{r, g, b, 255})
		}
	}
	return rgba
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
