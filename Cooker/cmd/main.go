package main

import (
	gen "github.com/mariobenissimo/Cooker/internal/generation"
)

func main() {

	//cooker := models.CreateCooker(":8082")
	//handlers.Start(*cooker)
	gen.Cook("./cooker")
}
