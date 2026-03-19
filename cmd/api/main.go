package main

import (
	"falzo/internal/app"
	logger "falzo/pkg"
)

func main() {
	logger.SetupLogger()
	app.Run()
}
