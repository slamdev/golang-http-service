package integration

import (
	"bytes"
	"fmt"
	"go.uber.org/config"
	"golang-http-service/configs"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func PopulateConfig(target interface{}) error {
	cfgFiles := []string{"application.yaml"}

	activeProfilesStr := os.Getenv("ACTIVE_PROFILES")
	if len(activeProfilesStr) != 0 {
		activeProfiles := strings.Split(activeProfilesStr, ",")
		for _, profile := range activeProfiles {
			file := fmt.Sprintf("application-%s.yaml", strings.ToLower(strings.TrimSpace(profile)))
			cfgFiles = append(cfgFiles, file)
		}
	}

	cfgSources := make([]config.YAMLOption, len(cfgFiles))
	for i, file := range cfgFiles {
		content, err := configs.Configs.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to open %s; %w", file, err)
		}
		cfgSources[i] = config.Source(bytes.NewReader(content))
	}

	if additionalLocation, set := os.LookupEnv("APP_CONFIG_ADDITIONAL_LOCATION"); set {
		additionalSources, err := collectFSConfigs(additionalLocation)
		if err != nil {
			return fmt.Errorf("failed to collect configs from APP_CONFIG_ADDITIONAL_LOCATION env var; %w", err)
		}
		for i := range additionalSources {
			f, err := os.Open(additionalSources[i])
			if err != nil {
				return fmt.Errorf("failed to open file; %w", err)
			}
			cfgSources = append(cfgSources, config.Source(f))
		}
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

func collectFSConfigs(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "application.yaml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to traverse through %s; %w", dir, err)
	}
	return files, nil
}
