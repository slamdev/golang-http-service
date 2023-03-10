package integration

import (
	"fmt"
	"go.uber.org/zap"
	"sync/atomic"
)

var loggerInitialized int32

func ConfigureLogger(production bool) error {
	var cfg zap.Config
	if production {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("failed to configure %v logger; %w", cfg, err)
	}

	zap.ReplaceGlobals(logger)

	atomic.StoreInt32(&loggerInitialized, 1)

	return nil
}

func IsLoggerInitialized() bool {
	return atomic.LoadInt32(&loggerInitialized) == 1
}
