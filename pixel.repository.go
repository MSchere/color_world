package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/storage/redis/v3"
)

type Pixel struct {
	X     int
	Y     int
	Color string
}

type MapCache struct {
	Image      []byte    `json:"image"`
	Version    int64     `json:"version"`
	LastUpdate time.Time `json:"last_update"`
}

const (
	SEA_COLOR  = "#5b6ee1"
	LAND_COLOR = "" // TODO: Fill land color variable
	PIXEL_SIZE = 5
	MAP_WIDTH  = 1024
	MAP_HEIGHT = 512
	SEA_RANGE  = 500
)

var (
	rdb *redis.Storage
)

func InitPixelRepository() {
	rdb = RedisConnection()
}

func GetPixels() ([]Pixel, error) {
	rdb := RedisConnection()
	width, height := MAP_WIDTH, MAP_HEIGHT
	totalPixels := width * height

	pixels := make([]Pixel, 0, totalPixels)
	var wg sync.WaitGroup                      // Wait group to wait for all goroutines to finish
	pixelChan := make(chan Pixel, totalPixels) // Channel to send pixels
	semaphore := make(chan int, 100)           // Limit concurrent goroutines

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			wg.Add(1)           // Increment wait group counter for each goroutine
			go func(x, y int) { // Goroutine to fetch pixel from Redis and send it to channel when done
				semaphore <- 1                 // Acquire semaphore to limit concurrent goroutines
				defer wg.Done()                // Decrement wait group counter when the goroutine finishes
				defer func() { <-semaphore }() // Release semaphore when the goroutine finishes

				key := fmt.Sprintf("%d:%d", x, y)
				color, err := rdb.Get(key)
				if err != nil || color == nil {
					color = []byte(SEA_COLOR)
				}
				colorStr := string(color)
				pixelChan <- Pixel{X: x, Y: y, Color: colorStr} // Send pixel to channel
				fmt.Printf("\rLoading map pixels... %d/%d", len(pixels), totalPixels)
			}(x, y)
		}
	}

	go func() {
		wg.Wait()        // Wait for all goroutines to finish
		close(pixelChan) // Close the channel
		fmt.Printf(" - Done!\n")
	}()

	for pixel := range pixelChan {
		pixels = append(pixels, pixel)
	}

	return pixels, nil
}

func UpdatePixel(newPixel Pixel) error {
	valid, errStr := isOwnPositionValid(newPixel)
	if !valid {
		return fmt.Errorf(errStr)
	}

	valid, errStr = isNeighbourPositionValid(newPixel)
	if !valid {
		return fmt.Errorf(errStr)
	}

	err := rdb.Set(fmt.Sprintf("%d:%d", newPixel.X, newPixel.Y), []byte(newPixel.Color), 0)
	if err != nil {
		return fmt.Errorf("Failed to update pixel in database")
	}

	fmt.Printf("Updated in-memory pixel %d:%d to %s\n", newPixel.X, newPixel.Y, newPixel.Color)
	return nil
}

func isOwnPositionValid(p Pixel) (bool, string) {
	existingPixelColor, err := rdb.Get(fmt.Sprintf("%d:%d", p.X, p.Y))
	if err != nil {
		return false, "Invalid pixel position, pixel does not exist"
	}
	if p.Color == SEA_COLOR {
		return false, "Invalid pixel position, cannot update sea pixels"
	}
	if p.Color == string(existingPixelColor) {
		return false, "Invalid pixel position, new color must be different from the current one"
	}
	return true, ""
}

// TODO: Complete this funcito
func isNeighbourPositionValid(p Pixel) (bool, string) {
	// check that any of the 8 surrounding pixels is the same color or that it is adjacent to the sea
	for x := p.X - 1; x <= p.X+1; x++ {
		for y := p.Y - 1; y <= p.Y+1; y++ {
			if x == p.X && y == p.Y { // Skip the pixel itself
				continue
			}
			neighbourColor, err := rdb.Get(fmt.Sprintf("%d:%d", x, y))

			if err != nil { // edge of the map
				// check at which of the 4 edges we are and get the new neighbour
				if x < 0 {
					neighbourColor, _ = rdb.Get(fmt.Sprintf("%d:%d", MAP_WIDTH, y))
				}
				if x > MAP_WIDTH {
					neighbourColor, _ = rdb.Get(fmt.Sprintf("%d:%d", 0, y))
				}
				if y < 0 {
					neighbourColor, _ = rdb.Get(fmt.Sprintf("%d:%d", x, MAP_HEIGHT))
				}
				if y > MAP_HEIGHT {
					neighbourColor, _ = rdb.Get(fmt.Sprintf("%d:%d", x, 0))
				}
			}

			if string(neighbourColor) == SEA_COLOR { // neighbouring the sea
				// check if there are any pixels at a SEA_RANGE pixel radious that are neighbouring the sea
			}

			if string(neighbourColor) == p.Color { // neighbouring a land pixel of the same color
				return true, ""
			}
		}
	}
	return false, ""
}

func GetMapCache() *MapCache {
	val, err := rdb.Get("map")
	var cache *MapCache = &MapCache{} // Default empty cache
	if err != nil {
		fmt.Println(err)
		return cache // Return empty cache
	}

	err = json.Unmarshal(val, &cache)
	return cache
}

func UpdateMapCache(newPixel Pixel) (*MapCache, error) {
	cache := GetMapCache()
	mapImage, err := UpdateImage(cache.Image, newPixel)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	newCache := &MapCache{
		Image:      mapImage,
		Version:    cache.Version + 1,
		LastUpdate: time.Now(),
	}
	cache.LastUpdate = time.Now()

	err = SetMapCache(newCache)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Updated map cache")
	return newCache, nil
}

func SetMapCache(cache *MapCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return rdb.Set("map", data, 0)
}
