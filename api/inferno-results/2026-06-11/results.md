# Inferno ONC (g)(10) Test Results — 2026-06-11

Suite: **SMART App Launch — Standalone Patient App** (ONC Certification (g)(10) Standardized API Test Kit)
Server: headless-ehr-fhir, built-in SMART on FHIR server (standalone mode)

## Summary: 44/47 passing (1 functional failures, 2 TLS-only failures expected in HTTP dev mode)

| Test | Result | Notes |
|------|--------|-------|
| well_known_endpoint | ✅ pass |  |
| well_known_capabilities_stu2 | ✅ pass |  |
| g10_smart_well_known_capabilities | ❌ fail | The following capabilities required for this scenario are missing: client-confidential-asymmetric |
| standalone_auth_tls | ⚠️ fail (TLS-only) | Server did not support any allowed TLS versions. |
| smart_app_redirect_stu2 | ✅ pass |  |
| smart_code_received | ✅ pass |  |
| standalone_token_tls | ⚠️ fail (TLS-only) | Server did not support any allowed TLS versions. |
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
| Test01 | ✅ pass | Scopes received indicate access to all necessary resources. |
| g10_patient_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_allergy_intolerance_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_care_plan_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_care_team_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_condition_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_device_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_diagnostic_report_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_document_reference_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_goal_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_immunization_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_medication_request_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_observation_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_procedure_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_us_core_6_encounter_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_us_core_6_service_request_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_us_core_6_coverage_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_us_core_6_medication_dispense_unrestricted_access | ✅ pass | Access expected to be granted and request properly returned 200 |
| g10_standalone_credentials_export | ✅ pass |  |
| g10_auth_incorrectly_permitted_tls_versions_messages_setup | ✅ pass |  |
| g10_token_incorrectly_permitted_tls_versions_messages_setup | ✅ pass |  |

Session ID: `h5GbI589F4z` — reproduce by running Inferno against this repo (see README).
