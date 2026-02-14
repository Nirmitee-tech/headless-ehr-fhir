package integration

import (
	"context"
	"testing"

	"github.com/ehr/ehr/internal/domain/terminology"
)

// seedLOINCData inserts test LOINC codes into the reference_loinc table.
func seedLOINCData(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()

	// Create the reference table if it doesn't exist
	createSQL := `CREATE TABLE IF NOT EXISTS reference_loinc (
		code VARCHAR(20) PRIMARY KEY,
		display VARCHAR(255) NOT NULL,
		component VARCHAR(100),
		property VARCHAR(50),
		time_aspect VARCHAR(20),
		system_uri VARCHAR(255) DEFAULT 'http://loinc.org',
		category VARCHAR(50)
	)`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, createSQL); err != nil {
		t.Fatalf("create reference_loinc table: %v", err)
	}

	// Insert test data
	insertSQL := `INSERT INTO reference_loinc (code, display, component, property, time_aspect, category) VALUES
		('8310-5',  'Body temperature',     'Body temperature',  'Temp',  'Pt', 'vital-signs'),
		('8867-4',  'Heart rate',           'Heart rate',        'NRat',  'Pt', 'vital-signs'),
		('9279-1',  'Respiratory rate',     'Respiratory rate',  'NRat',  'Pt', 'vital-signs'),
		('2160-0',  'Creatinine [Mass/volume] in Serum or Plasma', 'Creatinine', 'MCnc', 'Pt', 'laboratory'),
		('2345-7',  'Glucose [Mass/volume] in Serum or Plasma',    'Glucose',    'MCnc', 'Pt', 'laboratory')
		ON CONFLICT (code) DO NOTHING`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, insertSQL); err != nil {
		t.Fatalf("seed reference_loinc data: %v", err)
	}
}

// seedICD10Data inserts test ICD-10 codes into the reference_icd10 table.
func seedICD10Data(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()

	createSQL := `CREATE TABLE IF NOT EXISTS reference_icd10 (
		code VARCHAR(10) PRIMARY KEY,
		display VARCHAR(500) NOT NULL,
		category VARCHAR(100),
		chapter VARCHAR(10),
		system_uri VARCHAR(255) DEFAULT 'http://hl7.org/fhir/sid/icd-10-cm'
	)`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, createSQL); err != nil {
		t.Fatalf("create reference_icd10 table: %v", err)
	}

	insertSQL := `INSERT INTO reference_icd10 (code, display, category, chapter) VALUES
		('I10',    'Essential (primary) hypertension',                          'Circulatory',  'IX'),
		('E11.9',  'Type 2 diabetes mellitus without complications',            'Endocrine',    'IV'),
		('J06.9',  'Acute upper respiratory infection, unspecified',             'Respiratory',  'X'),
		('M54.5',  'Low back pain',                                             'Musculoskeletal', 'XIII'),
		('K21.0',  'Gastro-esophageal reflux disease with esophagitis',         'Digestive',    'XI')
		ON CONFLICT (code) DO NOTHING`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, insertSQL); err != nil {
		t.Fatalf("seed reference_icd10 data: %v", err)
	}
}

// seedSNOMEDData inserts test SNOMED codes into the reference_snomed table.
func seedSNOMEDData(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()

	createSQL := `CREATE TABLE IF NOT EXISTS reference_snomed (
		code VARCHAR(20) PRIMARY KEY,
		display VARCHAR(500) NOT NULL,
		semantic_tag VARCHAR(50),
		category VARCHAR(50),
		system_uri VARCHAR(255) DEFAULT 'http://snomed.info/sct'
	)`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, createSQL); err != nil {
		t.Fatalf("create reference_snomed table: %v", err)
	}

	insertSQL := `INSERT INTO reference_snomed (code, display, semantic_tag, category) VALUES
		('80146002',  'Appendectomy',                      'procedure', 'surgical'),
		('73761001',  'Colonoscopy',                       'procedure', 'diagnostic'),
		('38341003',  'Hypertensive disorder',             'disorder',  'cardiovascular'),
		('44054006',  'Type 2 diabetes mellitus',          'disorder',  'endocrine'),
		('195967001', 'Asthma',                            'disorder',  'respiratory')
		ON CONFLICT (code) DO NOTHING`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, insertSQL); err != nil {
		t.Fatalf("seed reference_snomed data: %v", err)
	}
}

