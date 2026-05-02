package main

import (
	"log"

	"cdek_wishlist/internal/app"
	"cdek_wishlist/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	if err := app.Run(config.Load()); err != nil {
		log.Fatal(err)
	}
}
