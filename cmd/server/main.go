package main

import (
	"flag"
)

var (
	configFile = flag.String("config", "config.yaml", "Configuration file path")
	port       = flag.Int("port", 8080, "Server port")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()
}
