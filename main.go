package main

import (
	"github.com/prathoss/cards/internal"
	"github.com/prathoss/cards/pkg"
)

func main() {
	pkg.SetupLogger()
	config := internal.NewConfigFromEnv()
	server := internal.NewServer(config)
	server.Run()
}
