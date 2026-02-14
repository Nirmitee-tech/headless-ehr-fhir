package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/ehr/ehr/internal/config"
	"github.com/ehr/ehr/internal/domain/admin"
	"github.com/ehr/ehr/internal/domain/behavioral"
	"github.com/ehr/ehr/internal/domain/billing"
	"github.com/ehr/ehr/internal/domain/cds"
	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/ehr/ehr/internal/domain/diagnostics"
	"github.com/ehr/ehr/internal/domain/documents"
	"github.com/ehr/ehr/internal/domain/emergency"
	"github.com/ehr/ehr/internal/domain/encounter"
	"github.com/ehr/ehr/internal/domain/identity"
	"github.com/ehr/ehr/internal/domain/inbox"
	"github.com/ehr/ehr/internal/domain/medication"
	"github.com/ehr/ehr/internal/domain/nursing"
	"github.com/ehr/ehr/internal/domain/obstetrics"
	"github.com/ehr/ehr/internal/domain/oncology"
	"github.com/ehr/ehr/internal/domain/portal"
	"github.com/ehr/ehr/internal/domain/research"
	"github.com/ehr/ehr/internal/domain/scheduling"
	"github.com/ehr/ehr/internal/domain/surgery"
	"github.com/ehr/ehr/internal/domain/terminology"
	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/internal/platform/middleware"
)

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

	// FHIR metadata (dynamic CapabilityStatement)
	fhirGroup.GET("/metadata", func(c echo.Context) error {
		return c.JSON(http.StatusOK, capBuilder.Build())
	})

	// SMART on FHIR discovery
	auth.RegisterSMARTEndpoints(fhirGroup, cfg.AuthIssuer)

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
	adminHandler := admin.NewHandler(adminSvc)
	adminHandler.RegisterRoutes(apiV1, fhirGroup)

	// Identity domain
	patientRepo := identity.NewPatientRepo(pool)
	practRepo := identity.NewPractitionerRepo(pool)
	identitySvc := identity.NewService(patientRepo, practRepo)
	identityHandler := identity.NewHandler(identitySvc)
	identityHandler.RegisterRoutes(apiV1, fhirGroup)

	// Encounter domain
	encRepo := encounter.NewRepo(pool)
	encSvc := encounter.NewService(encRepo)
	encHandler := encounter.NewHandler(encSvc)
	encHandler.RegisterRoutes(apiV1, fhirGroup)

	// Clinical domain
	condRepo := clinical.NewConditionRepoPG(pool)
	obsRepo := clinical.NewObservationRepoPG(pool)
	allergyRepo := clinical.NewAllergyRepoPG(pool)
	procRepo := clinical.NewProcedureRepoPG(pool)
	clinicalSvc := clinical.NewService(condRepo, obsRepo, allergyRepo, procRepo)
	clinicalHandler := clinical.NewHandler(clinicalSvc)
	clinicalHandler.RegisterRoutes(apiV1, fhirGroup)

	// Medication domain
	medRepo := medication.NewMedicationRepoPG(pool)
	medReqRepo := medication.NewMedicationRequestRepoPG(pool)
	medAdminRepo := medication.NewMedicationAdministrationRepoPG(pool)
	medDispRepo := medication.NewMedicationDispenseRepoPG(pool)
	medStmtRepo := medication.NewMedicationStatementRepoPG(pool)
	medSvc := medication.NewService(medRepo, medReqRepo, medAdminRepo, medDispRepo, medStmtRepo)
	medHandler := medication.NewHandler(medSvc)
	medHandler.RegisterRoutes(apiV1, fhirGroup)

	// Diagnostics domain
	srRepo := diagnostics.NewServiceRequestRepoPG(pool)
	specRepo := diagnostics.NewSpecimenRepoPG(pool)
	dxReportRepo := diagnostics.NewDiagnosticReportRepoPG(pool)
	imgRepo := diagnostics.NewImagingStudyRepoPG(pool)
	dxSvc := diagnostics.NewService(srRepo, specRepo, dxReportRepo, imgRepo)
	dxHandler := diagnostics.NewHandler(dxSvc)
	dxHandler.RegisterRoutes(apiV1, fhirGroup)

	// Scheduling domain
	schedRepo := scheduling.NewScheduleRepoPG(pool)
	slotRepo := scheduling.NewSlotRepoPG(pool)
	apptRepo := scheduling.NewAppointmentRepoPG(pool)
	wlRepo := scheduling.NewWaitlistRepoPG(pool)
	schedSvc := scheduling.NewService(schedRepo, slotRepo, apptRepo, wlRepo)
	schedHandler := scheduling.NewHandler(schedSvc)
	schedHandler.RegisterRoutes(apiV1, fhirGroup)

	// Billing domain
	covRepo := billing.NewCoverageRepoPG(pool)
	claimRepo := billing.NewClaimRepoPG(pool)
	claimRespRepo := billing.NewClaimResponseRepoPG(pool)
	invRepo := billing.NewInvoiceRepoPG(pool)
	billSvc := billing.NewService(covRepo, claimRepo, claimRespRepo, invRepo)
	billHandler := billing.NewHandler(billSvc)
	billHandler.RegisterRoutes(apiV1, fhirGroup)

	// Documents domain
	consentRepo := documents.NewConsentRepoPG(pool)
	docRefRepo := documents.NewDocumentReferenceRepoPG(pool)
	noteRepo := documents.NewClinicalNoteRepoPG(pool)
	compRepo := documents.NewCompositionRepoPG(pool)
	docSvc := documents.NewService(consentRepo, docRefRepo, noteRepo, compRepo)
	docHandler := documents.NewHandler(docSvc)
	docHandler.RegisterRoutes(apiV1, fhirGroup)

	// Inbox domain
	poolRepo := inbox.NewMessagePoolRepoPG(pool)
	msgRepo := inbox.NewInboxMessageRepoPG(pool)
	cosignRepo := inbox.NewCosignRequestRepoPG(pool)
	listRepo := inbox.NewPatientListRepoPG(pool)
	handoffRepo := inbox.NewHandoffRepoPG(pool)
	inboxSvc := inbox.NewService(poolRepo, msgRepo, cosignRepo, listRepo, handoffRepo)
	inboxHandler := inbox.NewHandler(inboxSvc)
	inboxHandler.RegisterRoutes(apiV1, fhirGroup)

	// Surgery domain
	orRoomRepo := surgery.NewORRoomRepoPG(pool)
	caseRepo := surgery.NewSurgicalCaseRepoPG(pool)
	prefCardRepo := surgery.NewPreferenceCardRepoPG(pool)
	implantRepo := surgery.NewImplantLogRepoPG(pool)
	surgerySvc := surgery.NewService(orRoomRepo, caseRepo, prefCardRepo, implantRepo)
	surgeryHandler := surgery.NewHandler(surgerySvc)
	surgeryHandler.RegisterRoutes(apiV1, fhirGroup)

	// Emergency domain
	triageRepo := emergency.NewTriageRepoPG(pool)
	edTrackRepo := emergency.NewEDTrackingRepoPG(pool)
	traumaRepo := emergency.NewTraumaRepoPG(pool)
	edSvc := emergency.NewService(triageRepo, edTrackRepo, traumaRepo)
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
	nurseHandler := nursing.NewHandler(nurseSvc)
	nurseHandler.RegisterRoutes(apiV1)

	// Behavioral health domain
	psychRepo := behavioral.NewPsychAssessmentRepoPG(pool)
	safetyRepo := behavioral.NewSafetyPlanRepoPG(pool)
	legalRepo := behavioral.NewLegalHoldRepoPG(pool)
	seclusionRepo := behavioral.NewSeclusionRestraintRepoPG(pool)
	groupRepo := behavioral.NewGroupTherapyRepoPG(pool)
	bhSvc := behavioral.NewService(psychRepo, safetyRepo, legalRepo, seclusionRepo, groupRepo)
	bhHandler := behavioral.NewHandler(bhSvc)
	bhHandler.RegisterRoutes(apiV1, fhirGroup)

	// Research domain
	studyRepo := research.NewStudyRepoPG(pool)
	enrollRepo := research.NewEnrollmentRepoPG(pool)
	advEventRepo := research.NewAdverseEventRepoPG(pool)
	devRepo := research.NewDeviationRepoPG(pool)
	resSvc := research.NewService(studyRepo, enrollRepo, advEventRepo, devRepo)
	resHandler := research.NewHandler(resSvc)
	resHandler.RegisterRoutes(apiV1, fhirGroup)

	// Portal domain
	portalAcctRepo := portal.NewPortalAccountRepoPG(pool)
	portalMsgRepo := portal.NewPortalMessageRepoPG(pool)
	questRepo := portal.NewQuestionnaireRepoPG(pool)
	questRespRepo := portal.NewQuestionnaireResponseRepoPG(pool)
	checkinRepo := portal.NewPatientCheckinRepoPG(pool)
	portalSvc := portal.NewService(portalAcctRepo, portalMsgRepo, questRepo, questRespRepo, checkinRepo)
	portalHandler := portal.NewHandler(portalSvc)
	portalHandler.RegisterRoutes(apiV1, fhirGroup)

	// Terminology domain
	loincRepo := terminology.NewLOINCRepoPG(pool)
	icd10Repo := terminology.NewICD10RepoPG(pool)
	snomedRepo := terminology.NewSNOMEDRepoPG(pool)
	rxnormRepo := terminology.NewRxNormRepoPG(pool)
	cptRepo := terminology.NewCPTRepoPG(pool)
	termSvc := terminology.NewService(loincRepo, icd10Repo, snomedRepo, rxnormRepo, cptRepo)
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
	cdsHandler := cds.NewHandler(cdsSvc)
	cdsHandler.RegisterRoutes(apiV1)

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
