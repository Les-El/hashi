package console

import (
	"github.com/Les-El/chexum/internal/config"
	"testing"
)

func TestInitStreams(t *testing.T) {
	cfg := &config.Config{}
	streams, cleanup, err := InitStreams(cfg)
	if err != nil {
		t.Fatalf("InitStreams failed: %v", err)
	}
	defer cleanup()

	if streams.Out == nil {
		t.Error("Expected Out stream to be initialized")
	}
	if streams.Err == nil {
		t.Error("Expected Err stream to be initialized")
	}
}
