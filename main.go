package main

import (
	"log/slog"
	"os"

	"github.com/prathoss/cards/internal"
	"github.com/prathoss/cards/pkg"
)

func main() {
	pkg.SetupLogger()
	config, err := internal.NewConfigFromEnv()
	if err != nil {
		slog.Error("could not load config from environment", pkg.Err(err))
		os.Exit(1)
	}
	server, err := internal.NewServer(config)
	if err != nil {
		slog.Error("could not create server", pkg.Err(err))
		os.Exit(1)
	}
	server.Run()
}
