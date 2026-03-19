package main

import (
	"falzo/internal/app"
	"falzo/internal/logger"
)

func main() {
	logger.SetupLogger()
	app.Run()
}
