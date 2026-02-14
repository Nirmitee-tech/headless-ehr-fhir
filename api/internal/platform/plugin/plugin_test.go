package plugin

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type testPlugin struct {
	name           string
	routesCalled   bool
	migrateCalled  bool
}

func (p *testPlugin) Name() string { return p.name }

func (p *testPlugin) RegisterRoutes(api *echo.Group, fhir *echo.Group) {
	p.routesCalled = true
}

func (p *testPlugin) Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	p.migrateCalled = true
	return nil
}

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()
	p := &testPlugin{name: "test-plugin"}
	reg.Register(p)

	plugins := reg.Plugins()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name() != "test-plugin" {
		t.Errorf("expected test-plugin, got %s", plugins[0].Name())
	}
}

func TestRegistry_RegisterRoutes(t *testing.T) {
	reg := NewRegistry()
	p := &testPlugin{name: "route-plugin"}
	reg.Register(p)

	e := echo.New()
	api := e.Group("/api")
	fhir := e.Group("/fhir")
	reg.RegisterRoutes(api, fhir)

	if !p.routesCalled {
		t.Error("expected RegisterRoutes to be called on plugin")
	}
}

func TestRegistry_Migrate(t *testing.T) {
	reg := NewRegistry()
	p := &testPlugin{name: "migrate-plugin"}
	reg.Register(p)

	err := reg.Migrate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.migrateCalled {
		t.Error("expected Migrate to be called on plugin")
	}
}

func TestRegistry_Empty(t *testing.T) {
	reg := NewRegistry()

	if len(reg.Plugins()) != 0 {
		t.Error("expected 0 plugins")
	}

	// Should not panic with no plugins
	e := echo.New()
	reg.RegisterRoutes(e.Group("/api"), e.Group("/fhir"))
	reg.Migrate(context.Background(), nil)
}

func TestRegistry_MultiplePlugins(t *testing.T) {
	reg := NewRegistry()
	p1 := &testPlugin{name: "alpha"}
	p2 := &testPlugin{name: "beta"}
	reg.Register(p1)
	reg.Register(p2)

	if len(reg.Plugins()) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(reg.Plugins()))
	}
}
