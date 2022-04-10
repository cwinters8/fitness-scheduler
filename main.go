package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"fitness-scheduler/scheduler"
	"fitness-scheduler/sessions"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setup() error {
	godotenv.Load()

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		return errors.New("failed to connect to database: " + err.Error())
	}

	app := fiber.New()
	app.Use(logger.New())

	apiURL := os.Getenv("API_URL")
	err = scheduler.Init(apiURL, db)
	if err != nil {
		return errors.New("failed to initialize scheduler: " + err.Error())
	}

	app.Post("/session", func(c *fiber.Ctx) error {
		var session sessions.Session
		err := c.BodyParser(&session)
		if err != nil {
			msg := "failed to parse session: " + err.Error()
			log.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		err = session.Save(db)
		if err != nil {
			msg := "failed to save session: " + err.Error()
			log.Println(msg)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": msg,
			})
		}
		return c.JSON(session)
	})

	app.Post("/notify", func(c *fiber.Ctx) error {
		var session sessions.Session
		err := c.BodyParser(&session)
		if err != nil {
			msg := "failed to parse session: " + err.Error()
			log.Println(msg)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": msg,
			})
		}
		msg := fmt.Sprintf("Notify user %d of session %d with title %s", session.UserID, session.ID, session.Title)
		log.Println(msg)
		return c.JSON(fiber.Map{
			"msg": msg,
		})
	})

	return app.Listen(":8000")
}

func main() {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
}
