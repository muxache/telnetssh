package model

import "time"

type Config struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Timeout        time.Duration
	TimeoutCommand time.Duration
}
