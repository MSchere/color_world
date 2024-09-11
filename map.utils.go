package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"time"
)

func GenerateMap() (*MapCache, error) {
	mapCache, err := GetMapCache()
	if err != nil {
		fmt.Println("Map cache key not found! Creating an empty map cache")
		mapCache = &MapCache{}
	}

	if mapCache.Image != nil {
		fmt.Println("Loaded map from cache")
		return mapCache, nil
	}

	fmt.Println("Regenerating map...")
	// If we're here, we need to generate the image

	imgBytes, err := GenerateImage()
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

// TODO: This function is buggy, it changes both the updated pixel and the previously updated ones
func GenerateImage() ([]byte, error) {
	width, height := 1024, 512
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			key := fmt.Sprintf("%d:%d", x, y)
			pixelMutex.RLock()
			r, g, b := hexToRGB(pixelMap[key])
			pixelMutex.RUnlock()
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	fmt.Println("Generated map image")
	return buf.Bytes(), nil
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
