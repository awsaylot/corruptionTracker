package config

import "clankClient/internal/version"

const (
	DefaultServerURL = "http://localhost:8080"
	ClientName       = version.Name
	ClientVersion    = version.Version
)
