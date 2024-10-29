package main

import (
	"fmt"
	"time"
)

// LoadMap loads the map from cache if it exists, otherwise it regenerates the map
func LoadMap() (*MapCache, error) {
	mapCache := GetMapCache()
	if mapCache.Image != nil {
		fmt.Println("Loaded map from cache")
		return mapCache, nil
	}
	// If we're here, we need to generate the image
	return RegenerateMap()
}

// RegenerateMap regenerates the map and updates the cache
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

// GetCircleFill returns a slice of pixels that form a circle
// centered at a given pixel with the specified radius using Bresenham's circle algorithm
func GetCircleFill(center Pixel, radius int) []Pixel {
	pixels := make([]Pixel, 0)

	// Helper function to add a pixel with proper wrapping
	addPixel := func(x, y int) {
		// Calculate actual coordinates relative to center
		actualX := center.X + x
		actualY := center.Y + y

		// Normalize coordinates to handle wrapping
		normalX, normalY := NormalizeCoordinate(actualX, actualY)

		pixels = append(pixels, Pixel{
			X:     normalX,
			Y:     normalY,
			Color: center.Color,
		})
	}

	// Helper function to fill a horizontal line between two x-coordinates
	fillLine := func(x1, x2, y int) {
		// Ensure x1 <= x2
		if x1 > x2 {
			x1, x2 = x2, x1
		}

		// Fill all pixels in the horizontal line
		for x := x1; x <= x2; x++ {
			addPixel(x, y)
		}
	}

	x := 0
	y := radius
	d := 3 - 2*radius

	// Keep track of the last point in each octant
	for y >= x {
		// Fill horizontal lines in each octant
		fillLine(-x, x, y) // Top and bottom
		fillLine(-x, x, -y)

		fillLine(-y, y, x) // Left and right
		fillLine(-y, y, -x)

		x++

		if d > 0 {
			y--
			d = d + 4*(x-y) + 10
		} else {
			d = d + 4*x + 6
		}
	}

	return pixels
}

// NormalizeCoordinate wraps coordinates around the map edges
func NormalizeCoordinate(x, y int) (int, int) {
	// Wrap x coordinate (longitude)
	x = ((x % MAP_WIDTH) + MAP_WIDTH) % MAP_WIDTH

	// Wrap y coordinate (latitude)
	y = ((y % MAP_HEIGHT) + MAP_HEIGHT) % MAP_HEIGHT

	return x, y
}
