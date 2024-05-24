package internal

import (
	"fmt"
	"os"
)

type Config struct {
	Address         string
	MongoConnection string
}

func NewConfigFromEnv() (Config, error) {
	address := os.Getenv("CARDS_ADDRESS")
	if address == "" {
		address = ":8080"
	}

	const mongoConnectionEnvVar = "CARDS_MONGO_CONN_STR"
	mongoConnection := os.Getenv(mongoConnectionEnvVar)
	if mongoConnection == "" {
		return Config{}, fmt.Errorf("%s environment variable is not set", mongoConnectionEnvVar)
	}
	return Config{
		Address:         address,
		MongoConnection: mongoConnection,
	}, nil
}
