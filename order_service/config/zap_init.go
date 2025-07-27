package config

import (
	"log"

	"go.uber.org/zap"
)

func InitLoggerZap() {
	config := zap.NewDevelopmentConfig()
	
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.OutputPaths = []string{
		// включить чтобы в папку записывались "logs/order_service.logs",
		// os.Getenv("LOGSPATH"),
		"stdout",
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Не удалось проинициализировать логгер зап: %v", err)
	}

	zap.ReplaceGlobals(logger)
}
