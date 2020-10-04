package pkg

import (
	"fmt"
	"github.com/markbates/pkger"
	"go.uber.org/config"
	"os"
	"strings"
)

func PopulateConfig(target interface{}) error {
	cfgDir, err := pkger.Open("/configs")
	if err != nil {
		return fmt.Errorf("failed to open configs dir; %w", err)
	}

	cfgFiles := []string{"/application.yaml"}

	activeProfilesStr := os.Getenv("ACTIVE_PROFILES")
	if len(activeProfilesStr) != 0 {
		activeProfiles := strings.Split(activeProfilesStr, ",")
		for _, profile := range activeProfiles {
			file := fmt.Sprintf("/application-%s.yaml", strings.ToLower(strings.TrimSpace(profile)))
			cfgFiles = append(cfgFiles, file)
		}
	}

	cfgSources := make([]config.YAMLOption, len(cfgFiles))
	for i, file := range cfgFiles {
		reader, err := cfgDir.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open %s; %w", file, err)
		}
		cfgSources[i] = config.Source(reader)
	}

	cfgSources = append(cfgSources, config.Expand(os.LookupEnv))

	yamlConfig, err := config.NewYAML(cfgSources...)
	if err != nil {
		return fmt.Errorf("failed to parse application.yaml; %w", err)
	}

	if err := yamlConfig.Get("").Populate(target); err != nil {
		return fmt.Errorf("failed to populate target struct; %w", err)
	}

	return nil
}
