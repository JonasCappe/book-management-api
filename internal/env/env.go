package env

import (
	"os"
	"strconv"
	"time"
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

func GetTimeDuration(key string) (duration time.Duration) {

	string_value := GetString(key, "15min")

	duration, err := time.ParseDuration(string_value)

	if err != nil {
		panic(err)
	}

	return duration
}
