package logger

import (
	"log"

	"go.uber.org/zap"
)

func MustInitializedLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	sugar := *logger.Sugar()
	return &sugar
}
