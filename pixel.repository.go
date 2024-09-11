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

var (
	rdb        *redis.Storage
	pixelMap   = make(map[string]string)
	mapCache   = &MapCache{}
	pixelMutex sync.RWMutex
)

func InitPixelRepository() {
	rdb = RedisConnection()
}

func LoadPixels() {
	width, height := 1024, 512
	totalPixels := width * height

	var wg sync.WaitGroup                  // Wait for all goroutines to finish
	semaphore := make(chan struct{}, 1000) // Limit concurrent goroutines

	fmt.Println("Loading map pixels...")
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			wg.Add(1) // Increment the WaitGroup counter
			go func(x, y int) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire semaphore
				defer func() { <-semaphore }() // Release semaphore

				key := fmt.Sprintf("%d:%d", x, y)
				color, err := rdb.Get(key)
				if err != nil {
					fmt.Println(err)
					return
				}

				pixelMutex.Lock()
				pixelMap[key] = string(color)
				pixelMutex.Unlock()

				fmt.Printf("\rLoading map pixels in memory... %d/%d", len(pixelMap), totalPixels)
			}(x, y)
		}
	}
	wg.Wait()
	fmt.Printf(" - Done!\n")
}

func UpdatePixel(x int, y int, color string) error {
	key := fmt.Sprintf("%d:%d", x, y)
	err := rdb.Set(key, []byte(color), 0)
	if err != nil {
		return err
	}

	pixelMutex.Lock()
	pixelMap[key] = color
	pixelMutex.Unlock()

	fmt.Printf("Updated in-memory pixel %d:%d to %s\n", x, y, pixelMap[key])
	return nil
}

func GetMapCache() (*MapCache, error) {
	val, err := rdb.Get("map")
	if err != nil {
		return nil, err
	}

	var cache MapCache
	err = json.Unmarshal(val, &cache)
	return &cache, err
}

func UpdateMapCache() (*MapCache, error) {
	cache, err := GetMapCache()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	mapImage, err := GenerateImage()
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
