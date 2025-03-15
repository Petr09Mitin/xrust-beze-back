package main

import (
	"github.com/Petr09Mitin/xrust-beze-back/internal/router"
)

func main() {
	c := router.NewChat()
	err := c.Start()
	if err != nil {
		panic(err)
	}
}
