package config

import (
	"errors"
	"os"
	"strconv"

	l "github.com/marcelhfm/home_server/pkg/log"
)

var ErrEnvVarEmpty = errors.New("getenv: env empty")

func GetenvStr(key string) string {
	v := os.Getenv(key)

	if v == "" {
		l.Log.Error().Msgf("Env var %s is empty", key)
		os.Exit(1)
	}
	return v
}

func GetenvInt(key string) int {
	s := GetenvStr(key)

	v, err := strconv.Atoi(s)
	if err != nil {
		l.Log.Error().Msgf("Env var %s could not be converted to int. Value %s", key, s)
		os.Exit(1)
	}

	return v
}
