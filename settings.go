package main

import (
	"errors"
	"os"
)

const FileName = "./apikey"
const INVALID_KEY = " Found SteamAPI Key is invalid"

func MakeAPIKeyFile() {

	os.Create(FileName)

}

type APIKey struct {
	Key string
}

func GetAPIKey() (*APIKey, error) {

	_, err := os.Stat(FileName)

	if err != nil {
		MakeAPIKeyFile()
	}

	data, err := os.ReadFile(FileName)
	if err != nil {
		return nil, err
	}

	Key := string(data)

	if Key == "" {
		return nil, errors.New(INVALID_KEY)
	}

	return &APIKey{
		Key: Key,
	}, nil

}
