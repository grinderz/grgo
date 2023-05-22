package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapFieldPkg(pkg string) zapcore.Field {
	return zap.String("pkg", pkg)
}
