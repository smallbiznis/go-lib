package server

import (
	"testing"

	"github.com/smallbiznis/go-lib/pkg/otelcol"
	"go.uber.org/fx"
)

func TestGrpcServer(t *testing.T) {
	if err := fx.New(
		otelcol.Resource,
		otelcol.TraceProvider,
		GrpcServerProvider,
		GrpcServerInvoke,
	).Err(); err != nil {
		t.Error(err)
	}
}
