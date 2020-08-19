package config

import "flag"

var (
	VideoFolder string
)

func Setup() {
	flag.StringVar(&VideoFolder, "f", "/downloads", "Storage folder of video")

	flag.Parse()
}
