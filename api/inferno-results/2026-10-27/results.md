# Inferno ONC (g)(10) Test Results: 2026-10-27

Suite: **SMART App Launch: Standalone Patient App** (ONC Certification (g)(10) Standardized API Test Kit)
Kit: onc-healthit/onc-certification-g10-test-kit **v8.0.7** (GitHub release 9 Sep 2026), run locally in Docker.
Options: US Core 6.1.0 / USCDI v3, SMART App Launch 2.0.0.
Server: headless-ehr-fhir, built-in SMART on FHIR server (standalone mode), reached over HTTPS. Client: confidential symmetric, PKCE S256.
Run time: 27 Oct 2026, 17:08 to 17:09 UTC. Includes the wildcard scope fix merged as #1.

## Summary: 47/47 passing

Scope: this is the Standalone Patient App group only. The full (g)(10) kit has 641 tests in 7 groups; the other groups
(Single Patient API, Limited Access, EHR Practitioner App, Multi-Patient/Bulk, Additional Authorization, Visual Inspection)
are not claimed by this run. Raw kit output is not published because it contains test tokens and tunnel URLs.

| Test | Result | Notes |
|------|--------|-------|
| well_known_endpoint | ✅ pass |  |
| well_known_capabilities_stu2 | ✅ pass |  |
| g10_smart_well_known_capabilities | ✅ pass |  |
| standalone_auth_tls | ✅ pass |  |
| smart_app_redirect_stu2 | ✅ pass |  |
| smart_code_received | ✅ pass |  |
| standalone_token_tls | ✅ pass |  |
| smart_token_exchange | ✅ pass |  |
| smart_token_response_body | ✅ pass |  |
| smart_token_response_headers | ✅ pass |  |
| g10_smart_scopes | ✅ pass |  |
| g10_unauthorized_access | ✅ pass |  |
| g10_patient_context | ✅ pass |  |
| smart_openid_decode_id_token | ✅ pass |  |
| smart_openid_retrieve_configuration | ✅ pass |  |
| smart_openid_required_configuration_fields | ✅ pass |  |
| smart_openid_retrieve_jwks | ✅ pass |  |
| smart_openid_token_header | ✅ pass |  |
| smart_openid_token_payload | ✅ pass |  |
| smart_openid_fhir_user_claim | ✅ pass |  |
| g10_token_refresh_without_scopes | ✅ pass |  |
| g10_token_refresh_body_without_scopes | ✅ pass |  |
| g10_token_refresh_with_scopes | ✅ pass |  |
| g10_token_refresh_body_with_scopes | ✅ pass |  |
| g10_patient_context | ✅ pass |  |
| g10_invalid_token_refresh | ✅ pass |  |
| Test01 | ✅ pass |  |
| g10_patient_unrestricted_access | ✅ pass |  |
| g10_allergy_intolerance_unrestricted_access | ✅ pass |  |
| g10_care_plan_unrestricted_access | ✅ pass |  |
| g10_care_team_unrestricted_access | ✅ pass |  |
| g10_condition_unrestricted_access | ✅ pass |  |
| g10_device_unrestricted_access | ✅ pass |  |
| g10_diagnostic_report_unrestricted_access | ✅ pass |  |
| g10_document_reference_unrestricted_access | ✅ pass |  |
| g10_goal_unrestricted_access | ✅ pass |  |
| g10_immunization_unrestricted_access | ✅ pass |  |
| g10_medication_request_access | ✅ pass |  |
| g10_observation_unrestricted_access | ✅ pass |  |
| g10_procedure_unrestricted_access | ✅ pass |  |
| g10_us_core_6_encounter_unrestricted_access | ✅ pass |  |
| g10_us_core_6_service_request_unrestricted_access | ✅ pass |  |
| g10_us_core_6_coverage_unrestricted_access | ✅ pass |  |
| g10_us_core_6_medication_dispense_unrestricted_access | ✅ pass |  |
| g10_standalone_credentials_export | ✅ pass |  |
| g10_auth_incorrectly_permitted_tls_versions_messages_setup | ✅ pass |  |
| g10_token_incorrectly_permitted_tls_versions_messages_setup | ✅ pass |  |
