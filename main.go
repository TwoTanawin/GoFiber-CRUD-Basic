package main

import (
	"log"
	"os"
	"time"

	_ "fiber-test/docs"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/html/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books []Book

func checkMiddleWare(c *fiber.Ctx) error {
	// start := time.Now()

	// fmt.Printf("URL = %s, Method = %s, Time = %s\n",
	// 	c.OriginalURL(), c.Method(), start,
	// )

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	// name := claims["name"].(string)

	if claims["role"] != "admin" {
		return fiber.ErrUnauthorized
	}

	return c.Next()
}

// @title Book API
// @description This is a sample server for a book API.
// @version 1.0
// @host localhost:8082
// @BasePath /
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("load .env error")
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/swagger/*", swagger.HandlerDefault)

	books = append(books, Book{ID: 1, Title: "KMUTNB", Author: "Two"})
	books = append(books, Book{ID: 2, Title: "AIT", Author: "TwoISE"})

	// print(books)

	app.Post("/login", login)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))

	app.Use(checkMiddleWare)

	app.Get("/books", getBooks)

	app.Get("/books/:id", getBook)

	app.Post("/books", createBook)

	app.Put("/books/:id", updateBook)

	app.Delete("/books/:id", deleteBook)

	app.Post("/upload", uploadFile)

	app.Get("/test-html", testHtml)

	app.Get("/config", getEnv)

	if err := app.Listen(":8082"); err != nil {
		panic(err)
	}
}

func uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("image")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err = c.SaveFile(file, "./uploads/"+file.Filename)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("File upload complete!")
}

func testHtml(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Hello World!",
	})
}

func getEnv(c *fiber.Ctx) error {
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "defaultsecret"
	}

	return c.JSON(fiber.Map{
		"SECRET": secret,
	})
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var memberUser = User{
	Email:    "test@test.com",
	Password: "pass123",
}

func login(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if user.Email != memberUser.Email || user.Password != memberUser.Password {
		return fiber.ErrUnauthorized
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["role"] = "admin"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"message": "Login Success",
		"token":   t,
	})
}
