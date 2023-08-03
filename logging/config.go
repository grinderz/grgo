package logging

type OutputFileConfig struct {
	Dir        string `yaml:"dir"        env:"DIR"         env-default:"logs"`
	TimeLayout string `yaml:"timeLayout" env:"TIME_LAYOUT" env-default:"2006-01-02"`
}

type PresetConfig struct {
	Level             string              `yaml:"level"             env:"LEVEL"               env-default:""`
	EnableCaller      bool                `yaml:"enableCaller"      env:"ENABLE_CALLER"       env-default:"false"`
	DisableStacktrace bool                `yaml:"disableStacktrace" env:"DISABLE_STACKTRACE"  env-default:"false"`
	LevelEncoder      string              `yaml:"levelEncoder"      env:"LEVEL_ENCODER"       env-default:""`
	TimeEncoder       string              `yaml:"timeEncoder"       env:"TIME_ENCODER"        env-default:""`
	TimeLayout        string              `yaml:"timeLayout"        env:"TIME_LAYOUT"         env-default:""`
	DurationEncoder   string              `yaml:"durationEncoder"   env:"DURATION_ENCODER"    env-default:"string"`
	Outputs           map[OutputEnum]bool `yaml:"outputs"           env:"OUTPUTS"             env-default:""`
	OutputFile        OutputFileConfig    `yaml:"outputFile"        env-prefix:"OUTPUT_FILE_"`
}

type Config struct {
	Preset PresetEnum `yaml:"preset" env:"PRESET" env-default:"production"`

	Development PresetConfig `yaml:"development" env-prefix:"DEVELOPMENT_"`
	Production  PresetConfig `yaml:"production"  env-prefix:"PRODUCTION_"`
}