// seedRxNormData inserts test RxNorm codes into the reference_medication table.
func seedRxNormData(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()

	createSQL := `CREATE TABLE IF NOT EXISTS reference_medication (
		rxnorm_code VARCHAR(20) PRIMARY KEY,
		display VARCHAR(500) NOT NULL,
		generic_name VARCHAR(255),
		drug_class VARCHAR(100),
		route VARCHAR(50),
		form VARCHAR(50),
		system_uri VARCHAR(255) DEFAULT 'http://www.nlm.nih.gov/research/umls/rxnorm'
	)`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, createSQL); err != nil {
		t.Fatalf("create reference_medication table: %v", err)
	}

	insertSQL := `INSERT INTO reference_medication (rxnorm_code, display, generic_name, drug_class, route, form) VALUES
		('860975',  'Metformin 500 mg oral tablet',          'Metformin',       'Biguanide',           'oral',   'tablet'),
		('197361', 'Amlodipine 5 mg oral tablet',            'Amlodipine',      'Calcium Channel Blocker', 'oral', 'tablet'),
		('314076', 'Lisinopril 10 mg oral tablet',           'Lisinopril',      'ACE Inhibitor',       'oral',   'tablet'),
		('198440', 'Atorvastatin 20 mg oral tablet',         'Atorvastatin',    'Statin',              'oral',   'tablet'),
		('310965', 'Omeprazole 20 mg oral capsule',          'Omeprazole',      'Proton Pump Inhibitor','oral',  'capsule')
		ON CONFLICT (rxnorm_code) DO NOTHING`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, insertSQL); err != nil {
		t.Fatalf("seed reference_medication data: %v", err)
	}
}

// seedCPTData inserts test CPT codes into the reference_cpt table.
func seedCPTData(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()

	createSQL := `CREATE TABLE IF NOT EXISTS reference_cpt (
		code VARCHAR(10) PRIMARY KEY,
		display VARCHAR(500) NOT NULL,
		category VARCHAR(100),
		subcategory VARCHAR(100),
		system_uri VARCHAR(255) DEFAULT 'http://www.ama-assn.org/go/cpt'
	)`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, createSQL); err != nil {
		t.Fatalf("create reference_cpt table: %v", err)
	}

	insertSQL := `INSERT INTO reference_cpt (code, display, category, subcategory) VALUES
		('99201', 'Office visit, new patient, minimal',                    'E&M', 'Office Visit - New'),
		('99213', 'Office visit, established patient, low complexity',     'E&M', 'Office Visit - Established'),
		('99285', 'Emergency department visit, high complexity',           'E&M', 'Emergency'),
		('36415', 'Collection of venous blood by venipuncture',            'Laboratory', 'Phlebotomy'),
		('71046', 'Radiologic exam, chest, 2 views',                       'Radiology', 'Chest X-Ray')
		ON CONFLICT (code) DO NOTHING`
	if err := execWithSchema(ctx, globalDB.Pool, tenantID, insertSQL); err != nil {
		t.Fatalf("seed reference_cpt data: %v", err)
	}
}

// ==================== LOINC Tests ====================

