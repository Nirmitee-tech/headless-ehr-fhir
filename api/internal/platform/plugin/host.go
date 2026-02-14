package plugin

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// DomainPlugin defines the interface that plugins must implement to extend the EHR.
type DomainPlugin interface {
	Name() string
	RegisterRoutes(api *echo.Group, fhir *echo.Group)
	Migrate(ctx context.Context, pool *pgxpool.Pool) error
}

// Registry holds registered plugins.
type Registry struct {
	plugins []DomainPlugin
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(p DomainPlugin) {
	r.plugins = append(r.plugins, p)
}

func (r *Registry) RegisterRoutes(api *echo.Group, fhir *echo.Group) {
	for _, p := range r.plugins {
		p.RegisterRoutes(api, fhir)
	}
}

func (r *Registry) Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	for _, p := range r.plugins {
		if err := p.Migrate(ctx, pool); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) Plugins() []DomainPlugin {
	return r.plugins
}
