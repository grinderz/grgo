package info

import (
	"fmt"
	"sync"

	"go.uber.org/zap/zapcore"
)

var (
	version   = "unknown"
	goOS      = "unknown"              //nolint:gochecknoglobals
	goArch    = "unknown"              //nolint:gochecknoglobals
	gitCommit = "unknown"              //nolint:gochecknoglobals
	buildDate = "1970-01-01T00:00:00Z" //nolint:gochecknoglobals
)

type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoOS      string `json:"go_os"`
	GoArch    string `json:"go_arch"`
}

var (
	instance Info      //nolint:gochecknoglobals
	once     sync.Once //nolint:gochecknoglobals
)

func GetInstance() Info {
	once.Do(func() {
		instance = newInfo()
	})

	return instance
}

func newInfo() Info {
	return Info{
		version,
		gitCommit,
		buildDate,
		goOS,
		goArch,
	}
}

func (i Info) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("version", i.Version)
	enc.AddString("gitCommit", i.GitCommit)
	enc.AddString("buildDate", i.BuildDate)
	enc.AddString("goOS", i.GoOS)
	enc.AddString("goArch", i.GoArch)

	return nil
}

func (i Info) String() string {
	return fmt.Sprintf("%#v", i)
}