func TestLOINCSearchAndGetByCode(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("loinc")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	seedLOINCData(t, ctx, tenantID)

	t.Run("Search_ByDisplay", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "temperature", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'temperature' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "8310-5" {
					found = true
					if r.Display != "Body temperature" {
						t.Fatalf("expected display 'Body temperature', got %q", r.Display)
					}
					if r.Component != "Body temperature" {
						t.Fatalf("expected component 'Body temperature', got %q", r.Component)
					}
					if r.Property != "Temp" {
						t.Fatalf("expected property 'Temp', got %q", r.Property)
					}
					if r.TimeAspect != "Pt" {
						t.Fatalf("expected time_aspect 'Pt', got %q", r.TimeAspect)
					}
					if r.SystemURI != "http://loinc.org" {
						t.Fatalf("expected system_uri 'http://loinc.org', got %q", r.SystemURI)
					}
					if r.Category != "vital-signs" {
						t.Fatalf("expected category 'vital-signs', got %q", r.Category)
					}
				}
			}
			if !found {
				t.Fatal("expected to find LOINC code 8310-5 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search LOINC by display: %v", err)
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "8867", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for '8867' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "8867-4" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find LOINC code 8867-4 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search LOINC by code: %v", err)
		}
	})

	t.Run("Search_ByComponent", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Creatinine", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Creatinine' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "2160-0" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find LOINC code 2160-0 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search LOINC by component: %v", err)
		}
	})

	t.Run("Search_WithLimit", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "rate", 1)
			if err != nil {
				return err
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result with limit=1, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search LOINC with limit: %v", err)
		}
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "zzz_nonexistent_zzz", 10)
			if err != nil {
				return err
			}
			if len(results) != 0 {
				t.Fatalf("expected 0 results for nonexistent search, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search LOINC no results: %v", err)
		}
	})

	t.Run("Search_DefaultLimit", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			// Passing 0 should default to 20; search for "rate" which matches Heart rate, Respiratory rate
			results, err := repo.Search(ctx, "rate", 0)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least 1 result for 'rate' with default limit")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search LOINC default limit: %v", err)
		}
	})

	t.Run("GetByCode_Found", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			code, err := repo.GetByCode(ctx, "8310-5")
			if err != nil {
				return err
			}
			if code.Code != "8310-5" {
				t.Fatalf("expected code '8310-5', got %q", code.Code)
			}
			if code.Display != "Body temperature" {
				t.Fatalf("expected display 'Body temperature', got %q", code.Display)
			}
			if code.Component != "Body temperature" {
				t.Fatalf("expected component 'Body temperature', got %q", code.Component)
			}
			if code.SystemURI != "http://loinc.org" {
				t.Fatalf("expected system_uri 'http://loinc.org', got %q", code.SystemURI)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode LOINC found: %v", err)
		}
	})

	t.Run("GetByCode_NotFound", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewLOINCRepoPG(globalDB.Pool)
			_, err := repo.GetByCode(ctx, "99999-9")
			if err == nil {
				t.Fatal("expected error for nonexistent code, got nil")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode LOINC not found: %v", err)
		}
	})
}

// ==================== ICD-10 Tests ====================

