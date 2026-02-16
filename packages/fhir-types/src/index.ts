// FHIR R4 primitive types
export type {
  FHIRString,
  FHIRBoolean,
  FHIRInteger,
  FHIRDecimal,
  FHIRUri,
  FHIRUrl,
  FHIRCanonical,
  FHIRBase64Binary,
  FHIRInstant,
  FHIRDate,
  FHIRDateTime,
  FHIRTime,
  FHIRCode,
  FHIROid,
  FHIRId,
  FHIRMarkdown,
  FHIRUnsignedInt,
  FHIRPositiveInt,
  FHIRUuid,
} from './primitives'

// FHIR R4 data types
export type {
  Coding,
  CodeableConcept,
  HumanName,
  Address,
  ContactPoint,
  Identifier,
  Reference,
  Period,
  Quantity,
  Range,
  Ratio,
  Money,
  Annotation,
  Attachment,
  Narrative,
  Dosage,
  Timing,
  Meta,
  Resource,
  DomainResource,
  Extension,
  BundleEntry,
  Bundle,
  OperationOutcome,
} from './datatypes'

// FHIR R4 resources
export type { Patient, PatientContact, PatientCommunication, PatientLink } from './resources/patient'
