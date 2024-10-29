package main

import (
	"encoding/base64"
	"log"
	"regexp"
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

	RegenerateMap()

	app.Get("/", renderMap)
	app.Post("/update-pixel", updatePixel)

	log.Fatal(app.Listen(":3000"))

}

func renderMap(c *fiber.Ctx) error {
	mapCache, err := LoadMap()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating map image")
	}

	return c.Render("index", fiber.Map{
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
	posX := x / PIXEL_SIZE

	y, err := strconv.Atoi(c.FormValue("y"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Y coordinate")
	}
	posY := y / PIXEL_SIZE

	color := c.FormValue("color")

	match, err := regexp.MatchString("^#[0-9a-fA-F]{6}$", color)
	if !match || err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid color")
	}

	err = UpdatePixel(Pixel{posX, posY, color})
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
	}

	mapCache, err := UpdateMapCache(Pixel{posX, posY, color})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating map cache")
	}
	return c.Render("mapContainer", fiber.Map{
		"MapImageBase64": base64.StdEncoding.EncodeToString(mapCache.Image),
		"LastUpdate":     mapCache.LastUpdate.Format(time.RFC3339),
		"Updates":        mapCache.Version,
	})
	// TODO: don't loose color picker value on rerender
}