func TestICD10SearchAndGetByCode(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("icd10")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	seedICD10Data(t, ctx, tenantID)

	t.Run("Search_ByDisplay", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "hypertension", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'hypertension' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "I10" {
					found = true
					if r.Category != "Circulatory" {
						t.Fatalf("expected category 'Circulatory', got %q", r.Category)
					}
					if r.Chapter != "IX" {
						t.Fatalf("expected chapter 'IX', got %q", r.Chapter)
					}
					if r.SystemURI != "http://hl7.org/fhir/sid/icd-10-cm" {
						t.Fatalf("expected system_uri 'http://hl7.org/fhir/sid/icd-10-cm', got %q", r.SystemURI)
					}
				}
			}
			if !found {
				t.Fatal("expected to find ICD-10 code I10 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search ICD-10 by display: %v", err)
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "E11", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'E11' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "E11.9" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find ICD-10 code E11.9 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search ICD-10 by code: %v", err)
		}
	})

	t.Run("Search_ByCategory", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Respiratory", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Respiratory' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "J06.9" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find ICD-10 code J06.9 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search ICD-10 by category: %v", err)
		}
	})

	t.Run("Search_WithLimit", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			// Search broadly to get multiple matches, then limit
			results, err := repo.Search(ctx, "e", 2)
			if err != nil {
				return err
			}
			if len(results) > 2 {
				t.Fatalf("expected at most 2 results with limit=2, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search ICD-10 with limit: %v", err)
		}
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "zzz_nonexistent_zzz", 10)
			if err != nil {
				return err
			}
			if len(results) != 0 {
				t.Fatalf("expected 0 results for nonexistent search, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search ICD-10 no results: %v", err)
		}
	})

	t.Run("GetByCode_Found", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			code, err := repo.GetByCode(ctx, "I10")
			if err != nil {
				return err
			}
			if code.Code != "I10" {
				t.Fatalf("expected code 'I10', got %q", code.Code)
			}
			if code.Display != "Essential (primary) hypertension" {
				t.Fatalf("expected display 'Essential (primary) hypertension', got %q", code.Display)
			}
			if code.Category != "Circulatory" {
				t.Fatalf("expected category 'Circulatory', got %q", code.Category)
			}
			if code.Chapter != "IX" {
				t.Fatalf("expected chapter 'IX', got %q", code.Chapter)
			}
			if code.SystemURI != "http://hl7.org/fhir/sid/icd-10-cm" {
				t.Fatalf("expected system_uri 'http://hl7.org/fhir/sid/icd-10-cm', got %q", code.SystemURI)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode ICD-10 found: %v", err)
		}
	})

	t.Run("GetByCode_NotFound", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewICD10RepoPG(globalDB.Pool)
			_, err := repo.GetByCode(ctx, "ZZZ.99")
			if err == nil {
				t.Fatal("expected error for nonexistent code, got nil")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode ICD-10 not found: %v", err)
		}
	})
}

// ==================== SNOMED Tests ====================

func TestSNOMEDSearchAndGetByCode(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("snomed")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	seedSNOMEDData(t, ctx, tenantID)

	t.Run("Search_ByDisplay", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Appendectomy", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Appendectomy' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "80146002" {
					found = true
					if r.Display != "Appendectomy" {
						t.Fatalf("expected display 'Appendectomy', got %q", r.Display)
					}
					if r.SemanticTag != "procedure" {
						t.Fatalf("expected semantic_tag 'procedure', got %q", r.SemanticTag)
					}
					if r.Category != "surgical" {
						t.Fatalf("expected category 'surgical', got %q", r.Category)
					}
					if r.SystemURI != "http://snomed.info/sct" {
						t.Fatalf("expected system_uri 'http://snomed.info/sct', got %q", r.SystemURI)
					}
				}
			}
			if !found {
				t.Fatal("expected to find SNOMED code 80146002 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search SNOMED by display: %v", err)
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "73761001", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for '73761001' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "73761001" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find SNOMED code 73761001 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search SNOMED by code: %v", err)
		}
	})

	t.Run("Search_ByCategory", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "cardiovascular", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'cardiovascular' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "38341003" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find SNOMED code 38341003 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search SNOMED by category: %v", err)
		}
	})

	t.Run("Search_WithLimit", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "disorder", 1)
			if err != nil {
				return err
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result with limit=1, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search SNOMED with limit: %v", err)
		}
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "zzz_nonexistent_zzz", 10)
			if err != nil {
				return err
			}
			if len(results) != 0 {
				t.Fatalf("expected 0 results for nonexistent search, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search SNOMED no results: %v", err)
		}
	})

	t.Run("GetByCode_Found", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			code, err := repo.GetByCode(ctx, "38341003")
			if err != nil {
				return err
			}
			if code.Code != "38341003" {
				t.Fatalf("expected code '38341003', got %q", code.Code)
			}
			if code.Display != "Hypertensive disorder" {
				t.Fatalf("expected display 'Hypertensive disorder', got %q", code.Display)
			}
			if code.SemanticTag != "disorder" {
				t.Fatalf("expected semantic_tag 'disorder', got %q", code.SemanticTag)
			}
			if code.Category != "cardiovascular" {
				t.Fatalf("expected category 'cardiovascular', got %q", code.Category)
			}
			if code.SystemURI != "http://snomed.info/sct" {
				t.Fatalf("expected system_uri 'http://snomed.info/sct', got %q", code.SystemURI)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode SNOMED found: %v", err)
		}
	})

	t.Run("GetByCode_NotFound", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewSNOMEDRepoPG(globalDB.Pool)
			_, err := repo.GetByCode(ctx, "999999999")
			if err == nil {
				t.Fatal("expected error for nonexistent code, got nil")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode SNOMED not found: %v", err)
		}
	})
}

