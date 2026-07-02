package main

import (
	"be/internal/config"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatal("load config failed:", err)
	}

	fmt.Println("Server port:", cfg.Server.Port)
	fmt.Println("Database host:", cfg.Database.Host)
	fmt.Println("Redis host:", cfg.Redis.Host)
}
