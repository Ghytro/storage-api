package config

import "os"

var DatabaseURL = os.Getenv("DATABASE_URL")
