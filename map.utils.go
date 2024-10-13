package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"time"
)

func LoadMap() (*MapCache, error) {
	mapCache := GetMapCache()
	if mapCache.Image != nil {
		fmt.Println("Loaded map from cache")
		return mapCache, nil
	}
	// If we're here, we need to generate the image
	return RegenerateMap()
}

func RegenerateMap() (*MapCache, error) {
	fmt.Println("Regenerating map...")
	mapCache := GetMapCache()

	pixels, err := GetPixels()
	if err != nil {
		return nil, err
	}

	imgBytes, err := GenerateImage(pixels)
	if err != nil {
		return nil, err
	}
	mapCache.Image = imgBytes
	mapCache.LastUpdate = time.Now()

	if err := SetMapCache(mapCache); err != nil {
		return nil, err
	}

	return mapCache, nil
}

func GenerateImage(pixels []Pixel) ([]byte, error) {
	width, height := 1024*PIXEL_SIZE, 512*PIXEL_SIZE
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for _, pixel := range pixels {
		r, g, b := hexToRGB(pixel.Color)
		for x := 0; x < PIXEL_SIZE; x++ {
			for y := 0; y < PIXEL_SIZE; y++ {
				rgba.Set(pixel.X*PIXEL_SIZE+x, pixel.Y*PIXEL_SIZE+y, color.RGBA{r, g, b, 255})
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
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
	rgba := img.(*image.RGBA)

	r, g, b := hexToRGB(newPixel.Color)
	for x := 0; x < PIXEL_SIZE; x++ {
		for y := 0; y < PIXEL_SIZE; y++ {
			rgba.Set(newPixel.X*PIXEL_SIZE+x, newPixel.Y*PIXEL_SIZE+y, color.RGBA{r, g, b, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		fmt.Println("Error encoding image")
		return nil, err
	}

	return buf.Bytes(), nil
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
