package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/ehr/ehr/internal/config"
	"github.com/ehr/ehr/internal/domain/admin"
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
	"github.com/ehr/ehr/internal/domain/encounter"
	"github.com/ehr/ehr/internal/domain/familyhistory"
	"github.com/ehr/ehr/internal/domain/identity"
	"github.com/ehr/ehr/internal/domain/immunization"
	"github.com/ehr/ehr/internal/domain/inbox"
	"github.com/ehr/ehr/internal/domain/medication"
	"github.com/ehr/ehr/internal/domain/nursing"
	"github.com/ehr/ehr/internal/domain/obstetrics"
	"github.com/ehr/ehr/internal/domain/oncology"
	"github.com/ehr/ehr/internal/domain/portal"
	"github.com/ehr/ehr/internal/domain/provenance"
	"github.com/ehr/ehr/internal/domain/relatedperson"
	"github.com/ehr/ehr/internal/domain/research"
	"github.com/ehr/ehr/internal/domain/scheduling"
	"github.com/ehr/ehr/internal/domain/surgery"
	"github.com/ehr/ehr/internal/domain/subscription"
	fhirtask "github.com/ehr/ehr/internal/domain/task"
	"github.com/ehr/ehr/internal/domain/terminology"
	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/internal/platform/hipaa"
	"github.com/ehr/ehr/internal/platform/middleware"
	"github.com/ehr/ehr/internal/platform/openapi"
	"github.com/ehr/ehr/internal/platform/reporting"
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

	// Billing domain
	capBuilder.AddResource("Coverage", fhir.DefaultInteractions(), []fhir.SearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})
	capBuilder.AddResource("Claim", fhir.DefaultInteractions(), []fhir.SearchParam{
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

	// Set advanced capabilities for all registered resource types
	defaultCaps := fhir.DefaultCapabilityOptions()
	for _, rt := range []string{
		"Patient", "Practitioner", "Organization", "Location", "Encounter",
		"Condition", "Observation", "AllergyIntolerance", "Procedure",
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
	includeRegistry.RegisterReference("Provenance", "target", "Patient")
	includeRegistry.RegisterReference("Provenance", "agent", "Practitioner")

	// Wire _include/_revinclude middleware into the FHIR search group.
	// Fetchers are registered below after services are initialized; since
	// the middleware holds a pointer to the registry, late registration works.
	fhirGroup.Use(fhir.IncludeMiddleware(includeRegistry))

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
	identitySvc := identity.NewService(patientRepo, practRepo, patientLinkRepo)
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
	wlRepo := scheduling.NewWaitlistRepoPG(pool)
	schedSvc := scheduling.NewService(schedRepo, slotRepo, apptRepo, wlRepo)
	schedSvc.SetVersionTracker(versionTracker)
	schedHandler := scheduling.NewHandler(schedSvc)
	schedHandler.RegisterRoutes(apiV1, fhirGroup)

	// Billing domain
	covRepo := billing.NewCoverageRepoPG(pool)
	claimRepo := billing.NewClaimRepoPG(pool)
	claimRespRepo := billing.NewClaimResponseRepoPG(pool)
	invRepo := billing.NewInvoiceRepoPG(pool)
	billSvc := billing.NewService(covRepo, claimRepo, claimRespRepo, invRepo)
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

	// Reporting framework
	reportHandler := reporting.NewHandler(pool)
	reportHandler.RegisterRoutes(apiV1)

	// OpenAPI spec
	openAPIGen := openapi.NewGenerator(capBuilder, "0.1.0", baseURL)
	openAPIGen.RegisterRoutes(apiV1)

	// FHIR $export — register service adapters for real data export
	exportManager := fhir.NewExportManager()

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

	exportHandler := fhir.NewExportHandler(exportManager)
	exportHandler.RegisterRoutes(fhirGroup)

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
