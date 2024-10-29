package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png" // Import PNG decoder
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func rgbToHex(r, g, b uint32) string {
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

func populateRedisFromImage(imagePath string, redisHost string, redisPort int) error {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisHost, redisPort),
	})

	// Open the image
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("error opening image: %v", err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}

	// Check image dimensions
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	if width != 1024 || height != 512 {
		return fmt.Errorf("image dimensions must be 1024x512, but got %dx%d", width, height)
	}

	cnt := 0
	totalPixels := width * height
	// Iterate over each pixel
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// Get the RGB value of the pixel
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert RGB to hex color code
			hexColor := rgbToHex(r, g, b)
			if hexColor != "#6abe30" && hexColor != "#5b6ee1" { // Skip if the color is not green or blue
				continue
			}

			// Create the key in the format "X:Y"
			key := fmt.Sprintf("%d:%d", x, y)

			// Store in Redis
			err := rdb.Set(context.Background(), key, hexColor, 0).Err()
			cnt++
			fmt.Printf("\rLoading map pixels... %d/%d", cnt, totalPixels)
			if err != nil {
				return fmt.Errorf("error setting value in Redis: %v", err)
			}
		}
	}

	fmt.Printf(" - Done!\n")
	return nil
}

func clearMapCache(redisHost string, redisPort int) error {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisHost, redisPort),
	})
	type MapCache struct {
		Image      []byte    `json:"image"`
		Version    int64     `json:"version"`
		LastUpdate time.Time `json:"last_update"`
	}

	newCache := &MapCache{
		Image:      nil,
		Version:    1,
		LastUpdate: time.Now(),
	}

	data, err := json.Marshal(newCache)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = rdb.Set(context.Background(), "map", data, 0).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func main() {
	imagePath := "map.png" // Replace with your image path
	redisHost := "192.168.1.145"
	redisPort := 6379

	err := populateRedisFromImage(imagePath, redisHost, redisPort)
	err = clearMapCache(redisHost, redisPort)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
