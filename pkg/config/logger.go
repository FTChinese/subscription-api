package config

import (
	"go.uber.org/zap"
	"log"
)

func MustGetLogger(prod bool) *zap.Logger {
	var logger *zap.Logger
	var err error
	if prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatal(err)
	}

	return logger
}
