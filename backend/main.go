package main

import (
	"backend/api"
	"backend/config"
	_ "backend/logger"
)

var magnetLink = "magnet:?xt=urn:btih:0627dec6bdcf98ad4bd1f3cfce1564df63acf695&dn=Alita.Battle.Angel.2019.1080p.BluRay.x264-SPARKS&tr=http%3A%2F%2Ftracker.trackerfix.com%3A80%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2860&tr=udp%3A%2F%2F9.rarbg.to%3A2780"

func main() {
	config.LoadConfig()
	api.StartServer()
}
