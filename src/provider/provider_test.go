package provider

import (
	"context"
	"testing"
	"time"
)

func TestNewProvider(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p := NewProvider(ctx)
	defer p.Release(ctx)
}
