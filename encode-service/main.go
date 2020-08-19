package main

import (
	"encode-service/api"
	"encode-service/config"
	_ "encode-service/logger"
)

func main() {
	config.Setup()
	api.StartServer()
}
