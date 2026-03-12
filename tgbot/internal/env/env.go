package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var loaded bool = false

func Init() {
	if loaded {
		return
	}

	if _, exists := os.LookupEnv("DOTENV"); !exists{
		loaded = true
		return
	}

	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	loaded = true
}

func GetEnvString(key string) string {
	if !loaded {
		Init()
	}

	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic(fmt.Sprintf("No string env var for key %s", key))
}

func GetEnvInt(key string) int {
	if !loaded {
		Init()
	}

	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	panic(fmt.Sprintf("No int env var for key %s", key))
}
