package main

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/ehr/ehr/internal/config"
	"github.com/ehr/ehr/internal/domain/admin"
	authpkg "github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/ccda"
	"github.com/ehr/ehr/internal/domain/behavioral"
	"github.com/ehr/ehr/internal/domain/billing"
	"github.com/ehr/ehr/internal/domain/careplan"
	"github.com/ehr/ehr/internal/domain/careteam"
	"github.com/ehr/ehr/internal/domain/cds"
	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/ehr/ehr/internal/domain/device"
	"github.com/ehr/ehr/internal/domain/diagnostics"
	"github.com/ehr/ehr/internal/domain/documents"
	"github.com/ehr/ehr/internal/domain/emergency"
	"github.com/ehr/ehr/internal/domain/conformance"
	"github.com/ehr/ehr/internal/domain/encounter"
	"github.com/ehr/ehr/internal/domain/episodeofcare"
	"github.com/ehr/ehr/internal/domain/familyhistory"
	"github.com/ehr/ehr/internal/domain/fhirlist"
	"github.com/ehr/ehr/internal/domain/financial"
	"github.com/ehr/ehr/internal/domain/healthcareservice"
	"github.com/ehr/ehr/internal/domain/identity"
	"github.com/ehr/ehr/internal/domain/immunization"
	"github.com/ehr/ehr/internal/domain/inbox"
	"github.com/ehr/ehr/internal/domain/measurereport"
	"github.com/ehr/ehr/internal/domain/medication"
	"github.com/ehr/ehr/internal/domain/nursing"
	"github.com/ehr/ehr/internal/domain/obstetrics"
	"github.com/ehr/ehr/internal/domain/oncology"
	"github.com/ehr/ehr/internal/domain/portal"
	"github.com/ehr/ehr/internal/domain/provenance"
	"github.com/ehr/ehr/internal/domain/relatedperson"
	"github.com/ehr/ehr/internal/domain/research"
	"github.com/ehr/ehr/internal/domain/scheduling"
	"github.com/ehr/ehr/internal/domain/supply"
	"github.com/ehr/ehr/internal/domain/surgery"
	"github.com/ehr/ehr/internal/domain/subscription"
	fhirtask "github.com/ehr/ehr/internal/domain/task"
	"github.com/ehr/ehr/internal/domain/terminology"
	"github.com/ehr/ehr/internal/domain/visionprescription"
	"github.com/ehr/ehr/internal/domain/workflow"
	fhirendpoint "github.com/ehr/ehr/internal/domain/endpoint"
	"github.com/ehr/ehr/internal/domain/bodystructure"
	"github.com/ehr/ehr/internal/domain/substance"
	fhirmedia "github.com/ehr/ehr/internal/domain/media"
	"github.com/ehr/ehr/internal/domain/devicerequest"
	"github.com/ehr/ehr/internal/domain/deviceusestatement"
	"github.com/ehr/ehr/internal/domain/coverageeligibility"
	"github.com/ehr/ehr/internal/domain/medicationknowledge"
	"github.com/ehr/ehr/internal/domain/organizationaffiliation"
	"github.com/ehr/ehr/internal/domain/person"
	fhirmeasure "github.com/ehr/ehr/internal/domain/measure"
	fhirlibrary "github.com/ehr/ehr/internal/domain/library"
	"github.com/ehr/ehr/internal/domain/devicedefinition"
	"github.com/ehr/ehr/internal/domain/devicemetric"
	"github.com/ehr/ehr/internal/domain/specimendefinition"
	"github.com/ehr/ehr/internal/domain/communicationrequest"
	"github.com/ehr/ehr/internal/domain/observationdefinition"
	"github.com/ehr/ehr/internal/domain/linkage"
	fhirbasic "github.com/ehr/ehr/internal/domain/basic"
	"github.com/ehr/ehr/internal/domain/verificationresult"
	"github.com/ehr/ehr/internal/domain/eventdefinition"
	"github.com/ehr/ehr/internal/domain/graphdefinition"
	"github.com/ehr/ehr/internal/domain/molecularsequence"
	"github.com/ehr/ehr/internal/domain/biologicallyderivedproduct"
	"github.com/ehr/ehr/internal/domain/catalogentry"
	"github.com/ehr/ehr/internal/domain/structuredefinition"
	"github.com/ehr/ehr/internal/domain/searchparameter"
	"github.com/ehr/ehr/internal/domain/codesystem"
	"github.com/ehr/ehr/internal/domain/valueset"
	"github.com/ehr/ehr/internal/domain/conceptmap"
	"github.com/ehr/ehr/internal/domain/implementationguide"
	"github.com/ehr/ehr/internal/domain/compartmentdefinition"
	"github.com/ehr/ehr/internal/domain/terminologycapabilities"
	"github.com/ehr/ehr/internal/domain/structuremap"
	"github.com/ehr/ehr/internal/domain/testscript"
	"github.com/ehr/ehr/internal/domain/testreport"
	"github.com/ehr/ehr/internal/domain/examplescenario"
	fhirevidence "github.com/ehr/ehr/internal/domain/evidence"
	"github.com/ehr/ehr/internal/domain/evidencevariable"
	"github.com/ehr/ehr/internal/domain/researchdefinition"
	"github.com/ehr/ehr/internal/domain/researchelementdefinition"
	"github.com/ehr/ehr/internal/domain/effectevidencesynthesis"
	"github.com/ehr/ehr/internal/domain/riskevidencesynthesis"
	"github.com/ehr/ehr/internal/domain/researchsubject"
	"github.com/ehr/ehr/internal/domain/documentmanifest"
	"github.com/ehr/ehr/internal/domain/substancespecification"
	"github.com/ehr/ehr/internal/domain/medicinalproduct"
	"github.com/ehr/ehr/internal/domain/medproductingredient"
	"github.com/ehr/ehr/internal/domain/medproductmanufactured"
	"github.com/ehr/ehr/internal/domain/medproductpackaged"
	"github.com/ehr/ehr/internal/domain/medproductauthorization"
	"github.com/ehr/ehr/internal/domain/medproductcontraindication"
	"github.com/ehr/ehr/internal/domain/medproductindication"
	"github.com/ehr/ehr/internal/domain/medproductinteraction"
	"github.com/ehr/ehr/internal/domain/medproductundesirableeffect"
	"github.com/ehr/ehr/internal/domain/medproductpharmaceutical"
	"github.com/ehr/ehr/internal/domain/substancepolymer"
	"github.com/ehr/ehr/internal/domain/substanceprotein"
	"github.com/ehr/ehr/internal/domain/substancenucleicacid"
	"github.com/ehr/ehr/internal/domain/substancesourcematerial"
	"github.com/ehr/ehr/internal/domain/substancereferenceinformation"
	"github.com/ehr/ehr/internal/domain/auditevent"
	"github.com/ehr/ehr/internal/domain/immunizationevaluation"
	"github.com/ehr/ehr/internal/platform/analytics"
	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/blobstore"
	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/internal/platform/hipaa"
	"github.com/ehr/ehr/internal/platform/hl7v2"
	"github.com/ehr/ehr/internal/platform/middleware"
	"github.com/ehr/ehr/internal/platform/notification"
	"github.com/ehr/ehr/internal/platform/openapi"
	"github.com/ehr/ehr/internal/platform/reporting"
	"github.com/ehr/ehr/internal/platform/sandbox"
	"github.com/ehr/ehr/internal/platform/bot"
	selfsched "github.com/ehr/ehr/internal/platform/scheduling"
	"github.com/ehr/ehr/internal/platform/telemetry"
	"github.com/ehr/ehr/internal/platform/webhook"
	"github.com/ehr/ehr/internal/platform/websocket"
)

// ConsentRepoAdapter adapts a documents.ConsentRepository to the
// auth.ConsentChecker interface, avoiding circular imports between the
// auth and documents packages.
type ConsentRepoAdapter struct {
	repo documents.ConsentRepository
}

// NewConsentRepoAdapter creates a new adapter.
func NewConsentRepoAdapter(repo documents.ConsentRepository) *ConsentRepoAdapter {
	return &ConsentRepoAdapter{repo: repo}
}

// ListActiveConsentsForPatient implements auth.ConsentChecker.
func (a *ConsentRepoAdapter) ListActiveConsentsForPatient(ctx context.Context, patientID uuid.UUID) ([]*auth.ConsentInfo, error) {
	consents, _, err := a.repo.ListByPatient(ctx, patientID, 100, 0)
	if err != nil {
		return nil, err
	}

	var result []*auth.ConsentInfo
	for _, c := range consents {
		info := &auth.ConsentInfo{
			Status: c.Status,
		}
		if c.Scope != nil {
			info.Scope = *c.Scope
		}
		if c.ProvisionType != nil {
			info.ProvisionType = *c.ProvisionType
		}
		if c.ProvisionAction != nil {
			info.ProvisionAction = *c.ProvisionAction
		}
		info.ProvisionStart = c.ProvisionStart
		info.ProvisionEnd = c.ProvisionEnd
		result = append(result, info)
	}
	return result, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "ehr-server",
		Short: "Headless EHR API Server",
	}

	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(migrateCmd())
	rootCmd.AddCommand(tenantCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the EHR API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
	}

	// migrate up
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Apply pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			schema, _ := cmd.Flags().GetString("schema")
			dir, _ := cmd.Flags().GetString("dir")

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			ctx := context.Background()
			pool, err := db.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
			if err != nil {
				return err
			}
			defer pool.Close()

			migrator := db.NewMigrator(pool, dir)
			fmt.Printf("Running migrations on schema: %s\n", schema)

			count, err := migrator.Up(ctx, schema)
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			fmt.Printf("Applied %d migration(s) successfully.\n", count)
			return nil
		},
	}
	upCmd.Flags().String("schema", "tenant_default", "Target schema for migrations")
	upCmd.Flags().String("dir", "./migrations", "Path to migrations directory")
	cmd.AddCommand(upCmd)

	// migrate status
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			schema, _ := cmd.Flags().GetString("schema")
			dir, _ := cmd.Flags().GetString("dir")

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			ctx := context.Background()
			pool, err := db.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
			if err != nil {
				return err
			}
			defer pool.Close()

			migrator := db.NewMigrator(pool, dir)
			statuses, err := migrator.Status(ctx, schema)
			if err != nil {
				return fmt.Errorf("failed to get migration status: %w", err)
			}

			fmt.Printf("Migration status for schema: %s\n", schema)
			fmt.Printf("%-10s %-40s %-10s %s\n", "VERSION", "NAME", "STATUS", "APPLIED AT")
			fmt.Println("---------- ---------------------------------------- ---------- --------------------")
			for _, s := range statuses {
				status := "pending"
				appliedAt := ""
				if s.Applied {
					status = "applied"
					if s.AppliedAt != nil {
						appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
					}
				}
				fmt.Printf("%-10d %-40s %-10s %s\n", s.Version, s.Name, status, appliedAt)
			}
			return nil
		},
	}
	statusCmd.Flags().String("schema", "tenant_default", "Target schema for migrations")
	statusCmd.Flags().String("dir", "./migrations", "Path to migrations directory")
	cmd.AddCommand(statusCmd)

	// migrate down - keep as warning
	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "Rollback last migration (not supported)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("WARNING: migrate down is destructive and not supported by the built-in runner.")
			fmt.Println("Use Atlas CLI for migration rollback: atlas schema apply --dir migrations/")
			return nil
		},
	})

	return cmd
}

func tenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage tenants",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tenant schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			ctx := context.Background()
			pool, err := db.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
			if err != nil {
				return err
			}
			defer pool.Close()

			fmt.Printf("Creating tenant schema: tenant_%s\n", name)
			if err := db.CreateTenantSchema(ctx, pool, name, ""); err != nil {
				return err
			}
			fmt.Println("Tenant created successfully. Run migrations with: scripts/migrate.sh", name)
			return nil
		},
	}
	createCmd.Flags().String("name", "", "Tenant identifier (alphanumeric)")

	cmd.AddCommand(createCmd)
	return cmd
}