// ==================== RxNorm Tests ====================

func TestRxNormSearchAndGetByCode(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("rxnorm")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	seedRxNormData(t, ctx, tenantID)

	t.Run("Search_ByDisplay", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Metformin", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Metformin' search")
			}
			found := false
			for _, r := range results {
				if r.RxNormCode == "860975" {
					found = true
					if r.GenericName != "Metformin" {
						t.Fatalf("expected generic_name 'Metformin', got %q", r.GenericName)
					}
					if r.DrugClass != "Biguanide" {
						t.Fatalf("expected drug_class 'Biguanide', got %q", r.DrugClass)
					}
					if r.Route != "oral" {
						t.Fatalf("expected route 'oral', got %q", r.Route)
					}
					if r.Form != "tablet" {
						t.Fatalf("expected form 'tablet', got %q", r.Form)
					}
					if r.SystemURI != "http://www.nlm.nih.gov/research/umls/rxnorm" {
						t.Fatalf("expected system_uri 'http://www.nlm.nih.gov/research/umls/rxnorm', got %q", r.SystemURI)
					}
				}
			}
			if !found {
				t.Fatal("expected to find RxNorm code 860975 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search RxNorm by display: %v", err)
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "197361", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for '197361' search")
			}
			found := false
			for _, r := range results {
				if r.RxNormCode == "197361" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find RxNorm code 197361 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search RxNorm by code: %v", err)
		}
	})

	t.Run("Search_ByGenericName", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Lisinopril", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Lisinopril' search")
			}
			found := false
			for _, r := range results {
				if r.RxNormCode == "314076" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find RxNorm code 314076 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search RxNorm by generic name: %v", err)
		}
	})

	t.Run("Search_ByDrugClass", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Statin", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Statin' search")
			}
			found := false
			for _, r := range results {
				if r.RxNormCode == "198440" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find RxNorm code 198440 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search RxNorm by drug class: %v", err)
		}
	})

	t.Run("Search_WithLimit", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "oral", 2)
			if err != nil {
				return err
			}
			if len(results) > 2 {
				t.Fatalf("expected at most 2 results with limit=2, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search RxNorm with limit: %v", err)
		}
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "zzz_nonexistent_zzz", 10)
			if err != nil {
				return err
			}
			if len(results) != 0 {
				t.Fatalf("expected 0 results for nonexistent search, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search RxNorm no results: %v", err)
		}
	})

	t.Run("GetByCode_Found", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			code, err := repo.GetByCode(ctx, "860975")
			if err != nil {
				return err
			}
			if code.RxNormCode != "860975" {
				t.Fatalf("expected code '860975', got %q", code.RxNormCode)
			}
			if code.Display != "Metformin 500 mg oral tablet" {
				t.Fatalf("expected display 'Metformin 500 mg oral tablet', got %q", code.Display)
			}
			if code.GenericName != "Metformin" {
				t.Fatalf("expected generic_name 'Metformin', got %q", code.GenericName)
			}
			if code.DrugClass != "Biguanide" {
				t.Fatalf("expected drug_class 'Biguanide', got %q", code.DrugClass)
			}
			if code.Route != "oral" {
				t.Fatalf("expected route 'oral', got %q", code.Route)
			}
			if code.Form != "tablet" {
				t.Fatalf("expected form 'tablet', got %q", code.Form)
			}
			if code.SystemURI != "http://www.nlm.nih.gov/research/umls/rxnorm" {
				t.Fatalf("expected system_uri 'http://www.nlm.nih.gov/research/umls/rxnorm', got %q", code.SystemURI)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode RxNorm found: %v", err)
		}
	})

	t.Run("GetByCode_NotFound", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewRxNormRepoPG(globalDB.Pool)
			_, err := repo.GetByCode(ctx, "9999999")
			if err == nil {
				t.Fatal("expected error for nonexistent code, got nil")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode RxNorm not found: %v", err)
		}
	})
}

