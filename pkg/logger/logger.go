package logger

import (
	"github.com/smallbiznis/go-lib/pkg/env"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	NewZapLogger = fx.Module("zap.Logger", fx.Options(
		fx.Provide(
			InitLogger,
		),
	))
)

func InitLogger() (log *zap.Logger) {
	fields := zap.Fields(
		zap.String("service_name", env.Lookup("SERVICE_NAME", "example")),
		zap.String("service_version", env.Lookup("SERVICE_VERSION", "v1.0.0")),
		zap.String("service_namespace", env.Lookup("SERVICE_NAMESPACE", "smallbiznis")),
	)

	log = zap.Must(zap.NewDevelopment(fields))
	if env.Lookup("ENV", "development") == "production" {
		log = zap.Must(zap.NewProduction(fields))
	}

	zap.ReplaceGlobals(log)

	return log
}
