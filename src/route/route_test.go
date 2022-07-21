package route

import (
	"context"
	"testing"
	"time"

	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/gorilla/mux"
)

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := mux.NewRouter()
	RegisterRoutes(r, p)
}
