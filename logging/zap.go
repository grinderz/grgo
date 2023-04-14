package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Zap *zap.Logger //nolint:gochecknoglobals

func New(appID string, cfg *Config) (*zap.Logger, error) {
	var (
		zcfg      zap.Config
		presetCfg *PresetConfig
	)

	switch cfg.Preset {
	case PresetDevelopment:
		zcfg = zap.NewDevelopmentConfig()
		presetCfg = &cfg.Development
	case PresetUnknown:
		fallthrough
	case PresetProduction:
		zcfg = zap.NewProductionConfig()
		presetCfg = &cfg.Production
	}

	zcfg.DisableCaller = !presetCfg.EnableCaller
	zcfg.DisableStacktrace = presetCfg.DisableStacktrace

	if err := parseLevel(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseLevelEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseTimeEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseDurationEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseOutputs(appID, presetCfg, &zcfg); err != nil {
		return nil, err
	}

	return zcfg.Build() //nolint:wrapcheck
}

func Setup(appID string, cfg *Config) {
	if cfg == nil {
		panic("empty logger config")
	}

	zp, err := New(appID, cfg)
	if err != nil {
		panic(err)
	}

	Zap = zp
}

func parseLevel(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.Level) == 0 {
		return nil
	}

	lvl, err := zap.ParseAtomicLevel(presetCfg.Level)
	if err != nil {
		return fmt.Errorf("parse log level failed: %w", err)
	}

	zcfg.Level = lvl

	return nil
}

func parseLevelEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.LevelEncoder) == 0 {
		return nil
	}

	var lvlEncoder zapcore.LevelEncoder

	if err := lvlEncoder.UnmarshalText([]byte(presetCfg.LevelEncoder)); err != nil {
		return fmt.Errorf("parse log level encoder failed: %w", err)
	}

	zcfg.EncoderConfig.EncodeLevel = lvlEncoder

	return nil
}

func parseTimeEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.TimeEncoder) > 0 {
		var tsEncoder zapcore.TimeEncoder

		if err := tsEncoder.UnmarshalText([]byte(presetCfg.TimeEncoder)); err != nil {
			return fmt.Errorf("parse log time encoder failed: %w", err)
		}

		zcfg.EncoderConfig.EncodeTime = tsEncoder
	} else if len(presetCfg.TimeLayout) > 0 {
		zcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(presetCfg.TimeLayout)
	}

	return nil
}

func parseDurationEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.DurationEncoder) == 0 {
		return nil
	}

	var durEncoder zapcore.DurationEncoder

	if err := durEncoder.UnmarshalText([]byte(presetCfg.DurationEncoder)); err != nil {
		return fmt.Errorf("parse log duration encoder failed: %w", err)
	}

	zcfg.EncoderConfig.EncodeDuration = durEncoder

	return nil
}

func parseOutputs(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.Outputs) == 0 {
		return nil
	}

	outputs := make([]string, 0, len(presetCfg.Outputs))
	fileEnabled := false

	for output, enabled := range presetCfg.Outputs {
		if !enabled {
			continue
		}

		if output == OutputFile && len(appID) > 0 {
			fileEnabled = true
			continue
		}

		outputs = append(outputs, output.String())
	}

	if len(outputs) > 0 {
		zcfg.OutputPaths = outputs
		zcfg.ErrorOutputPaths = outputs
	}

	if fileEnabled {
		if err := parseFileOutput(appID, presetCfg, zcfg); err != nil {
			return err
		}
	}

	return nil
}

func parseFileOutput(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
	var dir string

	if filepath.IsLocal(presetCfg.OutputFile.Dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to detect working directory: %w", err)
		}

		dir = filepath.Join(cwd, presetCfg.OutputFile.Dir)
	} else if filepath.IsAbs(presetCfg.OutputFile.Dir) {
		dir = presetCfg.OutputFile.Dir
	}

	if len(dir) > 0 {
		runTS := time.Now().Format(presetCfg.OutputFile.TimeLayout)
		location := filepath.Join(dir, fmt.Sprintf("%s-%s.log", appID, runTS))

		if len(location) > 0 {
			zcfg.OutputPaths = append(zcfg.OutputPaths, location)
		}
	}

	return nil
}