func runServer() error {
	// Logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if os.Getenv("ENV") == "development" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	}

	// Config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	// Database
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()
	logger.Info().Msg("connected to database")

	// Echo server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Global middleware
	e.Use(middleware.Recovery(logger))
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(logger))
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID", "X-Break-Glass"},
	}))

	// Auth middleware
	if cfg.IsDev() {
		e.Use(auth.DevAuthMiddleware())
	} else {
		e.Use(auth.JWTMiddleware(auth.JWTConfig{
			Issuer:   cfg.AuthIssuer,
			Audience: cfg.AuthAudience,
			JWKSURL:  cfg.AuthJWKSURL,
		}))
	}

	// Tenant middleware
	e.Use(db.TenantMiddleware(pool, cfg.DefaultTenant))

	// Audit middleware
	e.Use(middleware.Audit(logger))

	// API groups
	apiV1 := e.Group("/api/v1")
	fhirGroup := e.Group("/fhir")

	// Rate limiting middleware
	rateLimitCfg := middleware.RateLimitConfig{
		RequestsPerSecond: cfg.RateLimitRPS,
		BurstSize:         cfg.RateLimitBurst,
	}
	if rateLimitCfg.RequestsPerSecond <= 0 {
		rateLimitCfg = middleware.DefaultRateLimitConfig()
	}
	apiV1.Use(middleware.RateLimit(rateLimitCfg))
	fhirGroup.Use(middleware.RateLimit(rateLimitCfg))

	// ABAC + Consent enforcement middleware on FHIR group.
	// The consent repo is created early so the middleware can be wired before
	// domain handlers register their routes on fhirGroup.
	abacEngine := auth.NewABACEngine(auth.DefaultPolicies())
	fhirGroup.Use(auth.ABACMiddleware(abacEngine))

	consentRepo := documents.NewConsentRepoPG(pool)
	consentChecker := NewConsentRepoAdapter(consentRepo)
	fhirGroup.Use(auth.ConsentEnforcementMiddleware(consentChecker))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"version": "0.1.0",
		})
	})

	// Dynamic CapabilityStatement builder
	baseURL := fmt.Sprintf("http://localhost:%s/fhir", cfg.Port)
	capBuilder := fhir.NewCapabilityBuilder(baseURL, "0.1.0")

	// Configure SMART on FHIR OAuth URIs
	if cfg.AuthIssuer != "" {
		capBuilder.SetOAuthURIs(
			cfg.AuthIssuer+"/protocol/openid-connect/auth",
			cfg.AuthIssuer+"/protocol/openid-connect/token",
		)
	}

	// Register all domain resources with the capability builder
	// Identity domain
	capBuilder.AddResource("Patient", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "family", Type: "string"},
		{Name: "given", Type: "string"},
		{Name: "birthdate", Type: "date"},
		{Name: "gender", Type: "token"},
		{Name: "identifier", Type: "token"},
	})
	capBuilder.AddResource("Practitioner", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "family", Type: "string"},
		{Name: "identifier", Type: "token"},
	})
	capBuilder.AddResource("PractitionerRole", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "practitioner", Type: "reference"},
		{Name: "organization", Type: "reference"},
		{Name: "role", Type: "token"},
		{Name: "active", Type: "token"},
	})

	// Admin domain
	capBuilder.AddResource("Organization", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "type", Type: "token"},
		{Name: "active", Type: "token"},
	})
	capBuilder.AddResource("Location", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "status", Type: "token"},
	})

	capBuilder.AddResource("Group", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "type", Type: "token"},
		{Name: "name", Type: "string"},
		{Name: "active", Type: "token"},
	})

	// Encounter domain
	capBuilder.AddResource("Encounter", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "class", Type: "token"},
		{Name: "date", Type: "date"},
	})

	// Clinical domain
	capBuilder.AddResource("Condition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "clinical-status", Type: "token"},
		{Name: "category", Type: "token"},
		{Name: "code", Type: "token"},
	})
	capBuilder.AddResource("Observation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "category", Type: "token"},
		{Name: "code", Type: "token"},
		{Name: "date", Type: "date"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("AllergyIntolerance", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "clinical-status", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "criticality", Type: "token"},
	})
	capBuilder.AddResource("Procedure", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "code", Type: "token"},
		{Name: "date", Type: "date"},
	})
	capBuilder.AddResource("NutritionOrder", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Medication domain
	capBuilder.AddResource("Medication", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "code", Type: "token"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("MedicationRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "intent", Type: "token"},
		{Name: "date", Type: "date"},
	})
	capBuilder.AddResource("MedicationAdministration", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("MedicationDispense", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("MedicationStatement", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Diagnostics domain
	capBuilder.AddResource("ServiceRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "category", Type: "token"},
		{Name: "code", Type: "token"},
	})
	capBuilder.AddResource("DiagnosticReport", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "category", Type: "token"},
		{Name: "code", Type: "token"},
		{Name: "date", Type: "date"},
	})
	capBuilder.AddResource("ImagingStudy", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("Specimen", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Scheduling domain
	capBuilder.AddResource("Appointment", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "date", Type: "date"},
	})
	capBuilder.AddResource("Schedule", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "actor", Type: "reference"},
		{Name: "active", Type: "token"},
	})
	capBuilder.AddResource("Slot", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "schedule", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "start", Type: "date"},
	})
	capBuilder.AddResource("AppointmentResponse", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "appointment", Type: "reference"},
		{Name: "actor", Type: "reference"},
	})

	// Billing domain
	capBuilder.AddResource("Coverage", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("Claim", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("ClaimResponse", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("ExplanationOfBenefit", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("Invoice", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Documents domain
	capBuilder.AddResource("Consent", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "category", Type: "token"},
	})
	capBuilder.AddResource("DocumentReference", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "date", Type: "date"},
	})
	capBuilder.AddResource("Composition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "type", Type: "token"},
	})

	// Inbox domain
	capBuilder.AddResource("Communication", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Research domain
	capBuilder.AddResource("ResearchStudy", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "title", Type: "string"},
	})

	// Portal domain
	capBuilder.AddResource("Questionnaire", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("QuestionnaireResponse", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "questionnaire", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Immunization domain
	capBuilder.AddResource("Immunization", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "vaccine-code", Type: "token"},
		{Name: "date", Type: "date"},
	})
	capBuilder.AddResource("ImmunizationRecommendation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "vaccine-type", Type: "token"},
		{Name: "status", Type: "token"},
	})

	// CarePlan domain
	capBuilder.AddResource("CarePlan", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "category", Type: "token"},
	})
	capBuilder.AddResource("Goal", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "lifecycle-status", Type: "token"},
	})

	// FamilyHistory domain
	capBuilder.AddResource("FamilyMemberHistory", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "relationship", Type: "token"},
	})

	// RelatedPerson domain
	capBuilder.AddResource("RelatedPerson", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "relationship", Type: "token"},
	})

	// Provenance domain
	capBuilder.AddResource("Provenance", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "target", Type: "reference"},
		{Name: "agent", Type: "reference"},
	})

	// CareTeam domain
	capBuilder.AddResource("CareTeam", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "category", Type: "token"},
	})

	// Task domain
	capBuilder.AddResource("Task", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "owner", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "intent", Type: "token"},
		{Name: "priority", Type: "token"},
		{Name: "code", Type: "token"},
	})

	// Device domain
	capBuilder.AddResource("Device", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "manufacturer", Type: "string"},
	})

	// Subscription domain
	capBuilder.AddResource("Subscription", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "criteria", Type: "string"},
		{Name: "url", Type: "uri"},
	})

	// Clinical Safety resources
	capBuilder.AddResource("Flag", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("DetectedIssue", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("AdverseEvent", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "actuality", Type: "token"},
	})
	capBuilder.AddResource("ClinicalImpression", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("RiskAssessment", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Care Delivery resources
	capBuilder.AddResource("EpisodeOfCare", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("HealthcareService", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "active", Type: "token"},
	})
	capBuilder.AddResource("MeasureReport", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("List", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Financial resources
	capBuilder.AddResource("Account", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("InsurancePlan", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("PaymentNotice", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("PaymentReconciliation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("ChargeItem", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("ChargeItemDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("Contract", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("EnrollmentRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("EnrollmentResponse", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})

	// Workflow resources
	capBuilder.AddResource("ActivityDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("RequestGroup", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("GuidanceResponse", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Supply resources
	capBuilder.AddResource("SupplyRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("SupplyDelivery", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})

	// Conformance resources
	capBuilder.AddResource("NamingSystem", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("OperationDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("MessageDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("MessageHeader", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "event", Type: "token"},
	})

	// Specialty resources
	capBuilder.AddResource("VisionPrescription", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Simple resources (Endpoint, BodyStructure, Substance, Media, DeviceRequest, DeviceUseStatement)
	capBuilder.AddResource("Endpoint", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
		{Name: "organization", Type: "reference"},
	})
	capBuilder.AddResource("BodyStructure", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
	})
	capBuilder.AddResource("Substance", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "code", Type: "token"},
	})
	capBuilder.AddResource("Media", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "type", Type: "token"},
	})
	capBuilder.AddResource("DeviceRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "intent", Type: "token"},
	})
	capBuilder.AddResource("DeviceUseStatement", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	// Coverage Eligibility resources
	capBuilder.AddResource("CoverageEligibilityRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "purpose", Type: "token"},
	})
	capBuilder.AddResource("CoverageEligibilityResponse", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "outcome", Type: "token"},
		{Name: "request", Type: "reference"},
	})
	capBuilder.AddResource("MedicationKnowledge", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "code", Type: "token"},
	})
	capBuilder.AddResource("OrganizationAffiliation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "active", Type: "token"},
		{Name: "organization", Type: "reference"},
		{Name: "participating-organization", Type: "reference"},
		{Name: "specialty", Type: "token"},
	})
	capBuilder.AddResource("Person", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "name", Type: "string"},
		{Name: "gender", Type: "token"},
		{Name: "active", Type: "token"},
	})
	capBuilder.AddResource("Measure", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "url", Type: "uri"},
	})
	capBuilder.AddResource("Library", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "url", Type: "uri"},
		{Name: "type", Type: "token"},
	})
	capBuilder.AddResource("DeviceDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "manufacturer", Type: "string"},
		{Name: "model-number", Type: "string"},
		{Name: "type", Type: "token"},
	})
	capBuilder.AddResource("DeviceMetric", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "source", Type: "reference"},
		{Name: "type", Type: "token"},
		{Name: "category", Type: "token"},
	})
	capBuilder.AddResource("SpecimenDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "type", Type: "token"},
	})
	capBuilder.AddResource("CommunicationRequest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
		{Name: "priority", Type: "token"},
		{Name: "category", Type: "token"},
	})
	capBuilder.AddResource("ObservationDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "code", Type: "token"},
		{Name: "category", Type: "token"},
	})
	capBuilder.AddResource("Linkage", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "author", Type: "reference"},
		{Name: "source", Type: "reference"},
	})
	capBuilder.AddResource("Basic", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "code", Type: "token"},
		{Name: "subject", Type: "reference"},
		{Name: "author", Type: "reference"},
	})
	capBuilder.AddResource("VerificationResult", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "target", Type: "reference"},
	})
	capBuilder.AddResource("EventDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
		{Name: "url", Type: "uri"},
	})
	capBuilder.AddResource("GraphDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
		{Name: "url", Type: "uri"},
		{Name: "start", Type: "token"},
	})
	capBuilder.AddResource("MolecularSequence", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "type", Type: "token"},
		{Name: "patient", Type: "reference"},
	})
	capBuilder.AddResource("BiologicallyDerivedProduct", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "product-category", Type: "token"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("CatalogEntry", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "orderable", Type: "token"},
	})

	// Conformance & Terminology resources
	capBuilder.AddResource("StructureDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "token"},
		{Name: "kind", Type: "token"},
		{Name: "base", Type: "uri"},
	})
	capBuilder.AddResource("SearchParameter", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "base", Type: "token"},
	})
	capBuilder.AddResource("CodeSystem", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
		{Name: "content", Type: "token"},
	})
	capBuilder.AddResource("ValueSet", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
	})
	capBuilder.AddResource("ConceptMap", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
		{Name: "source", Type: "reference"},
		{Name: "target", Type: "reference"},
	})
	capBuilder.AddResource("ImplementationGuide", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("CompartmentDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "token"},
	})
	capBuilder.AddResource("TerminologyCapabilities", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("StructureMap", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("TestScript", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})

	// Research & Evidence resources
	capBuilder.AddResource("TestReport", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "result", Type: "token"},
	})
	capBuilder.AddResource("ExampleScenario", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("Evidence", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("EvidenceVariable", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("ResearchDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("ResearchElementDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("EffectEvidenceSynthesis", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("RiskEvidenceSynthesis", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "url", Type: "uri"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("ResearchSubject", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "study", Type: "reference"},
		{Name: "individual", Type: "reference"},
	})
	capBuilder.AddResource("DocumentManifest", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "subject", Type: "reference"},
		{Name: "type", Type: "token"},
	})
	capBuilder.AddResource("SubstanceSpecification", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "type", Type: "token"},
		{Name: "domain", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProduct", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "type", Type: "token"},
		{Name: "domain", Type: "token"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductIngredient", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "role", Type: "token"},
		{Name: "substance", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductManufactured", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "dose-form", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductPackaged", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "subject", Type: "reference"},
	})
	capBuilder.AddResource("MedicinalProductAuthorization", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "subject", Type: "reference"},
		{Name: "country", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductContraindication", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "subject", Type: "reference"},
		{Name: "disease", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductIndication", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "subject", Type: "reference"},
		{Name: "disease", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductInteraction", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "subject", Type: "reference"},
		{Name: "type", Type: "token"},
	})
	capBuilder.AddResource("MedicinalProductUndesirableEffect", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "subject", Type: "reference"},
	})
	capBuilder.AddResource("MedicinalProductPharmaceutical", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "route", Type: "token"},
	})
	capBuilder.AddResource("SubstancePolymer", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "class", Type: "token"},
	})
	capBuilder.AddResource("SubstanceProtein", fhir.DefaultInteractions(), nil)
	capBuilder.AddResource("SubstanceNucleicAcid", fhir.DefaultInteractions(), nil)
	capBuilder.AddResource("SubstanceSourceMaterial", fhir.DefaultInteractions(), nil)
	capBuilder.AddResource("SubstanceReferenceInformation", fhir.DefaultInteractions(), nil)
	capBuilder.AddResource("PlanDefinition", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "name", Type: "string"},
	})
	capBuilder.AddResource("Binary", fhir.DefaultInteractions(), nil)
	capBuilder.AddResource("AuditEvent", []string{"read", "vread", "search-type", "history-instance"}, []fhir.SearchParam{
		{Name: "action", Type: "token"},
		{Name: "type", Type: "token"},
		{Name: "outcome", Type: "token"},
		{Name: "agent", Type: "string"},
		{Name: "entity-type", Type: "token"},
	})
	capBuilder.AddResource("ImmunizationEvaluation", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "status", Type: "token"},
		{Name: "patient", Type: "reference"},
		{Name: "target-disease", Type: "token"},
		{Name: "dose-status", Type: "token"},
	})

	// Set advanced capabilities for all registered resource types
	defaultCaps := fhir.DefaultCapabilityOptions()
	for _, rt := range []string{
		"Patient", "Practitioner", "Organization", "Location", "Encounter",
		"Condition", "Observation", "AllergyIntolerance", "Procedure", "NutritionOrder",
		"Medication", "MedicationRequest", "MedicationAdministration", "MedicationDispense",
		"ServiceRequest", "DiagnosticReport", "ImagingStudy", "Specimen",
		"Appointment", "Schedule", "Slot",
		"Coverage", "Claim",
		"Consent", "DocumentReference", "Composition",
		"Communication",
		"ResearchStudy",
		"Questionnaire", "QuestionnaireResponse",
		"Immunization", "ImmunizationRecommendation",
		"CarePlan", "Goal",
		"FamilyMemberHistory",
		"RelatedPerson",
		"Provenance",
		"CareTeam",
		"Task",
		"Device",
		"Subscription",
		"Flag", "DetectedIssue", "AdverseEvent", "ClinicalImpression", "RiskAssessment",
		"EpisodeOfCare", "HealthcareService", "MeasureReport", "List",
		"Account", "InsurancePlan", "PaymentNotice", "PaymentReconciliation",
		"ChargeItem", "ChargeItemDefinition", "Contract", "EnrollmentRequest", "EnrollmentResponse",
		"ActivityDefinition", "RequestGroup", "GuidanceResponse",
		"SupplyRequest", "SupplyDelivery",
		"NamingSystem", "OperationDefinition", "MessageDefinition", "MessageHeader",
		"VisionPrescription",
		"Endpoint", "BodyStructure", "Substance", "Media", "DeviceRequest", "DeviceUseStatement",
		"CoverageEligibilityRequest", "CoverageEligibilityResponse",
		"MedicationKnowledge", "OrganizationAffiliation", "Person",
		"Measure", "Library", "DeviceDefinition", "DeviceMetric", "SpecimenDefinition",
		"CommunicationRequest", "ObservationDefinition", "Linkage", "Basic",
		"VerificationResult", "EventDefinition", "GraphDefinition",
		"MolecularSequence", "BiologicallyDerivedProduct", "CatalogEntry",
		"StructureDefinition", "SearchParameter", "CodeSystem", "ValueSet", "ConceptMap",
		"ImplementationGuide", "CompartmentDefinition", "TerminologyCapabilities",
		"StructureMap", "TestScript",
		"TestReport", "ExampleScenario", "Evidence", "EvidenceVariable",
		"ResearchDefinition", "ResearchElementDefinition",
		"EffectEvidenceSynthesis", "RiskEvidenceSynthesis",
		"ResearchSubject", "DocumentManifest", "SubstanceSpecification",
		"MedicinalProduct", "MedicinalProductIngredient", "MedicinalProductManufactured",
		"MedicinalProductPackaged", "MedicinalProductAuthorization",
		"MedicinalProductContraindication", "MedicinalProductIndication",
		"MedicinalProductInteraction", "MedicinalProductUndesirableEffect",
		"MedicinalProductPharmaceutical",
		"SubstancePolymer", "SubstanceProtein", "SubstanceNucleicAcid",
		"SubstanceSourceMaterial", "SubstanceReferenceInformation",
		"PlanDefinition", "Binary", "AuditEvent", "ImmunizationEvaluation",
	} {
		capBuilder.SetResourceCapabilities(rt, defaultCaps)
	}

	// Include registry for _include/_revinclude resolution
	includeRegistry := fhir.NewIncludeRegistry()

	// History repository for resource versioning
	historyRepo := fhir.NewHistoryRepository()
	versionTracker := fhir.NewVersionTracker(historyRepo)

	// Register common references for _include support
	for _, rt := range []string{"Condition", "Observation", "AllergyIntolerance", "Procedure",
		"MedicationRequest", "MedicationAdministration", "MedicationDispense",
		"ServiceRequest", "DiagnosticReport", "Encounter", "Appointment",
		"Claim", "Coverage", "Consent", "DocumentReference", "Composition",
		"Communication", "QuestionnaireResponse", "Specimen", "ImagingStudy",
		"Immunization", "ImmunizationRecommendation", "CarePlan", "Goal",
		"FamilyMemberHistory", "RelatedPerson", "CareTeam", "Task", "Device"} {
		includeRegistry.RegisterReference(rt, "patient", "Patient")
		includeRegistry.RegisterReference(rt, "subject", "Patient")
	}
	for _, rt := range []string{"Condition", "Observation", "Procedure",
		"MedicationRequest", "MedicationAdministration", "ServiceRequest",
		"DiagnosticReport", "Immunization", "CarePlan", "CareTeam", "Task"} {
		includeRegistry.RegisterReference(rt, "encounter", "Encounter")
	}
	for _, rt := range []string{"Flag", "DetectedIssue", "AdverseEvent", "ClinicalImpression",
		"RiskAssessment", "EpisodeOfCare", "MeasureReport", "ChargeItem",
		"RequestGroup", "GuidanceResponse", "VisionPrescription",
		"BodyStructure", "Media", "DeviceRequest", "DeviceUseStatement",
		"ImmunizationEvaluation"} {
		includeRegistry.RegisterReference(rt, "patient", "Patient")
		includeRegistry.RegisterReference(rt, "subject", "Patient")
	}
	for _, rt := range []string{"CoverageEligibilityRequest", "CoverageEligibilityResponse"} {
		includeRegistry.RegisterReference(rt, "patient", "Patient")
	}
	includeRegistry.RegisterReference("CoverageEligibilityRequest", "provider", "Practitioner")
	includeRegistry.RegisterReference("CoverageEligibilityRequest", "insurer", "Organization")
	includeRegistry.RegisterReference("CoverageEligibilityResponse", "insurer", "Organization")
	includeRegistry.RegisterReference("CoverageEligibilityResponse", "request", "CoverageEligibilityRequest")
	includeRegistry.RegisterReference("MedicationKnowledge", "manufacturer", "Organization")
	includeRegistry.RegisterReference("OrganizationAffiliation", "organization", "Organization")
	includeRegistry.RegisterReference("OrganizationAffiliation", "participating-organization", "Organization")
	includeRegistry.RegisterReference("OrganizationAffiliation", "location", "Location")
	includeRegistry.RegisterReference("Person", "organization", "Organization")
	includeRegistry.RegisterReference("DeviceDefinition", "owner", "Organization")
	includeRegistry.RegisterReference("DeviceMetric", "source", "Device")
	includeRegistry.RegisterReference("DeviceMetric", "parent", "Device")
	includeRegistry.RegisterReference("CommunicationRequest", "patient", "Patient")
	includeRegistry.RegisterReference("CommunicationRequest", "encounter", "Encounter")
	includeRegistry.RegisterReference("CommunicationRequest", "requester", "Practitioner")
	includeRegistry.RegisterReference("CommunicationRequest", "recipient", "Practitioner")
	includeRegistry.RegisterReference("CommunicationRequest", "sender", "Practitioner")
	includeRegistry.RegisterReference("MolecularSequence", "patient", "Patient")
	includeRegistry.RegisterReference("Linkage", "author", "Practitioner")
	includeRegistry.RegisterReference("Basic", "author", "Practitioner")
	includeRegistry.RegisterReference("Provenance", "target", "Patient")
	includeRegistry.RegisterReference("Provenance", "agent", "Practitioner")
	includeRegistry.RegisterReference("Endpoint", "organization", "Organization")

	// Wire _include/_revinclude middleware into the FHIR search group.
	// Fetchers are registered below after services are initialized; since
	// the middleware holds a pointer to the registry, late registration works.
	fhirGroup.Use(fhir.ContentNegotiationMiddleware())
	fhirGroup.Use(fhir.ConditionalReadMiddleware())
	fhirGroup.Use(fhir.IncludeMiddleware(includeRegistry))
	fhirGroup.Use(fhir.SearchMiddleware())
	fhirGroup.Use(fhir.PreferMiddleware())

	// FHIR metadata (dynamic CapabilityStatement)
	fhirGroup.GET("/metadata", func(c echo.Context) error {
		return c.JSON(http.StatusOK, capBuilder.Build())
	})

	// SMART on FHIR discovery — use DB-backed launch context store for
	// horizontal scalability (contexts survive restarts and are shared
	// across instances). Falls back to in-memory if pool is nil.
	smartStore := auth.NewPGLaunchContextStoreFromPool(pool, 5*time.Minute)
	auth.RegisterSMARTEndpoints(fhirGroup, cfg.AuthIssuer, smartStore)

	// FHIR Bundle handler (transaction/batch processing)
	bundleProcessor := &fhir.DefaultBundleProcessor{}
	bundleHandler := fhir.NewBundleHandler(bundleProcessor)
	bundleHandler.RegisterRoutes(fhirGroup)

	// FHIR History handler (system-level and type-level _history)
	historyHandler := fhir.NewHistoryHandler(historyRepo)
	historyHandler.RegisterRoutes(fhirGroup)

	// FHIR Async job status handler (Prefer: respond-async support)
	asyncStore := fhir.NewInMemoryAsyncJobStore()
	fhirGroup.GET("/_async/:jobId", fhir.AsyncStatusHandler(asyncStore))
	fhirGroup.DELETE("/_async/:jobId", fhir.AsyncDeleteHandler(asyncStore))

	// FHIR $meta operations (resource tag/security/profile management)
	metaStore := fhir.NewInMemoryMetaStore()
	metaHandler := fhir.NewMetaHandler(metaStore)
	metaHandler.RegisterRoutes(fhirGroup)

	// FHIR $diff operation (resource version comparison)
	fhirGroup.GET("/:resourceType/:id/$diff", fhir.DiffHandler(historyRepo))

	// FHIR Observation/$lastn (latest N observations per code)
	fhirGroup.GET("/Observation/$lastn", fhir.LastNHandler(nil))
	fhirGroup.POST("/Observation/$lastn", fhir.LastNHandler(nil))

	// FHIR Observation/$stats (observation statistics)
	fhirGroup.GET("/Observation/$stats", fhir.StatsHandler(nil))
	fhirGroup.POST("/Observation/$stats", fhir.StatsHandler(nil))

	// -- Register Domain Handlers --

	// Admin domain
	orgRepo := admin.NewOrganizationRepo(pool)
	deptRepo := admin.NewDepartmentRepo(pool)
	locRepo := admin.NewLocationRepo(pool)
	userRepo := admin.NewSystemUserRepo(pool)
	adminSvc := admin.NewService(orgRepo, deptRepo, locRepo, userRepo)
	adminSvc.SetVersionTracker(versionTracker)
	adminHandler := admin.NewHandler(adminSvc)
	adminHandler.RegisterRoutes(apiV1, fhirGroup)

	// FHIR Group resource
	fhirGroupRepo := admin.NewGroupRepo(pool)
	fhirGroupSvc := admin.NewGroupService(fhirGroupRepo)
	fhirGroupHandler := admin.NewGroupHandler(fhirGroupSvc)
	fhirGroupHandler.RegisterGroupRoutes(apiV1, fhirGroup)

	// Identity domain (with optional PHI encryption)
	var phiEncryptor hipaa.FieldEncryptor
	if cfg.HIPAAEncryptionKey != "" {
		keyBytes, err := hex.DecodeString(cfg.HIPAAEncryptionKey)
		if err != nil {
			logger.Fatal().Err(err).Msg("HIPAA_ENCRYPTION_KEY must be a valid hex-encoded 32-byte key")
		}
		enc, err := hipaa.NewRotatingEncryptor(keyBytes, 1)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to create PHI encryptor")
		}
		phiEncryptor = enc
		logger.Info().Msg("PHI field-level encryption enabled")
	} else {
		logger.Warn().Msg("HIPAA_ENCRYPTION_KEY not set; PHI field-level encryption is disabled")
	}

	var patientRepo identity.PatientRepository
	if phiEncryptor != nil {
		patientRepo = identity.NewPatientRepoWithEncryption(pool, phiEncryptor)
	} else {
		patientRepo = identity.NewPatientRepo(pool)
	}
	var practRepo identity.PractitionerRepository
	if phiEncryptor != nil {
		practRepo = identity.NewPractitionerRepoWithEncryption(pool, phiEncryptor)
	} else {
		practRepo = identity.NewPractitionerRepo(pool)
	}
	patientLinkRepo := identity.NewPatientLinkRepo(pool)
	practRoleRepo := identity.NewPractitionerRoleRepoPG(pool)
	identitySvc := identity.NewService(patientRepo, practRepo, patientLinkRepo, practRoleRepo)
	identitySvc.SetVersionTracker(versionTracker)
	identityHandler := identity.NewHandler(identitySvc)
	identityHandler.RegisterRoutes(apiV1, fhirGroup)

	// Encounter domain
	encRepo := encounter.NewRepo(pool)
	encSvc := encounter.NewService(encRepo)
	encSvc.SetVersionTracker(versionTracker)
	encHandler := encounter.NewHandler(encSvc)
	encHandler.RegisterRoutes(apiV1, fhirGroup)

	// Clinical domain
	condRepo := clinical.NewConditionRepoPG(pool)
	obsRepo := clinical.NewObservationRepoPG(pool)
	allergyRepo := clinical.NewAllergyRepoPG(pool)
	procRepo := clinical.NewProcedureRepoPG(pool)
	clinicalSvc := clinical.NewService(condRepo, obsRepo, allergyRepo, procRepo)
	clinicalSvc.SetVersionTracker(versionTracker)
	clinicalHandler := clinical.NewHandler(clinicalSvc)
	clinicalHandler.RegisterRoutes(apiV1, fhirGroup)

	// NutritionOrder domain
	nutritionRepo := clinical.NewNutritionOrderRepoPG(pool)
	nutritionSvc := clinical.NewNutritionOrderService(nutritionRepo)
	nutritionHandler := clinical.NewNutritionOrderHandler(nutritionSvc)
	nutritionHandler.RegisterNutritionOrderRoutes(apiV1, fhirGroup)

	// Diagnostics domain
	srRepo := diagnostics.NewServiceRequestRepoPG(pool)
	specRepo := diagnostics.NewSpecimenRepoPG(pool)
	dxReportRepo := diagnostics.NewDiagnosticReportRepoPG(pool)
	imgRepo := diagnostics.NewImagingStudyRepoPG(pool)
	orderHistRepo := diagnostics.NewOrderStatusHistoryRepoPG(pool)
	dxSvc := diagnostics.NewService(srRepo, specRepo, dxReportRepo, imgRepo, orderHistRepo)
	dxSvc.SetVersionTracker(versionTracker)
	dxHandler := diagnostics.NewHandler(dxSvc)
	dxHandler.RegisterRoutes(apiV1, fhirGroup)

	// Medication domain
	medRepo := medication.NewMedicationRepoPG(pool)
	medReqRepo := medication.NewMedicationRequestRepoPG(pool)
	medAdminRepo := medication.NewMedicationAdministrationRepoPG(pool)
	medDispRepo := medication.NewMedicationDispenseRepoPG(pool)
	medStmtRepo := medication.NewMedicationStatementRepoPG(pool)
	medSvc := medication.NewService(medRepo, medReqRepo, medAdminRepo, medDispRepo, medStmtRepo)
	medSvc.SetVersionTracker(versionTracker)
	medHandler := medication.NewHandler(medSvc, dxSvc)
	medHandler.RegisterRoutes(apiV1, fhirGroup)

	// Scheduling domain
	schedRepo := scheduling.NewScheduleRepoPG(pool)
	slotRepo := scheduling.NewSlotRepoPG(pool)
	apptRepo := scheduling.NewAppointmentRepoPG(pool)
	apptRespRepo := scheduling.NewAppointmentResponseRepoPG(pool)
	wlRepo := scheduling.NewWaitlistRepoPG(pool)
	schedSvc := scheduling.NewService(schedRepo, slotRepo, apptRepo, apptRespRepo, wlRepo)
	schedSvc.SetVersionTracker(versionTracker)
	schedHandler := scheduling.NewHandler(schedSvc)
	schedHandler.RegisterRoutes(apiV1, fhirGroup)

	// Billing domain
	covRepo := billing.NewCoverageRepoPG(pool)
	claimRepo := billing.NewClaimRepoPG(pool)
	claimRespRepo := billing.NewClaimResponseRepoPG(pool)
	eobRepo := billing.NewExplanationOfBenefitRepoPG(pool)
	invRepo := billing.NewInvoiceRepoPG(pool)
	billSvc := billing.NewService(covRepo, claimRepo, claimRespRepo, eobRepo, invRepo)
	billSvc.SetVersionTracker(versionTracker)
	billHandler := billing.NewHandler(billSvc)
	billHandler.RegisterRoutes(apiV1, fhirGroup)

	// Documents domain (consentRepo created earlier for consent enforcement middleware)
	docRefRepo := documents.NewDocumentReferenceRepoPG(pool)
	noteRepo := documents.NewClinicalNoteRepoPG(pool)
	compRepo := documents.NewCompositionRepoPG(pool)
	docTemplateRepo := documents.NewDocumentTemplateRepoPG(pool)
	docSvc := documents.NewService(consentRepo, docRefRepo, noteRepo, compRepo, docTemplateRepo)
	docSvc.SetVersionTracker(versionTracker)
	docHandler := documents.NewHandler(docSvc)
	docHandler.RegisterRoutes(apiV1, fhirGroup)

	// Inbox domain
	poolRepo := inbox.NewMessagePoolRepoPG(pool)
	msgRepo := inbox.NewInboxMessageRepoPG(pool)
	cosignRepo := inbox.NewCosignRequestRepoPG(pool)
	listRepo := inbox.NewPatientListRepoPG(pool)
	handoffRepo := inbox.NewHandoffRepoPG(pool)
	inboxSvc := inbox.NewService(poolRepo, msgRepo, cosignRepo, listRepo, handoffRepo)
	inboxSvc.SetVersionTracker(versionTracker)
	inboxHandler := inbox.NewHandler(inboxSvc)
	inboxHandler.RegisterRoutes(apiV1, fhirGroup)

	// Surgery domain
	orRoomRepo := surgery.NewORRoomRepoPG(pool)
	caseRepo := surgery.NewSurgicalCaseRepoPG(pool)
	prefCardRepo := surgery.NewPreferenceCardRepoPG(pool)
	implantRepo := surgery.NewImplantLogRepoPG(pool)
	surgerySvc := surgery.NewService(orRoomRepo, caseRepo, prefCardRepo, implantRepo)
	surgerySvc.SetVersionTracker(versionTracker)
	surgeryHandler := surgery.NewHandler(surgerySvc)
	surgeryHandler.RegisterRoutes(apiV1, fhirGroup)

	// Emergency domain
	triageRepo := emergency.NewTriageRepoPG(pool)
	edTrackRepo := emergency.NewEDTrackingRepoPG(pool)
	traumaRepo := emergency.NewTraumaRepoPG(pool)
	edSvc := emergency.NewService(triageRepo, edTrackRepo, traumaRepo)
	edSvc.SetVersionTracker(versionTracker)
	edHandler := emergency.NewHandler(edSvc)
	edHandler.RegisterRoutes(apiV1, fhirGroup)

	// Obstetrics domain
	pregRepo := obstetrics.NewPregnancyRepoPG(pool)
	prenatalRepo := obstetrics.NewPrenatalVisitRepoPG(pool)
	laborRepo := obstetrics.NewLaborRepoPG(pool)
	deliveryRepo := obstetrics.NewDeliveryRepoPG(pool)
	newbornRepo := obstetrics.NewNewbornRepoPG(pool)
	postpartumRepo := obstetrics.NewPostpartumRepoPG(pool)
	obSvc := obstetrics.NewService(pregRepo, prenatalRepo, laborRepo, deliveryRepo, newbornRepo, postpartumRepo)
	obSvc.SetVersionTracker(versionTracker)
	obHandler := obstetrics.NewHandler(obSvc)
	obHandler.RegisterRoutes(apiV1)

	// Oncology domain
	cancerDxRepo := oncology.NewCancerDiagnosisRepoPG(pool)
	protoRepo := oncology.NewTreatmentProtocolRepoPG(pool)
	chemoRepo := oncology.NewChemoCycleRepoPG(pool)
	radRepo := oncology.NewRadiationTherapyRepoPG(pool)
	markerRepo := oncology.NewTumorMarkerRepoPG(pool)
	boardRepo := oncology.NewTumorBoardRepoPG(pool)
	oncoSvc := oncology.NewService(cancerDxRepo, protoRepo, chemoRepo, radRepo, markerRepo, boardRepo)
	oncoSvc.SetVersionTracker(versionTracker)
	oncoHandler := oncology.NewHandler(oncoSvc)
	oncoHandler.RegisterRoutes(apiV1)

	// Nursing domain
	fsTemplateRepo := nursing.NewFlowsheetTemplateRepoPG(pool)
	fsEntryRepo := nursing.NewFlowsheetEntryRepoPG(pool)
	nurseAssessRepo := nursing.NewNursingAssessmentRepoPG(pool)
	fallRiskRepo := nursing.NewFallRiskRepoPG(pool)
	skinRepo := nursing.NewSkinAssessmentRepoPG(pool)
	painRepo := nursing.NewPainAssessmentRepoPG(pool)
	linesRepo := nursing.NewLinesDrainsRepoPG(pool)
	restraintRepo := nursing.NewRestraintRepoPG(pool)
	ioRepo := nursing.NewIntakeOutputRepoPG(pool)
	nurseSvc := nursing.NewService(fsTemplateRepo, fsEntryRepo, nurseAssessRepo, fallRiskRepo, skinRepo, painRepo, linesRepo, restraintRepo, ioRepo)
	nurseSvc.SetVersionTracker(versionTracker)
	nurseHandler := nursing.NewHandler(nurseSvc)
	nurseHandler.RegisterRoutes(apiV1)

	// Behavioral health domain
	psychRepo := behavioral.NewPsychAssessmentRepoPG(pool)
	safetyRepo := behavioral.NewSafetyPlanRepoPG(pool)
	legalRepo := behavioral.NewLegalHoldRepoPG(pool)
	seclusionRepo := behavioral.NewSeclusionRestraintRepoPG(pool)
	groupRepo := behavioral.NewGroupTherapyRepoPG(pool)
	bhSvc := behavioral.NewService(psychRepo, safetyRepo, legalRepo, seclusionRepo, groupRepo)
	bhSvc.SetVersionTracker(versionTracker)
	bhHandler := behavioral.NewHandler(bhSvc)
	bhHandler.RegisterRoutes(apiV1, fhirGroup)

	// Research domain
	studyRepo := research.NewStudyRepoPG(pool)
	enrollRepo := research.NewEnrollmentRepoPG(pool)
	advEventRepo := research.NewAdverseEventRepoPG(pool)
	devRepo := research.NewDeviationRepoPG(pool)
	resSvc := research.NewService(studyRepo, enrollRepo, advEventRepo, devRepo)
	resSvc.SetVersionTracker(versionTracker)
	resHandler := research.NewHandler(resSvc)
	resHandler.RegisterRoutes(apiV1, fhirGroup)

	// Portal domain
	portalAcctRepo := portal.NewPortalAccountRepoPG(pool)
	portalMsgRepo := portal.NewPortalMessageRepoPG(pool)
	questRepo := portal.NewQuestionnaireRepoPG(pool)
	questRespRepo := portal.NewQuestionnaireResponseRepoPG(pool)
	checkinRepo := portal.NewPatientCheckinRepoPG(pool)
	portalSvc := portal.NewService(portalAcctRepo, portalMsgRepo, questRepo, questRespRepo, checkinRepo)
	portalSvc.SetVersionTracker(versionTracker)
	portalHandler := portal.NewHandler(portalSvc)
	portalHandler.RegisterRoutes(apiV1, fhirGroup)

	// Terminology domain
	loincRepo := terminology.NewLOINCRepoPG(pool)
	icd10Repo := terminology.NewICD10RepoPG(pool)
	snomedRepo := terminology.NewSNOMEDRepoPG(pool)
	rxnormRepo := terminology.NewRxNormRepoPG(pool)
	cptRepo := terminology.NewCPTRepoPG(pool)
	termSvc := terminology.NewService(loincRepo, icd10Repo, snomedRepo, rxnormRepo, cptRepo)
	termSvc.SetVersionTracker(versionTracker)
	termHandler := terminology.NewHandler(termSvc)
	termHandler.RegisterRoutes(apiV1, fhirGroup)

	// CDS domain
	cdsRuleRepo := cds.NewCDSRuleRepoPG(pool)
	cdsAlertRepo := cds.NewCDSAlertRepoPG(pool)
	drugIntRepo := cds.NewDrugInteractionRepoPG(pool)
	orderSetRepo := cds.NewOrderSetRepoPG(pool)
	pathwayRepo := cds.NewClinicalPathwayRepoPG(pool)
	pathwayEnrollRepo := cds.NewPatientPathwayEnrollmentRepoPG(pool)
	formularyRepo := cds.NewFormularyRepoPG(pool)
	medReconcRepo := cds.NewMedReconciliationRepoPG(pool)
	cdsSvc := cds.NewService(cdsRuleRepo, cdsAlertRepo, drugIntRepo, orderSetRepo, pathwayRepo, pathwayEnrollRepo, formularyRepo, medReconcRepo)
	cdsSvc.SetVersionTracker(versionTracker)
	cdsHandler := cds.NewHandler(cdsSvc)
	cdsHandler.RegisterRoutes(apiV1)

	// Immunization domain
	immRepo := immunization.NewImmunizationRepoPG(pool)
	immRecRepo := immunization.NewRecommendationRepoPG(pool)
	immSvc := immunization.NewService(immRepo, immRecRepo)
	immSvc.SetVersionTracker(versionTracker)
	immHandler := immunization.NewHandler(immSvc)
	immHandler.RegisterRoutes(apiV1, fhirGroup)

	// CarePlan domain
	cpRepo := careplan.NewCarePlanRepoPG(pool)
	goalRepo := careplan.NewGoalRepoPG(pool)
	cpSvc := careplan.NewService(cpRepo, goalRepo)
	cpSvc.SetVersionTracker(versionTracker)
	cpHandler := careplan.NewHandler(cpSvc)
	cpHandler.RegisterRoutes(apiV1, fhirGroup)

	// FamilyHistory domain
	fmhRepo := familyhistory.NewFamilyMemberHistoryRepoPG(pool)
	fmhSvc := familyhistory.NewService(fmhRepo)
	fmhSvc.SetVersionTracker(versionTracker)
	fmhHandler := familyhistory.NewHandler(fmhSvc)
	fmhHandler.RegisterRoutes(apiV1, fhirGroup)

	// RelatedPerson domain
	rpRepo := relatedperson.NewRelatedPersonRepoPG(pool)
	rpSvc := relatedperson.NewService(rpRepo)
	rpSvc.SetVersionTracker(versionTracker)
	rpHandler := relatedperson.NewHandler(rpSvc)
	rpHandler.RegisterRoutes(apiV1, fhirGroup)

	// Provenance domain
	provRepo := provenance.NewProvenanceRepoPG(pool)
	provSvc := provenance.NewService(provRepo)
	provSvc.SetVersionTracker(versionTracker)
	provHandler := provenance.NewHandler(provSvc)
	provHandler.RegisterRoutes(apiV1, fhirGroup)

	// CareTeam domain
	ctRepo := careteam.NewCareTeamRepoPG(pool)
	ctSvc := careteam.NewService(ctRepo)
	ctSvc.SetVersionTracker(versionTracker)
	ctHandler := careteam.NewHandler(ctSvc)
	ctHandler.RegisterRoutes(apiV1, fhirGroup)

	// Task domain
	taskRepo := fhirtask.NewTaskRepoPG(pool)
	taskSvc := fhirtask.NewService(taskRepo)
	taskSvc.SetVersionTracker(versionTracker)
	taskHandler := fhirtask.NewHandler(taskSvc)
	taskHandler.RegisterRoutes(apiV1, fhirGroup)

	// Device domain
	deviceRepo := device.NewDeviceRepoPG(pool)
	devSvc := device.NewService(deviceRepo)
	devSvc.SetVersionTracker(versionTracker)
	devHandler := device.NewHandler(devSvc)
	devHandler.RegisterRoutes(apiV1, fhirGroup)

	// Subscription domain
	subRepo := subscription.NewSubscriptionRepoPG(pool)
	subSvc := subscription.NewService(subRepo)
	subSvc.SetVersionTracker(versionTracker)
	subHandler := subscription.NewHandler(subSvc)
	subHandler.RegisterRoutes(apiV1, fhirGroup)

	// Clinical Safety domain (Flag, DetectedIssue, AdverseEvent, ClinicalImpression, RiskAssessment)
	flagRepo := clinical.NewFlagRepoPG(pool)
	detectedIssueRepo := clinical.NewDetectedIssueRepoPG(pool)
	adverseEventRepo := clinical.NewAdverseEventRepoPG(pool)
	clinicalImpressionRepo := clinical.NewClinicalImpressionRepoPG(pool)
	riskAssessmentRepo := clinical.NewRiskAssessmentRepoPG(pool)
	clinicalSafetySvc := clinical.NewClinicalSafetyService(flagRepo, detectedIssueRepo, adverseEventRepo, clinicalImpressionRepo, riskAssessmentRepo)
	clinicalSafetySvc.SetVersionTracker(versionTracker)
	clinicalSafetyHandler := clinical.NewClinicalSafetyHandler(clinicalSafetySvc)
	clinicalSafetyHandler.RegisterRoutes(apiV1, fhirGroup)

	// EpisodeOfCare domain
	eocRepo := episodeofcare.NewEpisodeOfCareRepoPG(pool)
	eocSvc := episodeofcare.NewService(eocRepo)
	eocSvc.SetVersionTracker(versionTracker)
	eocHandler := episodeofcare.NewHandler(eocSvc)
	eocHandler.RegisterRoutes(apiV1, fhirGroup)

	// HealthcareService domain
	hcsRepo := healthcareservice.NewHealthcareServiceRepoPG(pool)
	hcsSvc := healthcareservice.NewService(hcsRepo)
	hcsSvc.SetVersionTracker(versionTracker)
	hcsHandler := healthcareservice.NewHandler(hcsSvc)
	hcsHandler.RegisterRoutes(apiV1, fhirGroup)

	// MeasureReport domain
	mrRepo := measurereport.NewMeasureReportRepoPG(pool)
	mrSvc := measurereport.NewService(mrRepo)
	mrSvc.SetVersionTracker(versionTracker)
	mrHandler := measurereport.NewHandler(mrSvc)
	mrHandler.RegisterRoutes(apiV1, fhirGroup)

	// FHIRList domain
	fhirListRepo := fhirlist.NewFHIRListRepoPG(pool)
	fhirListSvc := fhirlist.NewService(fhirListRepo)
	fhirListSvc.SetVersionTracker(versionTracker)
	fhirListHandler := fhirlist.NewHandler(fhirListSvc)
	fhirListHandler.RegisterRoutes(apiV1, fhirGroup)

	// Financial domain
	accountRepo := financial.NewAccountRepoPG(pool)
	insurancePlanRepo := financial.NewInsurancePlanRepoPG(pool)
	paymentNoticeRepo := financial.NewPaymentNoticeRepoPG(pool)
	paymentReconciliationRepo := financial.NewPaymentReconciliationRepoPG(pool)
	chargeItemRepo := financial.NewChargeItemRepoPG(pool)
	chargeItemDefRepo := financial.NewChargeItemDefinitionRepoPG(pool)
	contractRepo := financial.NewContractRepoPG(pool)
	enrollReqRepo := financial.NewEnrollmentRequestRepoPG(pool)
	enrollRespRepo := financial.NewEnrollmentResponseRepoPG(pool)
	financialSvc := financial.NewService(accountRepo, insurancePlanRepo, paymentNoticeRepo, paymentReconciliationRepo, chargeItemRepo, chargeItemDefRepo, contractRepo, enrollReqRepo, enrollRespRepo)
	financialSvc.SetVersionTracker(versionTracker)
	financialHandler := financial.NewHandler(financialSvc)
	financialHandler.RegisterRoutes(apiV1, fhirGroup)

	// Workflow domain (ActivityDefinition, RequestGroup, GuidanceResponse)
	actDefRepo := workflow.NewActivityDefinitionRepoPG(pool)
	reqGrpRepo := workflow.NewRequestGroupRepoPG(pool)
	guidRespRepo := workflow.NewGuidanceResponseRepoPG(pool)
	workflowSvc := workflow.NewService(actDefRepo, reqGrpRepo, guidRespRepo)
	workflowSvc.SetVersionTracker(versionTracker)
	workflowHandler := workflow.NewHandler(workflowSvc)
	workflowHandler.RegisterRoutes(apiV1, fhirGroup)

	// Supply domain
	supplyReqRepo := supply.NewSupplyRequestRepoPG(pool)
	supplyDelRepo := supply.NewSupplyDeliveryRepoPG(pool)
	supplySvc := supply.NewService(supplyReqRepo, supplyDelRepo)
	supplySvc.SetVersionTracker(versionTracker)
	supplyHandler := supply.NewHandler(supplySvc)
	supplyHandler.RegisterRoutes(apiV1, fhirGroup)

	// Conformance domain (NamingSystem, OperationDefinition, MessageDefinition, MessageHeader)
	namingSysRepo := conformance.NewNamingSystemRepoPG(pool)
	opDefRepo := conformance.NewOperationDefinitionRepoPG(pool)
	msgDefRepo := conformance.NewMessageDefinitionRepoPG(pool)
	msgHeaderRepo := conformance.NewMessageHeaderRepoPG(pool)
	conformanceSvc := conformance.NewService(namingSysRepo, opDefRepo, msgDefRepo, msgHeaderRepo)
	conformanceSvc.SetVersionTracker(versionTracker)
	conformanceHandler := conformance.NewHandler(conformanceSvc)
	conformanceHandler.RegisterRoutes(apiV1, fhirGroup)

	// VisionPrescription domain
	vpRepo := visionprescription.NewVisionPrescriptionRepoPG(pool)
	vpSvc := visionprescription.NewService(vpRepo)
	vpSvc.SetVersionTracker(versionTracker)
	vpHandler := visionprescription.NewHandler(vpSvc)
	vpHandler.RegisterRoutes(apiV1, fhirGroup)

	// Endpoint domain
	endpointRepo := fhirendpoint.NewEndpointRepoPG(pool)
	endpointSvc := fhirendpoint.NewService(endpointRepo)
	endpointSvc.SetVersionTracker(versionTracker)
	endpointHandler := fhirendpoint.NewHandler(endpointSvc)
	endpointHandler.RegisterRoutes(apiV1, fhirGroup)

	// BodyStructure domain
	bsRepo := bodystructure.NewBodyStructureRepoPG(pool)
	bsSvc := bodystructure.NewService(bsRepo)
	bsSvc.SetVersionTracker(versionTracker)
	bsHandler := bodystructure.NewHandler(bsSvc)
	bsHandler.RegisterRoutes(apiV1, fhirGroup)

	// Substance domain
	substanceRepo := substance.NewSubstanceRepoPG(pool)
	substanceSvc := substance.NewService(substanceRepo)
	substanceSvc.SetVersionTracker(versionTracker)
	substanceHandler := substance.NewHandler(substanceSvc)
	substanceHandler.RegisterRoutes(apiV1, fhirGroup)

	// Media domain
	mediaRepo := fhirmedia.NewMediaRepoPG(pool)
	mediaSvc := fhirmedia.NewService(mediaRepo)
	mediaSvc.SetVersionTracker(versionTracker)
	mediaHandler := fhirmedia.NewHandler(mediaSvc)
	mediaHandler.RegisterRoutes(apiV1, fhirGroup)

	// DeviceRequest domain
	devReqRepo := devicerequest.NewDeviceRequestRepoPG(pool)
	devReqSvc := devicerequest.NewService(devReqRepo)
	devReqSvc.SetVersionTracker(versionTracker)
	devReqHandler := devicerequest.NewHandler(devReqSvc)
	devReqHandler.RegisterRoutes(apiV1, fhirGroup)

	// DeviceUseStatement domain
	dusRepo := deviceusestatement.NewDeviceUseStatementRepoPG(pool)
	dusSvc := deviceusestatement.NewService(dusRepo)
	dusSvc.SetVersionTracker(versionTracker)
	dusHandler := deviceusestatement.NewHandler(dusSvc)
	dusHandler.RegisterRoutes(apiV1, fhirGroup)

	// CoverageEligibility domain
	ceReqRepo := coverageeligibility.NewCoverageEligibilityRequestRepoPG(pool)
	ceRespRepo := coverageeligibility.NewCoverageEligibilityResponseRepoPG(pool)
	ceSvc := coverageeligibility.NewService(ceReqRepo, ceRespRepo)
	ceSvc.SetVersionTracker(versionTracker)
	ceHandler := coverageeligibility.NewHandler(ceSvc)
	ceHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicationKnowledge domain
	mkRepo := medicationknowledge.NewMedicationKnowledgeRepoPG(pool)
	mkSvc := medicationknowledge.NewService(mkRepo)
	mkSvc.SetVersionTracker(versionTracker)
	mkHandler := medicationknowledge.NewHandler(mkSvc)
	mkHandler.RegisterRoutes(apiV1, fhirGroup)

	// OrganizationAffiliation domain
	oaRepo := organizationaffiliation.NewOrganizationAffiliationRepoPG(pool)
	oaSvc := organizationaffiliation.NewService(oaRepo)
	oaSvc.SetVersionTracker(versionTracker)
	oaHandler := organizationaffiliation.NewHandler(oaSvc)
	oaHandler.RegisterRoutes(apiV1, fhirGroup)

	// Person domain
	personRepo := person.NewPersonRepoPG(pool)
	personSvc := person.NewService(personRepo)
	personSvc.SetVersionTracker(versionTracker)
	personHandler := person.NewHandler(personSvc)
	personHandler.RegisterRoutes(apiV1, fhirGroup)

	// Measure domain
	fhirMeasureRepo := fhirmeasure.NewMeasureRepoPG(pool)
	fhirMeasureSvc := fhirmeasure.NewService(fhirMeasureRepo)
	fhirMeasureSvc.SetVersionTracker(versionTracker)
	fhirMeasureHandler := fhirmeasure.NewHandler(fhirMeasureSvc)
	fhirMeasureHandler.RegisterRoutes(apiV1, fhirGroup)

	// Library domain
	libraryRepo := fhirlibrary.NewLibraryRepoPG(pool)
	librarySvc := fhirlibrary.NewService(libraryRepo)
	librarySvc.SetVersionTracker(versionTracker)
	libraryHandler := fhirlibrary.NewHandler(librarySvc)
	libraryHandler.RegisterRoutes(apiV1, fhirGroup)

	// DeviceDefinition domain
	ddRepo := devicedefinition.NewDeviceDefinitionRepoPG(pool)
	ddSvc := devicedefinition.NewService(ddRepo)
	ddSvc.SetVersionTracker(versionTracker)
	ddHandler := devicedefinition.NewHandler(ddSvc)
	ddHandler.RegisterRoutes(apiV1, fhirGroup)

	// DeviceMetric domain
	dmRepo := devicemetric.NewDeviceMetricRepoPG(pool)
	dmSvc := devicemetric.NewService(dmRepo)
	dmSvc.SetVersionTracker(versionTracker)
	dmHandler := devicemetric.NewHandler(dmSvc)
	dmHandler.RegisterRoutes(apiV1, fhirGroup)

	// SpecimenDefinition domain
	sdRepo := specimendefinition.NewSpecimenDefinitionRepoPG(pool)
	sdSvc := specimendefinition.NewService(sdRepo)
	sdSvc.SetVersionTracker(versionTracker)
	sdHandler := specimendefinition.NewHandler(sdSvc)
	sdHandler.RegisterRoutes(apiV1, fhirGroup)

	// CommunicationRequest domain
	commReqRepo := communicationrequest.NewCommunicationRequestRepoPG(pool)
	commReqSvc := communicationrequest.NewService(commReqRepo)
	commReqSvc.SetVersionTracker(versionTracker)
	commReqHandler := communicationrequest.NewHandler(commReqSvc)
	commReqHandler.RegisterRoutes(apiV1, fhirGroup)

	// ObservationDefinition domain
	obsDefRepo := observationdefinition.NewObservationDefinitionRepoPG(pool)
	obsDefSvc := observationdefinition.NewService(obsDefRepo)
	obsDefSvc.SetVersionTracker(versionTracker)
	obsDefHandler := observationdefinition.NewHandler(obsDefSvc)
	obsDefHandler.RegisterRoutes(apiV1, fhirGroup)

	// Linkage domain
	linkageRepo := linkage.NewLinkageRepoPG(pool)
	linkageSvc := linkage.NewService(linkageRepo)
	linkageSvc.SetVersionTracker(versionTracker)
	linkageHandler := linkage.NewHandler(linkageSvc)
	linkageHandler.RegisterRoutes(apiV1, fhirGroup)

	// Basic domain
	basicRepo := fhirbasic.NewBasicRepoPG(pool)
	basicSvc := fhirbasic.NewService(basicRepo)
	basicSvc.SetVersionTracker(versionTracker)
	basicHandler := fhirbasic.NewHandler(basicSvc)
	basicHandler.RegisterRoutes(apiV1, fhirGroup)

	// VerificationResult domain
	vrRepo := verificationresult.NewVerificationResultRepoPG(pool)
	vrSvc := verificationresult.NewService(vrRepo)
	vrSvc.SetVersionTracker(versionTracker)
	vrHandler := verificationresult.NewHandler(vrSvc)
	vrHandler.RegisterRoutes(apiV1, fhirGroup)

	// EventDefinition domain
	edRepo := eventdefinition.NewEventDefinitionRepoPG(pool)
	edDefSvc := eventdefinition.NewService(edRepo)
	edDefSvc.SetVersionTracker(versionTracker)
	edDefHandler := eventdefinition.NewHandler(edDefSvc)
	edDefHandler.RegisterRoutes(apiV1, fhirGroup)

	// GraphDefinition domain
	gdRepo := graphdefinition.NewGraphDefinitionRepoPG(pool)
	gdSvc := graphdefinition.NewService(gdRepo)
	gdSvc.SetVersionTracker(versionTracker)
	gdHandler := graphdefinition.NewHandler(gdSvc)
	gdHandler.RegisterRoutes(apiV1, fhirGroup)

	// MolecularSequence domain
	msRepo := molecularsequence.NewMolecularSequenceRepoPG(pool)
	msSvc := molecularsequence.NewService(msRepo)
	msSvc.SetVersionTracker(versionTracker)
	msHandler := molecularsequence.NewHandler(msSvc)
	msHandler.RegisterRoutes(apiV1, fhirGroup)

	// BiologicallyDerivedProduct domain
	bdpRepo := biologicallyderivedproduct.NewBiologicallyDerivedProductRepoPG(pool)
	bdpSvc := biologicallyderivedproduct.NewService(bdpRepo)
	bdpSvc.SetVersionTracker(versionTracker)
	bdpHandler := biologicallyderivedproduct.NewHandler(bdpSvc)
	bdpHandler.RegisterRoutes(apiV1, fhirGroup)

	// CatalogEntry domain
	catRepo := catalogentry.NewCatalogEntryRepoPG(pool)
	catSvc := catalogentry.NewService(catRepo)
	catSvc.SetVersionTracker(versionTracker)
	catHandler := catalogentry.NewHandler(catSvc)
	catHandler.RegisterRoutes(apiV1, fhirGroup)

	// StructureDefinition domain
	structDefRepo := structuredefinition.NewStructureDefinitionRepoPG(pool)
	structDefSvc := structuredefinition.NewService(structDefRepo)
	structDefSvc.SetVersionTracker(versionTracker)
	structDefHandler := structuredefinition.NewHandler(structDefSvc)
	structDefHandler.RegisterRoutes(apiV1, fhirGroup)

	// SearchParameter domain
	spRepo := searchparameter.NewSearchParameterRepoPG(pool)
	spSvc := searchparameter.NewService(spRepo)
	spSvc.SetVersionTracker(versionTracker)
	spHandler := searchparameter.NewHandler(spSvc)
	spHandler.RegisterRoutes(apiV1, fhirGroup)

	// CodeSystem domain
	csRepo := codesystem.NewCodeSystemRepoPG(pool)
	csSvc := codesystem.NewService(csRepo)
	csSvc.SetVersionTracker(versionTracker)
	csHandler := codesystem.NewHandler(csSvc)
	csHandler.RegisterRoutes(apiV1, fhirGroup)

	// ValueSet domain
	vsRepo := valueset.NewValueSetRepoPG(pool)
	vsSvc := valueset.NewService(vsRepo)
	vsSvc.SetVersionTracker(versionTracker)
	vsHandler := valueset.NewHandler(vsSvc)
	vsHandler.RegisterRoutes(apiV1, fhirGroup)

	// ConceptMap domain
	cmRepo := conceptmap.NewConceptMapRepoPG(pool)
	cmSvc := conceptmap.NewService(cmRepo)
	cmSvc.SetVersionTracker(versionTracker)
	cmHandler := conceptmap.NewHandler(cmSvc)
	cmHandler.RegisterRoutes(apiV1, fhirGroup)

	// ImplementationGuide domain
	igRepo := implementationguide.NewImplementationGuideRepoPG(pool)
	igSvc := implementationguide.NewService(igRepo)
	igSvc.SetVersionTracker(versionTracker)
	igHandler := implementationguide.NewHandler(igSvc)
	igHandler.RegisterRoutes(apiV1, fhirGroup)

	// CompartmentDefinition domain
	cdRepo := compartmentdefinition.NewCompartmentDefinitionRepoPG(pool)
	cdSvc := compartmentdefinition.NewService(cdRepo)
	cdSvc.SetVersionTracker(versionTracker)
	cdHandler := compartmentdefinition.NewHandler(cdSvc)
	cdHandler.RegisterRoutes(apiV1, fhirGroup)

	// TerminologyCapabilities domain
	tcRepo := terminologycapabilities.NewTerminologyCapabilitiesRepoPG(pool)
	tcSvc := terminologycapabilities.NewService(tcRepo)
	tcSvc.SetVersionTracker(versionTracker)
	tcHandler := terminologycapabilities.NewHandler(tcSvc)
	tcHandler.RegisterRoutes(apiV1, fhirGroup)

	// StructureMap domain
	smRepo := structuremap.NewStructureMapRepoPG(pool)
	smSvc := structuremap.NewService(smRepo)
	smSvc.SetVersionTracker(versionTracker)
	smHandler := structuremap.NewHandler(smSvc)
	smHandler.RegisterRoutes(apiV1, fhirGroup)

	// TestScript domain
	tsRepo := testscript.NewTestScriptRepoPG(pool)
	tsSvc := testscript.NewService(tsRepo)
	tsSvc.SetVersionTracker(versionTracker)
	tsHandler := testscript.NewHandler(tsSvc)
	tsHandler.RegisterRoutes(apiV1, fhirGroup)

	// TestReport domain
	trRepo := testreport.NewTestReportRepoPG(pool)
	trSvc := testreport.NewService(trRepo)
	trSvc.SetVersionTracker(versionTracker)
	trHandler := testreport.NewHandler(trSvc)
	trHandler.RegisterRoutes(apiV1, fhirGroup)

	// ExampleScenario domain
	esRepo := examplescenario.NewExampleScenarioRepoPG(pool)
	esSvc := examplescenario.NewService(esRepo)
	esSvc.SetVersionTracker(versionTracker)
	esHandler := examplescenario.NewHandler(esSvc)
	esHandler.RegisterRoutes(apiV1, fhirGroup)

	// Evidence domain
	evRepo := fhirevidence.NewEvidenceRepoPG(pool)
	evSvc := fhirevidence.NewService(evRepo)
	evSvc.SetVersionTracker(versionTracker)
	evHandler := fhirevidence.NewHandler(evSvc)
	evHandler.RegisterRoutes(apiV1, fhirGroup)

	// EvidenceVariable domain
	evvRepo := evidencevariable.NewEvidenceVariableRepoPG(pool)
	evvSvc := evidencevariable.NewService(evvRepo)
	evvSvc.SetVersionTracker(versionTracker)
	evvHandler := evidencevariable.NewHandler(evvSvc)
	evvHandler.RegisterRoutes(apiV1, fhirGroup)

	// ResearchDefinition domain
	rdRepo := researchdefinition.NewResearchDefinitionRepoPG(pool)
	rdSvc := researchdefinition.NewService(rdRepo)
	rdSvc.SetVersionTracker(versionTracker)
	rdHandler := researchdefinition.NewHandler(rdSvc)
	rdHandler.RegisterRoutes(apiV1, fhirGroup)

	// ResearchElementDefinition domain
	redRepo := researchelementdefinition.NewResearchElementDefinitionRepoPG(pool)
	redSvc := researchelementdefinition.NewService(redRepo)
	redSvc.SetVersionTracker(versionTracker)
	redHandler := researchelementdefinition.NewHandler(redSvc)
	redHandler.RegisterRoutes(apiV1, fhirGroup)

	// EffectEvidenceSynthesis domain
	eesRepo := effectevidencesynthesis.NewEffectEvidenceSynthesisRepoPG(pool)
	eesSvc := effectevidencesynthesis.NewService(eesRepo)
	eesSvc.SetVersionTracker(versionTracker)
	eesHandler := effectevidencesynthesis.NewHandler(eesSvc)
	eesHandler.RegisterRoutes(apiV1, fhirGroup)

	// RiskEvidenceSynthesis domain
	resRepo := riskevidencesynthesis.NewRiskEvidenceSynthesisRepoPG(pool)
	resSynSvc := riskevidencesynthesis.NewService(resRepo)
	resSynSvc.SetVersionTracker(versionTracker)
	resSynHandler := riskevidencesynthesis.NewHandler(resSynSvc)
	resSynHandler.RegisterRoutes(apiV1, fhirGroup)

	// ResearchSubject domain
	rsRepo := researchsubject.NewResearchSubjectRepoPG(pool)
	rsSvc := researchsubject.NewService(rsRepo)
	rsSvc.SetVersionTracker(versionTracker)
	rsHandler := researchsubject.NewHandler(rsSvc)
	rsHandler.RegisterRoutes(apiV1, fhirGroup)

	// DocumentManifest domain
	docManRepo := documentmanifest.NewDocumentManifestRepoPG(pool)
	docManSvc := documentmanifest.NewService(docManRepo)
	docManSvc.SetVersionTracker(versionTracker)
	docManHandler := documentmanifest.NewHandler(docManSvc)
	docManHandler.RegisterRoutes(apiV1, fhirGroup)

	// SubstanceSpecification domain
	ssRepo := substancespecification.NewSubstanceSpecificationRepoPG(pool)
	ssSvc := substancespecification.NewService(ssRepo)
	ssSvc.SetVersionTracker(versionTracker)
	ssHandler := substancespecification.NewHandler(ssSvc)
	ssHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProduct domain
	mpRepo := medicinalproduct.NewMedicinalProductRepoPG(pool)
	mpSvc := medicinalproduct.NewService(mpRepo)
	mpSvc.SetVersionTracker(versionTracker)
	mpHandler := medicinalproduct.NewHandler(mpSvc)
	mpHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductIngredient domain
	mpiRepo := medproductingredient.NewMedicinalProductIngredientRepoPG(pool)
	mpiSvc := medproductingredient.NewService(mpiRepo)
	mpiSvc.SetVersionTracker(versionTracker)
	mpiHandler := medproductingredient.NewHandler(mpiSvc)
	mpiHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductManufactured domain
	mpmRepo := medproductmanufactured.NewMedicinalProductManufacturedRepoPG(pool)
	mpmSvc := medproductmanufactured.NewService(mpmRepo)
	mpmSvc.SetVersionTracker(versionTracker)
	mpmHandler := medproductmanufactured.NewHandler(mpmSvc)
	mpmHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductPackaged domain
	mppRepo := medproductpackaged.NewMedicinalProductPackagedRepoPG(pool)
	mppSvc := medproductpackaged.NewService(mppRepo)
	mppSvc.SetVersionTracker(versionTracker)
	mppHandler := medproductpackaged.NewHandler(mppSvc)
	mppHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductAuthorization domain
	mpaRepo := medproductauthorization.NewMedicinalProductAuthorizationRepoPG(pool)
	mpaSvc := medproductauthorization.NewService(mpaRepo)
	mpaSvc.SetVersionTracker(versionTracker)
	mpaHandler := medproductauthorization.NewHandler(mpaSvc)
	mpaHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductContraindication domain
	mpcRepo := medproductcontraindication.NewMedicinalProductContraindicationRepoPG(pool)
	mpcSvc := medproductcontraindication.NewService(mpcRepo)
	mpcSvc.SetVersionTracker(versionTracker)
	mpcHandler := medproductcontraindication.NewHandler(mpcSvc)
	mpcHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductIndication domain
	mpindRepo := medproductindication.NewMedicinalProductIndicationRepoPG(pool)
	mpindSvc := medproductindication.NewService(mpindRepo)
	mpindSvc.SetVersionTracker(versionTracker)
	mpindHandler := medproductindication.NewHandler(mpindSvc)
	mpindHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductInteraction domain
	mpixRepo := medproductinteraction.NewMedicinalProductInteractionRepoPG(pool)
	mpixSvc := medproductinteraction.NewService(mpixRepo)
	mpixSvc.SetVersionTracker(versionTracker)
	mpixHandler := medproductinteraction.NewHandler(mpixSvc)
	mpixHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductUndesirableEffect domain
	mpueRepo := medproductundesirableeffect.NewMedicinalProductUndesirableEffectRepoPG(pool)
	mpueSvc := medproductundesirableeffect.NewService(mpueRepo)
	mpueSvc.SetVersionTracker(versionTracker)
	mpueHandler := medproductundesirableeffect.NewHandler(mpueSvc)
	mpueHandler.RegisterRoutes(apiV1, fhirGroup)

	// MedicinalProductPharmaceutical domain
	mpphRepo := medproductpharmaceutical.NewMedicinalProductPharmaceuticalRepoPG(pool)
	mpphSvc := medproductpharmaceutical.NewService(mpphRepo)
	mpphSvc.SetVersionTracker(versionTracker)
	mpphHandler := medproductpharmaceutical.NewHandler(mpphSvc)
	mpphHandler.RegisterRoutes(apiV1, fhirGroup)

	// SubstancePolymer domain
	subPolyRepo := substancepolymer.NewSubstancePolymerRepoPG(pool)
	subPolySvc := substancepolymer.NewService(subPolyRepo)
	subPolySvc.SetVersionTracker(versionTracker)
	subPolyHandler := substancepolymer.NewHandler(subPolySvc)
	subPolyHandler.RegisterRoutes(apiV1, fhirGroup)

	// SubstanceProtein domain
	sprRepo := substanceprotein.NewSubstanceProteinRepoPG(pool)
	sprSvc := substanceprotein.NewService(sprRepo)
	sprSvc.SetVersionTracker(versionTracker)
	sprHandler := substanceprotein.NewHandler(sprSvc)
	sprHandler.RegisterRoutes(apiV1, fhirGroup)

	// SubstanceNucleicAcid domain
	snaRepo := substancenucleicacid.NewSubstanceNucleicAcidRepoPG(pool)
	snaSvc := substancenucleicacid.NewService(snaRepo)
	snaSvc.SetVersionTracker(versionTracker)
	snaHandler := substancenucleicacid.NewHandler(snaSvc)
	snaHandler.RegisterRoutes(apiV1, fhirGroup)

	// SubstanceSourceMaterial domain
	ssmRepo := substancesourcematerial.NewSubstanceSourceMaterialRepoPG(pool)
	ssmSvc := substancesourcematerial.NewService(ssmRepo)
	ssmSvc.SetVersionTracker(versionTracker)
	ssmHandler := substancesourcematerial.NewHandler(ssmSvc)
	ssmHandler.RegisterRoutes(apiV1, fhirGroup)

	// SubstanceReferenceInformation domain
	sriRepo := substancereferenceinformation.NewSubstanceReferenceInformationRepoPG(pool)
	sriSvc := substancereferenceinformation.NewService(sriRepo)
	sriSvc.SetVersionTracker(versionTracker)
	sriHandler := substancereferenceinformation.NewHandler(sriSvc)
	sriHandler.RegisterRoutes(apiV1, fhirGroup)

	// AuditEvent domain (read-only FHIR endpoints for existing audit_event table)
	aeRepo := auditevent.NewAuditEventRepoPG(pool)
	aeSvc := auditevent.NewService(aeRepo)
	aeHandler := auditevent.NewHandler(aeSvc)
	aeHandler.RegisterRoutes(apiV1, fhirGroup)

	// ImmunizationEvaluation domain
	ieRepo := immunizationevaluation.NewImmunizationEvaluationRepoPG(pool)
	ieSvc := immunizationevaluation.NewService(ieRepo)
	ieSvc.SetVersionTracker(versionTracker)
	ieHandler := immunizationevaluation.NewHandler(ieSvc)
	ieHandler.RegisterRoutes(apiV1, fhirGroup)

	// Notification engine — listens for resource events and delivers webhooks
	notifyAdapter := subscription.NewNotifyRepoAdapter(subRepo)
	notifyEngine := fhir.NewNotificationEngine(notifyAdapter, logger)
	versionTracker.AddListener(notifyEngine)
	notifyCtx, notifyCancel := context.WithCancel(ctx)
	defer notifyCancel()
	go notifyEngine.Start(notifyCtx)

	// -- Register resource fetchers for _include/_revinclude resolution --
	// Each fetcher retrieves a resource by its FHIR ID and returns the FHIR map.
	includeRegistry.RegisterFetcher("Patient", func(ctx context.Context, id string) (map[string]interface{}, error) {
		p, err := identitySvc.GetPatientByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return p.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Practitioner", func(ctx context.Context, id string) (map[string]interface{}, error) {
		p, err := identitySvc.GetPractitionerByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return p.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Organization", func(ctx context.Context, id string) (map[string]interface{}, error) {
		o, err := adminSvc.GetOrganizationByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return o.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Location", func(ctx context.Context, id string) (map[string]interface{}, error) {
		l, err := adminSvc.GetLocationByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return l.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Encounter", func(ctx context.Context, id string) (map[string]interface{}, error) {
		enc, err := encSvc.GetEncounterByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return enc.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Condition", func(ctx context.Context, id string) (map[string]interface{}, error) {
		c, err := clinicalSvc.GetConditionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return c.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Observation", func(ctx context.Context, id string) (map[string]interface{}, error) {
		o, err := clinicalSvc.GetObservationByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return o.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("AllergyIntolerance", func(ctx context.Context, id string) (map[string]interface{}, error) {
		a, err := clinicalSvc.GetAllergyByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return a.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Procedure", func(ctx context.Context, id string) (map[string]interface{}, error) {
		p, err := clinicalSvc.GetProcedureByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return p.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("MedicationRequest", func(ctx context.Context, id string) (map[string]interface{}, error) {
		mr, err := medSvc.GetMedicationRequestByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return mr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("MedicationAdministration", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ma, err := medSvc.GetMedicationAdministrationByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ma.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("MedicationDispense", func(ctx context.Context, id string) (map[string]interface{}, error) {
		md, err := medSvc.GetMedicationDispenseByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return md.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("ServiceRequest", func(ctx context.Context, id string) (map[string]interface{}, error) {
		sr, err := dxSvc.GetServiceRequestByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return sr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("DiagnosticReport", func(ctx context.Context, id string) (map[string]interface{}, error) {
		dr, err := dxSvc.GetDiagnosticReportByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return dr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Specimen", func(ctx context.Context, id string) (map[string]interface{}, error) {
		sp, err := dxSvc.GetSpecimenByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return sp.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("ImagingStudy", func(ctx context.Context, id string) (map[string]interface{}, error) {
		is, err := dxSvc.GetImagingStudyByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return is.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Coverage", func(ctx context.Context, id string) (map[string]interface{}, error) {
		cov, err := billSvc.GetCoverageByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return cov.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Claim", func(ctx context.Context, id string) (map[string]interface{}, error) {
		cl, err := billSvc.GetClaimByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return cl.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Consent", func(ctx context.Context, id string) (map[string]interface{}, error) {
		con, err := docSvc.GetConsentByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return con.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("DocumentReference", func(ctx context.Context, id string) (map[string]interface{}, error) {
		d, err := docSvc.GetDocumentReferenceByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return d.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Composition", func(ctx context.Context, id string) (map[string]interface{}, error) {
		comp, err := docSvc.GetCompositionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return comp.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Appointment", func(ctx context.Context, id string) (map[string]interface{}, error) {
		a, err := schedSvc.GetAppointmentByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return a.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Immunization", func(ctx context.Context, id string) (map[string]interface{}, error) {
		im, err := immSvc.GetImmunizationByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return im.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("ImmunizationRecommendation", func(ctx context.Context, id string) (map[string]interface{}, error) {
		r, err := immSvc.GetRecommendationByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return r.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("CarePlan", func(ctx context.Context, id string) (map[string]interface{}, error) {
		cp, err := cpSvc.GetCarePlanByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return cp.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Goal", func(ctx context.Context, id string) (map[string]interface{}, error) {
		g, err := cpSvc.GetGoalByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return g.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("FamilyMemberHistory", func(ctx context.Context, id string) (map[string]interface{}, error) {
		f, err := fmhSvc.GetFamilyMemberHistoryByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return f.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("RelatedPerson", func(ctx context.Context, id string) (map[string]interface{}, error) {
		relPerson, err := rpSvc.GetRelatedPersonByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return relPerson.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Provenance", func(ctx context.Context, id string) (map[string]interface{}, error) {
		prov, err := provSvc.GetProvenanceByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return prov.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("CareTeam", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ct, err := ctSvc.GetCareTeamByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ct.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Task", func(ctx context.Context, id string) (map[string]interface{}, error) {
		t, err := taskSvc.GetTaskByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return t.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Device", func(ctx context.Context, id string) (map[string]interface{}, error) {
		d, err := devSvc.GetDeviceByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return d.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Subscription", func(ctx context.Context, id string) (map[string]interface{}, error) {
		s, err := subSvc.GetSubscriptionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return s.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Flag", func(ctx context.Context, id string) (map[string]interface{}, error) {
		f, err := clinicalSafetySvc.GetFlagByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return f.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("DetectedIssue", func(ctx context.Context, id string) (map[string]interface{}, error) {
		d, err := clinicalSafetySvc.GetDetectedIssueByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return d.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("AdverseEvent", func(ctx context.Context, id string) (map[string]interface{}, error) {
		a, err := clinicalSafetySvc.GetAdverseEventByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return a.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("ClinicalImpression", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ci, err := clinicalSafetySvc.GetClinicalImpressionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ci.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("RiskAssessment", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ra, err := clinicalSafetySvc.GetRiskAssessmentByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ra.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("EpisodeOfCare", func(ctx context.Context, id string) (map[string]interface{}, error) {
		e, err := eocSvc.GetEpisodeOfCareByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return e.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("HealthcareService", func(ctx context.Context, id string) (map[string]interface{}, error) {
		h, err := hcsSvc.GetHealthcareServiceByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return h.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("MeasureReport", func(ctx context.Context, id string) (map[string]interface{}, error) {
		mr, err := mrSvc.GetMeasureReportByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return mr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("List", func(ctx context.Context, id string) (map[string]interface{}, error) {
		l, err := fhirListSvc.GetFHIRListByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return l.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Account", func(ctx context.Context, id string) (map[string]interface{}, error) {
		a, err := financialSvc.GetAccountByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return a.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("InsurancePlan", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ip, err := financialSvc.GetInsurancePlanByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ip.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("ActivityDefinition", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ad, err := workflowSvc.GetActivityDefinitionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ad.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("RequestGroup", func(ctx context.Context, id string) (map[string]interface{}, error) {
		rg, err := workflowSvc.GetRequestGroupByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return rg.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("GuidanceResponse", func(ctx context.Context, id string) (map[string]interface{}, error) {
		gr, err := workflowSvc.GetGuidanceResponseByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return gr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("SupplyRequest", func(ctx context.Context, id string) (map[string]interface{}, error) {
		sr, err := supplySvc.GetSupplyRequestByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return sr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("SupplyDelivery", func(ctx context.Context, id string) (map[string]interface{}, error) {
		sd, err := supplySvc.GetSupplyDeliveryByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return sd.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("NamingSystem", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ns, err := conformanceSvc.GetNamingSystemByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ns.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("OperationDefinition", func(ctx context.Context, id string) (map[string]interface{}, error) {
		od, err := conformanceSvc.GetOperationDefinitionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return od.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("MessageDefinition", func(ctx context.Context, id string) (map[string]interface{}, error) {
		md, err := conformanceSvc.GetMessageDefinitionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return md.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("MessageHeader", func(ctx context.Context, id string) (map[string]interface{}, error) {
		mh, err := conformanceSvc.GetMessageHeaderByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return mh.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("VisionPrescription", func(ctx context.Context, id string) (map[string]interface{}, error) {
		vp, err := vpSvc.GetVisionPrescriptionByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return vp.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Endpoint", func(ctx context.Context, id string) (map[string]interface{}, error) {
		ep, err := endpointSvc.GetEndpointByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return ep.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("BodyStructure", func(ctx context.Context, id string) (map[string]interface{}, error) {
		bs, err := bsSvc.GetBodyStructureByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return bs.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Substance", func(ctx context.Context, id string) (map[string]interface{}, error) {
		sub, err := substanceSvc.GetSubstanceByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return sub.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("Media", func(ctx context.Context, id string) (map[string]interface{}, error) {
		m, err := mediaSvc.GetMediaByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return m.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("DeviceRequest", func(ctx context.Context, id string) (map[string]interface{}, error) {
		dr, err := devReqSvc.GetDeviceRequestByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return dr.ToFHIR(), nil
	})
	includeRegistry.RegisterFetcher("DeviceUseStatement", func(ctx context.Context, id string) (map[string]interface{}, error) {
		dus, err := dusSvc.GetDeviceUseStatementByFHIRID(ctx, id)
		if err != nil {
			return nil, err
		}
		return dus.ToFHIR(), nil
	})

	// Reporting framework
	reportHandler := reporting.NewHandler(pool)
	reportHandler.RegisterRoutes(apiV1)

	// OpenAPI spec
	openAPIGen := openapi.NewGenerator(capBuilder, "0.1.0", baseURL)
	openAPIGen.RegisterRoutes(apiV1)

	// FHIR $export — register service adapters for real data export
	exportManager := fhir.NewExportManagerWithOptions(fhir.ExportOptions{
		MaxConcurrentJobs: 10,
		JobTTL:            time.Hour,
	})
	exportCleanupCtx, exportCleanupCancel := context.WithCancel(ctx)
	defer exportCleanupCancel()
	exportManager.StartCleanup(exportCleanupCtx)

	exportManager.RegisterExporter("Patient", &fhir.ServiceExporter{
		ResourceType: "Patient",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			patients, _, err := identitySvc.ListPatients(ctx, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(patients))
			for i, p := range patients {
				out[i] = p.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Encounter", &fhir.ServiceExporter{
		ResourceType: "Encounter",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := encSvc.ListEncounters(ctx, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, e := range items {
				out[i] = e.ToFHIR()
			}
			return out, nil
		},
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := encSvc.ListEncountersByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, e := range items {
				out[i] = e.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Condition", &fhir.ServiceExporter{
		ResourceType: "Condition",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := clinicalSvc.SearchConditions(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, c := range items {
				out[i] = c.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Observation", &fhir.ServiceExporter{
		ResourceType: "Observation",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := clinicalSvc.SearchObservations(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, o := range items {
				out[i] = o.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("MedicationRequest", &fhir.ServiceExporter{
		ResourceType: "MedicationRequest",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := medSvc.SearchMedicationRequests(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, mr := range items {
				out[i] = mr.ToFHIR()
			}
			return out, nil
		},
	})

	exportManager.RegisterExporter("AllergyIntolerance", &fhir.ServiceExporter{
		ResourceType: "AllergyIntolerance",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := clinicalSvc.ListAllergiesByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Procedure", &fhir.ServiceExporter{
		ResourceType: "Procedure",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := clinicalSvc.ListProceduresByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("DiagnosticReport", &fhir.ServiceExporter{
		ResourceType: "DiagnosticReport",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := dxSvc.ListDiagnosticReportsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Immunization", &fhir.ServiceExporter{
		ResourceType: "Immunization",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := immSvc.ListImmunizationsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("CarePlan", &fhir.ServiceExporter{
		ResourceType: "CarePlan",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := cpSvc.ListCarePlansByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Coverage", &fhir.ServiceExporter{
		ResourceType: "Coverage",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := billSvc.ListCoveragesByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("DocumentReference", &fhir.ServiceExporter{
		ResourceType: "DocumentReference",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := docSvc.ListDocumentReferencesByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("MedicationAdministration", &fhir.ServiceExporter{
		ResourceType: "MedicationAdministration",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := medSvc.ListMedicationAdministrationsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("ServiceRequest", &fhir.ServiceExporter{
		ResourceType: "ServiceRequest",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := dxSvc.ListServiceRequestsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})

	exportManager.RegisterExporter("Practitioner", &fhir.ServiceExporter{
		ResourceType: "Practitioner",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := identitySvc.ListPractitioners(ctx, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("MedicationDispense", &fhir.ServiceExporter{
		ResourceType: "MedicationDispense",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := medSvc.SearchMedicationDispenses(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := medSvc.ListMedicationDispensesByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("ImagingStudy", &fhir.ServiceExporter{
		ResourceType: "ImagingStudy",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := dxSvc.ListImagingStudiesByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Specimen", &fhir.ServiceExporter{
		ResourceType: "Specimen",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := dxSvc.ListSpecimensByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("ImmunizationRecommendation", &fhir.ServiceExporter{
		ResourceType: "ImmunizationRecommendation",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := immSvc.ListRecommendationsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Goal", &fhir.ServiceExporter{
		ResourceType: "Goal",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := cpSvc.ListGoalsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("CareTeam", &fhir.ServiceExporter{
		ResourceType: "CareTeam",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := ctSvc.SearchCareTeams(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := ctSvc.ListCareTeamsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Claim", &fhir.ServiceExporter{
		ResourceType: "Claim",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := billSvc.SearchClaims(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := billSvc.ListClaimsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Consent", &fhir.ServiceExporter{
		ResourceType: "Consent",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := docSvc.SearchConsents(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := docSvc.ListConsentsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Composition", &fhir.ServiceExporter{
		ResourceType: "Composition",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := docSvc.ListCompositionsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("FamilyMemberHistory", &fhir.ServiceExporter{
		ResourceType: "FamilyMemberHistory",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := fmhSvc.ListFamilyMemberHistoriesByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("RelatedPerson", &fhir.ServiceExporter{
		ResourceType: "RelatedPerson",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := rpSvc.ListRelatedPersonsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Appointment", &fhir.ServiceExporter{
		ResourceType: "Appointment",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := schedSvc.ListAppointmentsByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Device", &fhir.ServiceExporter{
		ResourceType: "Device",
		ListFn: func(ctx context.Context, since *time.Time) ([]map[string]interface{}, error) {
			items, _, err := devSvc.SearchDevices(ctx, nil, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})
	exportManager.RegisterExporter("Task", &fhir.ServiceExporter{
		ResourceType: "Task",
		ListByPatientFn: func(ctx context.Context, patientID string, since *time.Time) ([]map[string]interface{}, error) {
			pid, err := uuid.Parse(patientID)
			if err != nil {
				return nil, err
			}
			items, _, err := taskSvc.ListTasksByPatient(ctx, pid, 10000, 0)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]interface{}, len(items))
			for i, v := range items {
				out[i] = v.ToFHIR()
			}
			return out, nil
		},
	})

	// Group export resolver — returns 404 until a Group domain is implemented.
	exportManager.SetGroupResolver(func(ctx context.Context, groupID string) ([]string, error) {
		gid, err := uuid.Parse(groupID)
		if err != nil {
			return nil, fmt.Errorf("group not found: %s", groupID)
		}
		members, err := fhirGroupSvc.ListMembers(ctx, gid)
		if err != nil {
			return nil, fmt.Errorf("group not found: %s", groupID)
		}
		var patientIDs []string
		for _, m := range members {
			if !m.Inactive {
				patientIDs = append(patientIDs, m.EntityID)
			}
		}
		return patientIDs, nil
	})

	exportHandler := fhir.NewExportHandler(exportManager)
	exportHandler.RegisterRoutes(fhirGroup)

	// FHIR Binary resource
	binaryStore := fhir.NewInMemoryBinaryStore()
	binaryHandler := fhir.NewBinaryHandler(binaryStore)
	binaryHandler.RegisterRoutes(fhirGroup)

	// FHIR Patient/$everything — aggregates all patient compartment data
	everythingHandler := fhir.NewEverythingHandler()
	everythingHandler.SetPatientFetcher(func(ctx context.Context, fhirID string) (map[string]interface{}, error) {
		p, err := identitySvc.GetPatientByFHIRID(ctx, fhirID)
		if err != nil {
			return nil, err
		}
		return p.ToFHIR(), nil
	})
	everythingHandler.RegisterFetcher("Encounter", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := encSvc.ListEncountersByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Condition", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := clinicalSvc.ListConditionsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Observation", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := clinicalSvc.ListObservationsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("AllergyIntolerance", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := clinicalSvc.ListAllergiesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Procedure", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := clinicalSvc.ListProceduresByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("MedicationRequest", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := medSvc.ListMedicationRequestsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("MedicationAdministration", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := medSvc.ListMedicationAdministrationsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("MedicationDispense", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := medSvc.ListMedicationDispensesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("ServiceRequest", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := dxSvc.ListServiceRequestsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("DiagnosticReport", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := dxSvc.ListDiagnosticReportsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("ImagingStudy", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := dxSvc.ListImagingStudiesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Specimen", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := dxSvc.ListSpecimensByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Immunization", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := immSvc.ListImmunizationsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("ImmunizationRecommendation", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := immSvc.ListRecommendationsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("CarePlan", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := cpSvc.ListCarePlansByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Goal", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := cpSvc.ListGoalsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("CareTeam", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := ctSvc.ListCareTeamsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Coverage", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := billSvc.ListCoveragesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Claim", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := billSvc.ListClaimsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Consent", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := docSvc.ListConsentsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("DocumentReference", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := docSvc.ListDocumentReferencesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Composition", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := docSvc.ListCompositionsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("FamilyMemberHistory", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := fmhSvc.ListFamilyMemberHistoriesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("RelatedPerson", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := rpSvc.ListRelatedPersonsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Appointment", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := schedSvc.ListAppointmentsByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("QuestionnaireResponse", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := portalSvc.ListQuestionnaireResponsesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Device", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := devSvc.ListDevicesByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterFetcher("Task", func(ctx context.Context, patientID string) ([]map[string]interface{}, error) {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return nil, err
		}
		items, _, err := taskSvc.ListTasksByPatient(ctx, pid, 10000, 0)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]interface{}, len(items))
		for i, v := range items {
			out[i] = v.ToFHIR()
		}
		return out, nil
	})
	everythingHandler.RegisterRoutes(fhirGroup)

	// FHIR Patient compartment search — /fhir/Patient/:pid/:resourceType
	compartmentHandler := fhir.NewCompartmentHandler()
	compartmentHandler.RegisterSearchHandler("Encounter", encHandler.SearchEncountersFHIR)
	compartmentHandler.RegisterSearchHandler("Condition", clinicalHandler.SearchConditionsFHIR)
	compartmentHandler.RegisterSearchHandler("Observation", clinicalHandler.SearchObservationsFHIR)
	compartmentHandler.RegisterSearchHandler("AllergyIntolerance", clinicalHandler.SearchAllergiesFHIR)
	compartmentHandler.RegisterSearchHandler("Procedure", clinicalHandler.SearchProceduresFHIR)
	compartmentHandler.RegisterSearchHandler("MedicationRequest", medHandler.SearchMedicationRequestsFHIR)
	compartmentHandler.RegisterSearchHandler("MedicationAdministration", medHandler.SearchMedicationAdministrationsFHIR)
	compartmentHandler.RegisterSearchHandler("MedicationDispense", medHandler.SearchMedicationDispensesFHIR)
	compartmentHandler.RegisterSearchHandler("ServiceRequest", dxHandler.SearchServiceRequestsFHIR)
	compartmentHandler.RegisterSearchHandler("DiagnosticReport", dxHandler.SearchDiagnosticReportsFHIR)
	compartmentHandler.RegisterSearchHandler("Immunization", immHandler.SearchImmunizationsFHIR)
	compartmentHandler.RegisterSearchHandler("CarePlan", cpHandler.SearchCarePlansFHIR)
	compartmentHandler.RegisterSearchHandler("CareTeam", ctHandler.SearchCareTeamsFHIR)
	compartmentHandler.RegisterSearchHandler("Goal", cpHandler.SearchGoalsFHIR)
	compartmentHandler.RegisterSearchHandler("Coverage", billHandler.SearchCoveragesFHIR)
	compartmentHandler.RegisterSearchHandler("Claim", billHandler.SearchClaimsFHIR)
	compartmentHandler.RegisterSearchHandler("DocumentReference", docHandler.SearchDocumentReferencesFHIR)
	compartmentHandler.RegisterSearchHandler("Appointment", schedHandler.SearchAppointmentsFHIR)
	compartmentHandler.RegisterRoutes(fhirGroup)

	// CDS Hooks (HL7 CDS Hooks 2.0) — external clinical decision support
	cdsHooksHandler := fhir.NewCDSHooksHandler()

	// Service 1: patient-risk-alerts (hook: patient-view)
	cdsHooksHandler.RegisterService(fhir.CDSService{
		Hook:        "patient-view",
		Title:       "Patient Risk Alerts",
		Description: "Displays active CDS alerts when a patient chart is opened",
		ID:          "patient-risk-alerts",
		Prefetch: map[string]string{
			"patient": "Patient/{{context.patientId}}",
		},
	}, func(ctx context.Context, req fhir.CDSHookRequest) (*fhir.CDSHookResponse, error) {
		patientIDStr, _ := req.Context["patientId"].(string)
		if patientIDStr == "" {
			return &fhir.CDSHookResponse{Cards: []fhir.CDSCard{}}, nil
		}
		patientID, err := uuid.Parse(patientIDStr)
		if err != nil {
			return &fhir.CDSHookResponse{Cards: []fhir.CDSCard{}}, nil
		}
		alerts, _, err := cdsSvc.ListCDSAlertsByPatient(ctx, patientID, 100, 0)
		if err != nil {
			return nil, err
		}
		var cards []fhir.CDSCard
		for _, a := range alerts {
			if a.Status != "fired" {
				continue
			}
			indicator := "info"
			if a.Severity != nil {
				switch *a.Severity {
				case "critical", "high":
					indicator = "critical"
				case "moderate", "medium":
					indicator = "warning"
				}
			}
			card := fhir.CDSCard{
				UUID:      a.ID.String(),
				Summary:   a.Summary,
				Indicator: indicator,
				Source:    fhir.CDSSource{Label: "EHR CDS Engine"},
			}
			if a.Detail != nil {
				card.Detail = *a.Detail
			}
			if a.SuggestedAction != nil {
				card.Suggestions = []fhir.CDSSuggestion{
					{Label: *a.SuggestedAction},
				}
			}
			cards = append(cards, card)
		}
		if cards == nil {
			cards = []fhir.CDSCard{}
		}
		return &fhir.CDSHookResponse{Cards: cards}, nil
	})

	// Service 2: drug-interaction-check (hook: order-select)
	cdsHooksHandler.RegisterService(fhir.CDSService{
		Hook:        "order-select",
		Title:       "Drug Interaction Check",
		Description: "Checks for drug-drug interactions when a medication is selected",
		ID:          "drug-interaction-check",
		Prefetch: map[string]string{
			"patient": "Patient/{{context.patientId}}",
		},
	}, func(ctx context.Context, req fhir.CDSHookRequest) (*fhir.CDSHookResponse, error) {
		interactions, _, err := cdsSvc.ListDrugInteractions(ctx, 1000, 0)
		if err != nil {
			return nil, err
		}
		// Extract medication name from context.draftOrders
		var draftMedName string
		if draftOrders, ok := req.Context["draftOrders"].(map[string]interface{}); ok {
			if entries, ok := draftOrders["entry"].([]interface{}); ok {
				for _, entry := range entries {
					if e, ok := entry.(map[string]interface{}); ok {
						if res, ok := e["resource"].(map[string]interface{}); ok {
							if name, ok := res["medicationCodeableConcept"].(map[string]interface{}); ok {
								if text, ok := name["text"].(string); ok {
									draftMedName = text
								}
							}
						}
					}
				}
			}
		}
		var cards []fhir.CDSCard
		if draftMedName != "" {
			for _, ix := range interactions {
				if !ix.Active {
					continue
				}
				if ix.MedicationAName == draftMedName || ix.MedicationBName == draftMedName {
					indicator := "warning"
					if ix.Severity == "critical" || ix.Severity == "high" {
						indicator = "critical"
					}
					otherMed := ix.MedicationBName
					if ix.MedicationBName == draftMedName {
						otherMed = ix.MedicationAName
					}
					card := fhir.CDSCard{
						UUID:      ix.ID.String(),
						Summary:   fmt.Sprintf("Potential interaction: %s + %s", draftMedName, otherMed),
						Indicator: indicator,
						Source:    fhir.CDSSource{Label: "EHR CDS Engine"},
					}
					if ix.Description != nil {
						card.Detail = *ix.Description
					}
					if ix.Management != nil {
						card.Suggestions = []fhir.CDSSuggestion{
							{Label: *ix.Management},
						}
					}
					cards = append(cards, card)
				}
			}
		}
		if cards == nil {
			cards = []fhir.CDSCard{}
		}
		return &fhir.CDSHookResponse{Cards: cards}, nil
	})

	// Service 3: formulary-check (hook: order-select)
	cdsHooksHandler.RegisterService(fhir.CDSService{
		Hook:        "order-select",
		Title:       "Formulary Check",
		Description: "Checks formulary status when a medication is selected",
		ID:          "formulary-check",
		Prefetch: map[string]string{
			"patient": "Patient/{{context.patientId}}",
		},
	}, func(ctx context.Context, req fhir.CDSHookRequest) (*fhir.CDSHookResponse, error) {
		formularies, _, err := cdsSvc.ListFormularies(ctx, 100, 0)
		if err != nil {
			return nil, err
		}
		// Extract medication name from context.draftOrders
		var draftMedName string
		if draftOrders, ok := req.Context["draftOrders"].(map[string]interface{}); ok {
			if entries, ok := draftOrders["entry"].([]interface{}); ok {
				for _, entry := range entries {
					if e, ok := entry.(map[string]interface{}); ok {
						if res, ok := e["resource"].(map[string]interface{}); ok {
							if name, ok := res["medicationCodeableConcept"].(map[string]interface{}); ok {
								if text, ok := name["text"].(string); ok {
									draftMedName = text
								}
							}
						}
					}
				}
			}
		}
		var cards []fhir.CDSCard
		if draftMedName != "" {
			for _, f := range formularies {
				if !f.Active {
					continue
				}
				items, err := cdsSvc.GetFormularyItems(ctx, f.ID)
				if err != nil {
					continue
				}
				for _, item := range items {
					if item.MedicationName == draftMedName {
						if item.RequiresPriorAuth {
							cards = append(cards, fhir.CDSCard{
								Summary:   fmt.Sprintf("%s requires prior authorization on %s", draftMedName, f.Name),
								Indicator: "warning",
								Source:    fhir.CDSSource{Label: "EHR CDS Engine"},
							})
						}
						if item.PreferredStatus != nil && *item.PreferredStatus == "non-preferred" {
							cards = append(cards, fhir.CDSCard{
								Summary:   fmt.Sprintf("%s is non-preferred on %s", draftMedName, f.Name),
								Indicator: "info",
								Source:    fhir.CDSSource{Label: "EHR CDS Engine"},
								Detail:    "Consider a preferred alternative.",
							})
						}
					}
				}
			}
		}
		if cards == nil {
			cards = []fhir.CDSCard{}
		}
		return &fhir.CDSHookResponse{Cards: cards}, nil
	})

	// Shared feedback handler for all CDS Hooks services
	cdsFeedbackHandler := func(ctx context.Context, serviceID string, fb fhir.CDSFeedbackRequest) error {
		action := fb.Outcome
		if action == "" {
			action = "acknowledged"
		}
		var reason *string
		if len(fb.OverrideReasons) > 0 {
			r := fb.OverrideReasons[0].Display
			if r == "" {
				r = fb.OverrideReasons[0].Code
			}
			reason = &r
		}
		comment := fmt.Sprintf("CDS Hooks feedback for service %s, card %s", serviceID, fb.Card)
		resp := &cds.CDSAlertResponse{
			ID:             uuid.New(),
			AlertID:        uuid.New(),
			PractitionerID: uuid.New(),
			Action:         action,
			Reason:         reason,
			Comment:        &comment,
		}
		// Best-effort: if the card UUID is a valid alert ID, use it
		if alertID, err := uuid.Parse(fb.Card); err == nil {
			resp.AlertID = alertID
		}
		_ = cdsSvc.AddAlertResponse(ctx, resp)
		return nil
	}
	cdsHooksHandler.RegisterFeedbackHandler("patient-risk-alerts", cdsFeedbackHandler)
	cdsHooksHandler.RegisterFeedbackHandler("drug-interaction-check", cdsFeedbackHandler)
	cdsHooksHandler.RegisterFeedbackHandler("formulary-check", cdsFeedbackHandler)

	cdsHooksHandler.RegisterRoutes(e)

	// FHIR $validate — resource validation against structure and business rules
	resourceValidator := fhir.NewResourceValidator()
	validateHandler := fhir.NewValidateHandler(resourceValidator)
	validateHandler.RegisterRoutes(fhirGroup)

	// US Core Profile Validation (USCDI v3) — StructureDefinition-based profile validation
	profileRegistry := fhir.NewProfileRegistry()
	fhir.RegisterUSCoreProfiles(profileRegistry)
	profileValidator := fhir.NewProfileValidator(profileRegistry)
	profileHandler := fhir.NewProfileHandler(profileValidator, profileRegistry)
	profileHandler.RegisterRoutes(fhirGroup)

	// CQL Engine & FHIR Measure/$evaluate-measure — clinical quality measures
	measureEvaluator := fhir.NewMeasureEvaluator()
	measureHandler := fhir.NewMeasureHandler(measureEvaluator)
	measureHandler.RegisterRoutes(fhirGroup)

	// FHIR Patient/$merge — Master Data Management (MDM) with survivorship rules
	mdmService := fhir.NewMDMService()
	mergeHandler := fhir.NewMergeHandler(mdmService)
	mergeHandler.RegisterRoutes(fhirGroup)

	// FHIR Narrative Generation — auto-generate text.div XHTML for resources
	narrativeGenerator := fhir.NewNarrativeGenerator()
	fhirGroup.Use(fhir.NarrativeMiddleware(narrativeGenerator))

	// Server-Side Scripting (Bots) — FHIRPath-based automation engine
	botEngine := bot.NewBotEngine()
	bot.RegisterExampleBots(botEngine)
	botHandler := bot.NewBotHandler(botEngine)
	botHandler.RegisterRoutes(apiV1)

	// C-CDA Generation & Parsing — Continuity of Care Documents
	ccdaGenerator := ccda.NewGenerator("EHR System", "2.16.840.1.113883.3.0000")
	ccdaParser := ccda.NewParser()
	ccdaFetcher := &ccdaDataFetcher{
		identitySvc: identitySvc,
		clinicalSvc: clinicalSvc,
		medSvc:      medSvc,
		dxSvc:       dxSvc,
		immSvc:      immSvc,
		encSvc:      encSvc,
		cpSvc:       cpSvc,
	}
	ccdaHandler := ccda.NewHandler(ccdaGenerator, ccdaParser, ccdaFetcher)
	ccdaHandler.RegisterRoutes(apiV1)

	// SMART on FHIR App Launch — OAuth2 authorization server for SMART apps
	smartSigningKey, randomKey, err := resolveSmartSigningKey(os.Getenv("SMART_SIGNING_KEY"))
	if err != nil {
		logger.Fatal().Err(err).Msg("SMART signing key error")
	}
	if randomKey {
		logger.Warn().Msg("SMART_SIGNING_KEY not set; using random key (tokens will not survive restart)")
	}
	smartIssuer := "http://localhost:" + cfg.Port
	if issuer := os.Getenv("SMART_ISSUER"); issuer != "" {
		smartIssuer = issuer
	}
	smartServer := authpkg.NewSMARTServer(smartIssuer, smartSigningKey)
	smartCleanupCtx, smartCleanupCancel := context.WithCancel(context.Background())
	defer smartCleanupCancel()
	smartServer.StartCleanup(smartCleanupCtx)
	smartHandler := authpkg.NewSMARTHandler(smartServer)
	smartHandler.RegisterRoutes(e)

	// HL7v2 Interface Engine — parse and generate ADT, ORM, ORU messages
	hl7v2Handler := hl7v2.NewHandler()
	hl7v2Handler.RegisterRoutes(apiV1)

	// FHIR Patient/$match — probabilistic patient matching
	patientMatcher := fhir.NewPatientMatcher(&patientMatchSearcher{identitySvc: identitySvc})
	matchHandler := fhir.NewMatchHandler(patientMatcher)
	matchHandler.RegisterRoutes(fhirGroup)

	// FHIR ConceptMap/$translate — code system translation
	conceptMapTranslator := fhir.NewConceptMapTranslator()
	translateHandler := fhir.NewTranslateHandler(conceptMapTranslator)
	translateHandler.RegisterRoutes(fhirGroup)

	// FHIR CodeSystem/$subsumes — hierarchical code subsumption testing
	subsumptionChecker := fhir.NewSubsumptionChecker()
	subsumesHandler := fhir.NewSubsumesHandler(subsumptionChecker)
	subsumesHandler.RegisterRoutes(fhirGroup)

	// FHIR ValueSet/$validate-code — check code membership in value sets
	valueSetValidator := fhir.NewValueSetValidator()
	valueSetValidateHandler := fhir.NewValueSetValidateHandler(valueSetValidator)
	valueSetValidateHandler.RegisterRoutes(fhirGroup)

	// FHIR terminology service — $expand and $lookup operations
	terminologySvc := fhir.NewInMemoryTerminologyService()
	expandHandler := fhir.NewExpandHandler(terminologySvc)
	expandHandler.RegisterRoutes(fhirGroup)
	lookupHandler := fhir.NewLookupHandler(terminologySvc)
	lookupHandler.RegisterRoutes(fhirGroup)

	// FHIR Composition/$document — generate Document Bundles from Compositions
	documentResolver := &fhirResourceResolver{fhirGroup: fhirGroup}
	documentGenerator := fhir.NewDocumentGenerator(documentResolver)
	documentHandler := fhir.NewDocumentHandler(documentGenerator)
	documentHandler.RegisterRoutes(fhirGroup)

	// FHIR $process-message — process FHIR Message Bundles
	messageProcessor := fhir.NewMessageProcessor()
	messageHandler := fhir.NewMessageHandler(messageProcessor)
	messageHandler.RegisterRoutes(fhirGroup)

	// HL7v2 MLLP TCP listener (optional, started when MLLP_ADDR is set)
	if mllpAddr := os.Getenv("MLLP_ADDR"); mllpAddr != "" {
		mllpServer := hl7v2.NewMLLPServer(mllpAddr, hl7v2.DefaultHandler())
		go func() {
			if err := mllpServer.Start(); err != nil {
				logger.Error().Err(err).Msg("MLLP server failed")
			}
		}()
		defer mllpServer.Stop()
		logger.Info().Str("addr", mllpAddr).Msg("MLLP server started")
	}

	// Patient self-scheduling API
	selfSchedMgr := selfsched.NewSelfScheduleManager()
	selfSchedHandler := selfsched.NewSelfScheduleHandler(selfSchedMgr, selfSchedMgr)
	selfSchedHandler.RegisterRoutes(apiV1)

	// WebSocket real-time updates
	wsHub := websocket.NewHub()
	wsHandler := websocket.NewWebSocketHandler(wsHub)
	wsHandler.RegisterRoutes(apiV1)

	// Email/SMS notification service
	notifTemplates := notification.NewTemplateEngine()
	notifMgr := notification.NewNotificationManager(nil, nil, notifTemplates)
	notifHandler := notification.NewNotificationHandler(notifMgr)
	notifHandler.RegisterRoutes(apiV1)

	// Document/Blob storage
	blobStore := blobstore.NewInMemoryBlobStore()
	blobHandler := blobstore.NewBlobHandler(blobStore)
	blobHandler.RegisterRoutes(apiV1)

	// HTTP Cache/ETag middleware (applied to FHIR read endpoints)
	cacheStore := middleware.NewInMemoryCacheStore()
	cacheConfig := middleware.DefaultCacheConfig()
	cacheConfig.CacheStore = cacheStore
	_ = cacheConfig // ETag middleware can be applied per-route group as needed

	// Audit trail search and export
	auditSearcher := hipaa.NewAuditSearcher()
	auditSearchHandler := hipaa.NewAuditSearchHandler(auditSearcher)
	auditSearchHandler.RegisterRoutes(apiV1)

	// FHIR Bulk Import/Edit operations
	bulkStore := fhir.NewInMemoryResourceStore()
	bulkMgr := fhir.NewBulkOperationManager(bulkStore)
	bulkOpsHandler := fhir.NewBulkOpsHandler(bulkMgr)
	bulkOpsHandler.RegisterRoutes(fhirGroup)

	// FHIR $graphql — GraphQL query interface for FHIR resources
	graphqlEngine := fhir.NewGraphQLEngine()
	graphqlHandler := fhir.NewGraphQLHandler(graphqlEngine)
	graphqlHandler.RegisterRoutes(fhirGroup)

	// FHIR CodeSystem/$closure — transitive closure tables
	closureMgr := fhir.NewClosureManager()
	closureHandler := fhir.NewClosureHandler(closureMgr)
	closureHandler.RegisterRoutes(fhirGroup)

	// API Key Management
	apiKeyStore := auth.NewInMemoryAPIKeyStore()
	apiKeyMgr := auth.NewAPIKeyManager(apiKeyStore)
	apiKeyHandler := auth.NewAPIKeyHandler(apiKeyMgr)
	apiKeyHandler.RegisterRoutes(apiV1)

	// Per-client rate limiting
	clientRateLimiter := middleware.NewClientRateLimiter()
	rateLimitAdminHandler := middleware.NewRateLimitHandler(clientRateLimiter)
	rateLimitAdminHandler.RegisterRoutes(apiV1)

	// SMART Backend Services (client_credentials with JWT assertion)
	backendSvcStore := auth.NewInMemoryBackendServiceStore()
	backendSigningKey, _, _ := resolveSmartSigningKey(os.Getenv("SMART_SIGNING_KEY"))
	backendSvcMgr := auth.NewBackendServiceManager(backendSvcStore, backendSigningKey, cfg.AuthIssuer, cfg.AuthIssuer+"/auth/token")
	auth.RegisterBackendServiceEndpoints(fhirGroup, backendSvcMgr)

	// Webhook Management API
	webhookStore := webhook.NewInMemoryWebhookStore()
	webhookMgr := webhook.NewWebhookManager(webhookStore)
	webhookHandler := webhook.NewWebhookHandler(webhookMgr)
	webhookHandler.RegisterRoutes(apiV1)

	// API Usage Analytics
	usageTracker := analytics.NewUsageTracker(100000)
	usageHandler := analytics.NewUsageHandler(usageTracker)
	usageHandler.RegisterRoutes(apiV1)

	// Sandbox / Synthetic Data Seeder
	sandboxHandler := sandbox.NewSeedHandler()
	sandboxHandler.RegisterRoutes(apiV1)

	// Detailed CapabilityStatement (extended metadata endpoints)
	capabilityDetailed := fhir.DefaultCapabilityBuilder()
	capabilityHandler := fhir.NewCapabilityHandler(capabilityDetailed)
	capabilityHandler.RegisterRoutes(fhirGroup)

	// Suppress unused variable warnings for optional components
	_ = wsHub
	_ = notifMgr
	_ = blobStore
	_ = cacheStore
	_ = apiKeyMgr
	_ = clientRateLimiter
	_ = usageTracker

	// -- Platform Feature Wiring --

	// Shared FHIRPath engine — used by PlanDefinition/$apply and SQL-on-FHIR
	fhirPathEngine := fhir.NewFHIRPathEngine()

	// Auto-Provenance Middleware — automatically creates Provenance resources on writes
	provenanceStore := fhir.NewProvenanceStore()
	fhirGroup.Use(fhir.AutoProvenanceMiddleware(provenanceStore))

	// OpenTelemetry Observability — tracing, metrics, Prometheus endpoint
	telemetryProvider := telemetry.NewTelemetryProvider(telemetry.TelemetryConfig{
		ServiceName:    "ehr-server",
		ServiceVersion: "0.1.0",
		Environment:    cfg.Env,
	})
	e.Use(telemetryProvider.TracingMiddleware())
	e.Use(telemetryProvider.MetricsMiddleware())
	e.GET("/metrics", telemetryProvider.PrometheusHandler())

	// Topic-Based Subscriptions (R5-style) — clinical event notification
	topicEngine := fhir.NewSubscriptionTopicEngine()
	topicEngine.RegisterBuiltInTopics()
	topicHandler := fhir.NewTopicHandler(topicEngine)
	topicHandler.RegisterRoutes(fhirGroup)

	// PlanDefinition/$apply — clinical protocol automation
	planHandler := fhir.NewPlanDefinitionHandler(fhirPathEngine)
	planHandler.RegisterRoutes(fhirGroup)

	// SQL-on-FHIR ViewDefinitions — tabular views over FHIR resources
	viewEngine := fhir.NewViewDefinitionEngine(fhirPathEngine)
	viewHandler := fhir.NewViewDefinitionHandler(viewEngine)
	viewHandler.LoadBuiltIns()
	viewHandler.RegisterRoutes(fhirGroup)

	// Suppress unused warnings for new platform features
	_ = provenanceStore
	_ = telemetryProvider
	_ = topicEngine
	_ = fhirPathEngine

	// DB health check endpoint
	e.GET("/health/db", db.HealthHandler(pool))

	// Graceful shutdown
	go func() {
		addr := ":" + cfg.Port
		logger.Info().Str("addr", addr).Msg("starting server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server shutdown failed")
	}
	logger.Info().Msg("server stopped")
	return nil
}

// ccdaDataFetcher implements ccda.DataFetcher by aggregating data from domain services.
type ccdaDataFetcher struct {
	identitySvc *identity.Service
	clinicalSvc *clinical.Service
	medSvc      *medication.Service
	dxSvc       *diagnostics.Service
	immSvc      *immunization.Service
	encSvc      *encounter.Service
	cpSvc       *careplan.Service
}

func (f *ccdaDataFetcher) FetchPatientData(ctx context.Context, patientID string) (*ccda.PatientData, error) {
	pid, err := uuid.Parse(patientID)
	if err != nil {
		// Try FHIR ID lookup
		p, lookupErr := f.identitySvc.GetPatientByFHIRID(ctx, patientID)
		if lookupErr != nil {
			return nil, fmt.Errorf("patient not found: %s", patientID)
		}
		pid = p.ID
	}

	data := &ccda.PatientData{}

	// Patient demographics
	patient, err := f.identitySvc.GetPatient(ctx, pid)
	if err != nil {
		return nil, err
	}
	data.Patient = patient.ToFHIR()

	// Allergies
	if allergies, _, err := f.clinicalSvc.ListAllergiesByPatient(ctx, pid, 1000, 0); err == nil {
		for _, a := range allergies {
			data.Allergies = append(data.Allergies, a.ToFHIR())
		}
	}

	// Conditions
	if conditions, _, err := f.clinicalSvc.ListConditionsByPatient(ctx, pid, 1000, 0); err == nil {
		for _, c := range conditions {
			data.Conditions = append(data.Conditions, c.ToFHIR())
		}
	}

	// Medications
	if meds, _, err := f.medSvc.ListMedicationRequestsByPatient(ctx, pid, 1000, 0); err == nil {
		for _, m := range meds {
			data.Medications = append(data.Medications, m.ToFHIR())
		}
	}

	// Procedures
	if procs, _, err := f.clinicalSvc.ListProceduresByPatient(ctx, pid, 1000, 0); err == nil {
		for _, p := range procs {
			data.Procedures = append(data.Procedures, p.ToFHIR())
		}
	}

	// Results (lab observations)
	if obs, _, err := f.clinicalSvc.ListObservationsByPatient(ctx, pid, 1000, 0); err == nil {
		for _, o := range obs {
			fhirObs := o.ToFHIR()
			switch classifyObservation(fhirObs) {
			case "social-history":
				data.SocialHistory = append(data.SocialHistory, fhirObs)
			case "vital-signs":
				data.VitalSigns = append(data.VitalSigns, fhirObs)
			default:
				data.Results = append(data.Results, fhirObs)
			}
		}
	}

	// Immunizations
	if imms, _, err := f.immSvc.ListImmunizationsByPatient(ctx, pid, 1000, 0); err == nil {
		for _, i := range imms {
			data.Immunizations = append(data.Immunizations, i.ToFHIR())
		}
	}

	// Encounters
	if encs, _, err := f.encSvc.ListEncountersByPatient(ctx, pid, 1000, 0); err == nil {
		for _, e := range encs {
			data.Encounters = append(data.Encounters, e.ToFHIR())
		}
	}

	// Care Plans
	if plans, _, err := f.cpSvc.ListCarePlansByPatient(ctx, pid, 1000, 0); err == nil {
		for _, p := range plans {
			data.CarePlans = append(data.CarePlans, p.ToFHIR())
		}
	}

	return data, nil
}

// classifyObservation returns the FHIR observation category code (e.g.
// "vital-signs", "social-history", "laboratory") or an empty string when
// the category cannot be determined.
func classifyObservation(fhirObs map[string]interface{}) string {
	cats, ok := fhirObs["category"].([]map[string]interface{})
	if !ok || len(cats) == 0 {
		return ""
	}
	codings, ok := cats[0]["coding"].([]map[string]interface{})
	if !ok || len(codings) == 0 {
		return ""
	}
	code, _ := codings[0]["code"].(string)
	return code
}

// resolveSmartSigningKey returns the SMART signing key from the environment
// variable SMART_SIGNING_KEY (hex-encoded) or generates a random 32-byte
// key.  The second return value is true when a random key was generated.
func resolveSmartSigningKey(envValue string) ([]byte, bool, error) {
	if envValue != "" {
		decoded, err := hex.DecodeString(envValue)
		if err != nil {
			return nil, false, fmt.Errorf("invalid SMART_SIGNING_KEY hex value: %w", err)
		}
		return decoded, false, nil
	}
	key := make([]byte, 32)
	if _, err := crypto_rand.Read(key); err != nil {
		return nil, false, fmt.Errorf("failed to generate random SMART signing key: %w", err)
	}
	return key, true, nil
}

// patientMatchSearcher adapts the identity service to the fhir.PatientSearcher interface.
type patientMatchSearcher struct {
	identitySvc *identity.Service
}

func (s *patientMatchSearcher) SearchByDemographics(ctx context.Context, params map[string]string, limit int) ([]fhir.PatientRecord, error) {
	patients, _, err := s.identitySvc.SearchPatients(ctx, params, limit, 0)
	if err != nil {
		return nil, err
	}
	records := make([]fhir.PatientRecord, 0, len(patients))
	for _, p := range patients {
		rec := fhir.PatientRecord{
			ID:           p.ID.String(),
			FHIRResource: p.ToFHIR(),
			FirstName:    p.FirstName,
			LastName:     p.LastName,
			Gender:       stringVal(p.Gender),
			MRN:          p.MRN,
			Email:        stringVal(p.Email),
		}
		if p.BirthDate != nil {
			rec.BirthDate = p.BirthDate.Format("2006-01-02")
		}
		if p.PhoneMobile != nil {
			rec.Phone = *p.PhoneMobile
		} else if p.PhoneHome != nil {
			rec.Phone = *p.PhoneHome
		}
		if p.AddressLine1 != nil {
			rec.AddressLine = *p.AddressLine1
		}
		if p.City != nil {
			rec.City = *p.City
		}
		if p.PostalCode != nil {
			rec.PostalCode = *p.PostalCode
		}
		records = append(records, rec)
	}
	return records, nil
}

func stringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// fhirResourceResolver implements fhir.ResourceResolver for the $document operation.
// It resolves FHIR references like "Patient/123" by delegating to domain services.
// For now it returns a minimal stub — full resolution would require a service registry.
type fhirResourceResolver struct {
	fhirGroup *echo.Group
}

func (r *fhirResourceResolver) ResolveReference(ctx context.Context, reference string) (map[string]interface{}, error) {
	// Return a minimal resource stub with the reference preserved.
	// Full implementation would parse "ResourceType/id" and query the appropriate service.
	parts := strings.SplitN(reference, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference format: %s", reference)
	}
	return map[string]interface{}{
		"resourceType": parts[0],
		"id":           parts[1],
	}, nil
}
