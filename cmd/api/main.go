package main

import (
	"falzo/internal/app"
	"falzo/pkg/logger"
)

func main() {
	logger.SetupLogger()
	app.Run()
}
