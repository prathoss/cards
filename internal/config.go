package internal

import "os"

type Config struct {
	Address string
}

func NewConfigFromEnv() Config {
	address := os.Getenv("CARDS_ADDRESS")
	if address == "" {
		address = ":8080"
	}
	return Config{
		Address: address,
	}
}
