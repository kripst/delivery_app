package logger

import "go.uber.org/zap"

func NewLogger() (*zap.Logger, error) {
	zapCfg := zap.NewDevelopmentConfig()

	zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapCfg.OutputPaths = []string{
		// включить чтобы в папку записывались "logs/order_service.logs",
		// os.Getenv("LOGSPATH"),
		"stdout",
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return  logger, nil
}