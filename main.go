package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {

	app := fiber.New(fiber.Config{
		Prefork: false,
	})

	// Middleware จะทำงานแบบ pipeline
	// ให้ middleware ทำงานทุก path
	/*
		app.Use(func(c *fiber.Ctx) error {
			// จะทำงานก่อนและหลังเรียกแต่ละ path
			fmt.Println("before next")
			c.Next()
			fmt.Println("after next")
			return nil
		})
	*/
	// กำหนดแค่บาง path
	app.Use("/error", func(c *fiber.Ctx) error {
		fmt.Println("middleware")
		err := c.Next()
		fmt.Println("end middleware")
		return err
	})

	// ถ้าต้องการส่งค่าที่อยู่ใน Middleware ให้ path ต่อไป
	app.Use("/getmiddleware", func(c *fiber.Ctx) error {
		c.Locals("id", 1)
		c.Locals("name", "mid Man")
		err := c.Next()
		fmt.Println("end middleware")
		return err
	})

	// End middleware

	app.Use(requestid.New(requestid.Config{
		// Header: ,
	}))
	app.Use(cors.New(cors.Config{
		// AllowOrigins: "*",
		// AllowMethods: "GET",
		// AllowHeaders: "*",
	}))

	app.Use(logger.New(logger.Config{
		TimeZone: "Asia/Bangkok",
	}))

	app.Get("index", func(c *fiber.Ctx) error {
		return c.SendString("GET: index")
	})

	app.Post("index", func(c *fiber.Ctx) error {
		return c.SendString("POST: index")
	})

	app.Get("/getmiddleware", func(c *fiber.Ctx) error {
		// pass ค่าตัวแปรที่ได้จาก middleware
		id := c.Locals("id")
		name := c.Locals("name")
		return c.SendString(fmt.Sprintf("hello %v , %v", id, name))
		// hello 1 , mid Man
	})

	// parameters
	app.Get("/index/params/:say", func(c *fiber.Ctx) error {
		say := c.Params("say")
		return c.SendString("say: " + say)
	})

	// parameters optional
	app.Get("/index/params/:say/:say2?", func(c *fiber.Ctx) error {
		say := c.Params("say")
		say2 := c.Params("say2")
		return c.SendString("say: " + say + ",\n say2: " + say2)
	})

	// paramsInt
	app.Get("/index/paint/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return fiber.ErrBadRequest
		}
		return c.SendString(fmt.Sprintf("ID: %v", id))
	})

	// QueryString
	app.Get("/index/qry", func(c *fiber.Ctx) error {
		qryString := c.Query("qry")
		return c.SendString("qry: " + qryString)
	})

	// QueryParser
	app.Get("/qrypar", func(c *fiber.Ctx) error {
		person := Person{}
		c.QueryParser(&person)
		return c.JSON(person)
		// http://localhost:8000/qrypar?id=1&name=som
	})

	// Wildcards
	app.Get("/wildcards/*", func(c *fiber.Ctx) error {
		wildcards := c.Params("*")
		return c.SendString(wildcards)
		/*
			http://localhost:8000/wildcards/a/b/c/d/e/1
			a/b/c/d/e/1
		*/
	})

	// Static file
	app.Static("/", "./wwwroot", fiber.Static{
		Index:         "index.html",
		CacheDuration: time.Second * 10,
	})

	// NewError
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "content not found")
	})

	// Group เพื่อแยก version
	// เพิ่ม Middleware ได้
	v1 := app.Group("/v1", func(c *fiber.Ctx) error {
		// Set หรือ Get Header
		c.Set("Version", "v1")
		return c.Next()
	})
	v1.Get("/index", func(c *fiber.Ctx) error {
		return c.SendString("Index v1")
		// http://localhost:8000/v1/index
	})

	v2 := app.Group("/v2", func(c *fiber.Ctx) error {
		c.Set("Version", "v2")
		return c.Next()
	})
	v2.Get("/index", func(c *fiber.Ctx) error {
		return c.SendString("Index v2")
	})

	// Mount จะแยกการทำงานของ app.fiber ออกไปอีกชุด
	userApp := fiber.New()
	userApp.Get("/login", func(c *fiber.Ctx) error {
		return c.SendString("login")
	})
	// ให้ userApp handdle /user ให้
	app.Mount("/user", userApp)
	// http://localhost:8000/user/login

	// Server
	// Config
	app.Server().MaxConnsPerIP = 1
	app.Get("/server", func(c *fiber.Ctx) error {
		time.Sleep(time.Second * 10)
		return c.SendString("server")
	})

	// Environment
	app.Get("/env", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"BaseURL":     c.BaseURL(),
			"Hostname":    c.Hostname(),
			"IP":          c.IP(),
			"IPs":         c.IPs(),
			"OriginalURL": c.OriginalURL(),
			"Path":        c.Path(),
			"Protocol":    c.Protocol(),
			"Subdomains":  c.Subdomains(),
		})
	})
	/*
			{
		    "BaseURL": "http://localhost:8000",
		    "Hostname": "localhost:8000",
		    "IP": "127.0.0.1",
		    "IPs": [],
		    "OriginalURL": "/env",
		    "Path": "/env",
		    "Protocol": "http",
		    "Subdomains": [
		        "localhost:8000"
		    ]
		}
	*/

	/*
			http://localhost:8000/env?id=123
			{
		    "BaseURL": "http://localhost:8000",
		    "Hostname": "localhost:8000",
		    "IP": "127.0.0.1",
		    "IPs": [],
		    "OriginalURL": "/env?id=123",
		    "Path": "/env",
		    "Protocol": "http",
		    "Subdomains": [
		        "localhost:8000"
		    ]
		}

	*/

	// Body
	// ส่ง body ใน POST
	app.Post("/body", func(c *fiber.Ctx) error {
		// เช็คว่า body เป็น json
		fmt.Printf("Is json %v\n", c.Is("json"))
		fmt.Println(string(c.Body()))
		return nil
	})

	// อ่านค่า json ที่ส่งมาและแปลงเป็น struct
	app.Post("/bodytostruct", func(c *fiber.Ctx) error {
		// เช็คว่า body เป็น json
		fmt.Printf("Is json %v\n", c.Is("json"))
		fmt.Println(string(c.Body()))
		person := Person{}
		err := c.BodyParser(&person)
		if err != nil {
			return err
		}
		fmt.Println(person)
		return nil

	})
	// อ่านค่า json ที่ส่งมาและแปลงเป็น map
	app.Post("/bodytomap", func(c *fiber.Ctx) error {
		// เช็คว่า body เป็น json
		fmt.Printf("Is json %v\n", c.Is("json"))

		// ใช้ interface เพื่อให้รับค่าทุก type
		data := map[string]interface{}{}
		err := c.BodyParser(&data)
		if err != nil {
			return err
		}
		fmt.Println(data)
		return nil

	})

	app.Listen(":8000")

}

type Person struct {
	Id   int    `json:id`
	Name string `json:name`
}
