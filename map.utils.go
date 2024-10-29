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

// GetCirclePerimeter returns a slice of pixels that form the perimeter of a circle
// centered at the given pixel with the specified radius using Bresenham's circle algorithm
func GetCirclePerimeter(center Pixel, radius int) []Pixel {
	pixels := make([]Pixel, 0)

	x := 0            // x-coordinate
	y := radius       // y-coordinate
	d := 3 - 2*radius // Decision parameter

	// Helper function to add a pixel with proper wrapping
	addPixel := func(xOffset, yOffset int) {
		// Calculate actual coordinates relative to center
		actualX := center.X + xOffset
		actualY := center.Y + yOffset

		// Normalize coordinates to handle wrapping
		normalX, normalY := NormalizeCoordinate(actualX, actualY)

		pixels = append(pixels, Pixel{
			X:     normalX,
			Y:     normalY,
			Color: center.Color,
		})
	}

	// Generate the circle points using Bresenham's circle algorithm
	for y >= x {
		// Add points in all octants with wrapping
		addPixel(x, y)
		addPixel(-x, y)
		addPixel(x, -y)
		addPixel(-x, -y)
		addPixel(y, x)
		addPixel(-y, x)
		addPixel(y, -x)
		addPixel(-y, -x)

		// Update x
		x++

		// Update decision parameter and y
		if d > 0 {
			y--
			d = d + 4*(x-y) + 10
		} else {
			d = d + 4*x + 6
		}
	}

	return pixels
}

// GetShortestPath returns a slice of pixels that represent the shortest path
// between the two given pixels, accounting for map wrapping using Bresenham's line algorithm
func GetShortestPath(start, end Pixel) []Pixel {
	path := make([]Pixel, 0)

	// Calculate the initial delta values
	dx := end.X - start.X
	dy := end.Y - start.Y

	// Determine the step direction
	xStep, yStep := 1, 1
	if dx < 0 {
		xStep = -1
	}
	if dy < 0 {
		yStep = -1
	}

	// Normalize the deltas to handle map wrapping
	if abs(dx) > MAP_WIDTH/2 {
		if dx > 0 {
			dx -= MAP_WIDTH
		} else {
			dx += MAP_WIDTH
		}
	}
	if abs(dy) > MAP_HEIGHT/2 {
		if dy > 0 {
			dy -= MAP_HEIGHT
		} else {
			dy += MAP_HEIGHT
		}
	}

	// Initialize Bresenham's algorithm
	x, y := start.X, start.Y
	err := 0
	deltaErr := abs(dy)

	// Add the start pixel
	path = append(path, Pixel{
		X:     x,
		Y:     y,
		Color: start.Color,
	})

	// Draw the line
	for (x != end.X) || (y != end.Y) {
		err += deltaErr
		if err >= dx {
			y += yStep
			err -= dx
		}
		x += xStep

		// Normalize the coordinates to handle map wrapping
		normalX, normalY := NormalizeCoordinate(x, y)
		path = append(path, Pixel{
			X:     normalX,
			Y:     normalY,
			Color: start.Color,
		})
	}

	// Add the end pixel
	path = append(path, Pixel{
		X:     end.X,
		Y:     end.Y,
		Color: end.Color,
	})

	return path
}

// NormalizeCoordinate wraps coordinates around the map edges
func NormalizeCoordinate(x, y int) (int, int) {
	// Wrap x coordinate (longitude)
	x = ((x % MAP_WIDTH) + MAP_WIDTH) % MAP_WIDTH

	// Wrap y coordinate (latitude)
	y = ((y % MAP_HEIGHT) + MAP_HEIGHT) % MAP_HEIGHT

	return x, y
}

// abs returns the absolute value of the given integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
