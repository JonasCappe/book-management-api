package env

import (
	"os"
	"strconv"
)


func GetString(key, fallback string) (value string) {
	value, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return value
}

func GetInt(key string, fallback int) (value int) {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	value, err := strconv.Atoi(val)
	
	if err != nil {
		return fallback
	}

	return value

}