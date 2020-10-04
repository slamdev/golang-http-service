package pkg

import (
	"fmt"
	"go.uber.org/zap"
)

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

	return nil
}
