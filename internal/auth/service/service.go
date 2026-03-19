package service

import (
	"falzo/internal/auth"
	"falzo/internal/config"
	"falzo/internal/database"
)

type authService struct {
	cfg config.AuthConfig
	db  database.Client
}

func New(cfg config.AuthConfig, db database.Client) auth.Service {
	return &authService{cfg: cfg, db: db}
}
