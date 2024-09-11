package main

import (
	"encoding/base64"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func main() {

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/static/", "./static")
	app.Static("/assets/", "./assets")
	app.Use(logger.New())

	InitPixelRepository()

	go LoadPixels()

	app.Get("/", renderMap)
	app.Post("/update-pixel", updatePixel)

	log.Fatal(app.Listen(":3000"))

}

func renderMap(c *fiber.Ctx) error {
	mapCache, err := GenerateMap()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating map image")
	}

	return c.Render("index", fiber.Map{
		"Title":          "Color World",
		"MapImageBase64": base64.StdEncoding.EncodeToString(mapCache.Image),
		"LastUpdate":     mapCache.LastUpdate.Format(time.RFC3339),
		"Updates":        mapCache.Version,
	})
}

func updatePixel(c *fiber.Ctx) error {
	x, err := strconv.Atoi(c.FormValue("x"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid X coordinate")
	}

	y, err := strconv.Atoi(c.FormValue("y"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Y coordinate")
	}

	color := c.FormValue("color")

	// TODO: Validate color with regex
	if color == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid color")
	}

	err = UpdatePixel(x, y, color)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating pixel")
	}

	mapCache, err = UpdateMapCache()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating map cache")
	}

	return c.Render("appContainer", fiber.Map{
		"MapImageBase64": base64.StdEncoding.EncodeToString(mapCache.Image),
		"LastUpdate":     mapCache.LastUpdate.Format(time.RFC3339),
		"Updates":        mapCache.Version,
	})
}