package main

import "os"

type Config struct {
	DatabaseURL string
	ServerPort  string
}

func initConfig() {
	conf = Config{
		ServerPort:  os.Getenv("SERVER_PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