// ==================== CPT Tests ====================

func TestCPTSearchAndGetByCode(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("cpt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	seedCPTData(t, ctx, tenantID)

	t.Run("Search_ByDisplay", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Office visit", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Office visit' search")
			}
			// Should find both 99201 and 99213
			foundNew := false
			foundEstablished := false
			for _, r := range results {
				if r.Code == "99201" {
					foundNew = true
					if r.Category != "E&M" {
						t.Fatalf("expected category 'E&M', got %q", r.Category)
					}
					if r.Subcategory != "Office Visit - New" {
						t.Fatalf("expected subcategory 'Office Visit - New', got %q", r.Subcategory)
					}
					if r.SystemURI != "http://www.ama-assn.org/go/cpt" {
						t.Fatalf("expected system_uri 'http://www.ama-assn.org/go/cpt', got %q", r.SystemURI)
					}
				}
				if r.Code == "99213" {
					foundEstablished = true
				}
			}
			if !foundNew {
				t.Fatal("expected to find CPT code 99201 in results")
			}
			if !foundEstablished {
				t.Fatal("expected to find CPT code 99213 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search CPT by display: %v", err)
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "36415", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for '36415' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "36415" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find CPT code 36415 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search CPT by code: %v", err)
		}
	})

	t.Run("Search_ByCategory", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Radiology", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Radiology' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "71046" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find CPT code 71046 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search CPT by category: %v", err)
		}
	})

	t.Run("Search_BySubcategory", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "Emergency", 10)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				t.Fatal("expected at least one result for 'Emergency' search")
			}
			found := false
			for _, r := range results {
				if r.Code == "99285" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected to find CPT code 99285 in results")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search CPT by subcategory: %v", err)
		}
	})

	t.Run("Search_WithLimit", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "visit", 1)
			if err != nil {
				return err
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result with limit=1, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search CPT with limit: %v", err)
		}
	})

	t.Run("Search_NoResults", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			results, err := repo.Search(ctx, "zzz_nonexistent_zzz", 10)
			if err != nil {
				return err
			}
			if len(results) != 0 {
				t.Fatalf("expected 0 results for nonexistent search, got %d", len(results))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Search CPT no results: %v", err)
		}
	})

	t.Run("GetByCode_Found", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			code, err := repo.GetByCode(ctx, "99213")
			if err != nil {
				return err
			}
			if code.Code != "99213" {
				t.Fatalf("expected code '99213', got %q", code.Code)
			}
			if code.Display != "Office visit, established patient, low complexity" {
				t.Fatalf("expected display 'Office visit, established patient, low complexity', got %q", code.Display)
			}
			if code.Category != "E&M" {
				t.Fatalf("expected category 'E&M', got %q", code.Category)
			}
			if code.Subcategory != "Office Visit - Established" {
				t.Fatalf("expected subcategory 'Office Visit - Established', got %q", code.Subcategory)
			}
			if code.SystemURI != "http://www.ama-assn.org/go/cpt" {
				t.Fatalf("expected system_uri 'http://www.ama-assn.org/go/cpt', got %q", code.SystemURI)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode CPT found: %v", err)
		}
	})

	t.Run("GetByCode_NotFound", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := terminology.NewCPTRepoPG(globalDB.Pool)
			_, err := repo.GetByCode(ctx, "00000")
			if err == nil {
				t.Fatal("expected error for nonexistent code, got nil")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetByCode CPT not found: %v", err)
		}
	})
}
