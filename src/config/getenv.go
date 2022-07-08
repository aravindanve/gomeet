package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
)

func GetenvString(name string) string {
	s := os.Getenv(name)
	if s == "" {
		panic(fmt.Sprintf("env variable %s (string) missing", name))
	}
	return s
}

func GetenvStringWithDefault(name string, defaultValue string) string {
	s := os.Getenv(name)
	if s == "" {
		return defaultValue
	}
	return s
}

func GetenvInt(name string) int {
	s := os.Getenv(name)
	if s == "" {
		panic(fmt.Sprintf("env variable %s (int) missing", name))
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("unable to parse env variable %s (int) from %s", name, s))
	}
	return i
}

func GetenvIntWithDefault(name string, defaultValue int) int {
	s := os.Getenv(name)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("unable to parse env variable %s (int) from %s", name, s))
	}
	return i
}

func GetenvBool(name string) bool {
	s := os.Getenv(name)
	if s == "" {
		panic(fmt.Sprintf("env variable %s (int) missing", name))
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(fmt.Sprintf("unable to parse env variable %s (bool) from %s", name, s))
	}
	return b
}

func GetenvBoolWithDefault(name string, defaultValue bool) bool {
	s := os.Getenv(name)
	if s == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(fmt.Sprintf("unable to parse env variable %s (bool) from %s", name, s))
	}
	return b
}

func GetenvStringBase64(name string) []byte {
	s := os.Getenv(name)
	if s == "" {
		panic(fmt.Sprintf("env variable %s (base64 string) missing", name))
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("unable to decode env variable %s (base64 string): %s", name, err.Error()))
	}
	return b
}

func GetenvStringBase64WithDefault(name string, defaultValue []byte) []byte {
	s := os.Getenv(name)
	if s == "" {
		return defaultValue
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(fmt.Sprintf("unable to decode env variable %s (base64 string): %s", name, err.Error()))
	}
	return b
}
