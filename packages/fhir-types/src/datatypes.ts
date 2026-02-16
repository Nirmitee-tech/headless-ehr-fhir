import type { FHIRCode, FHIRDateTime, FHIRDecimal, FHIRMarkdown, FHIRString, FHIRUri, FHIRUrl, FHIRPositiveInt, FHIRBoolean, FHIRBase64Binary, FHIRDate, FHIRInstant, FHIRUnsignedInt } from './primitives'

export interface Coding {
  system?: FHIRUri
  version?: FHIRString
  code?: FHIRCode
  display?: FHIRString
  userSelected?: FHIRBoolean
}

export interface CodeableConcept {
  coding?: Coding[]
  text?: FHIRString
}

export interface HumanName {
  use?: 'usual' | 'official' | 'temp' | 'nickname' | 'anonymous' | 'old' | 'maiden'
  text?: FHIRString
  family?: FHIRString
  given?: FHIRString[]
  prefix?: FHIRString[]
  suffix?: FHIRString[]
  period?: Period
}

export interface Address {
  use?: 'home' | 'work' | 'temp' | 'old' | 'billing'
  type?: 'postal' | 'physical' | 'both'
  text?: FHIRString
  line?: FHIRString[]
  city?: FHIRString
  district?: FHIRString
  state?: FHIRString
  postalCode?: FHIRString
  country?: FHIRString
  period?: Period
}

export interface ContactPoint {
  system?: 'phone' | 'fax' | 'email' | 'pager' | 'url' | 'sms' | 'other'
  value?: FHIRString
  use?: 'home' | 'work' | 'temp' | 'old' | 'mobile'
  rank?: FHIRPositiveInt
  period?: Period
}

export interface Identifier {
  use?: 'usual' | 'official' | 'temp' | 'secondary' | 'old'
  type?: CodeableConcept
  system?: FHIRUri
  value?: FHIRString
  period?: Period
  assigner?: Reference
}

export interface Reference {
  reference?: FHIRString
  type?: FHIRUri
  identifier?: Identifier
  display?: FHIRString
}

export interface Period {
  start?: FHIRDateTime
  end?: FHIRDateTime
}

export interface Quantity {
  value?: FHIRDecimal
  comparator?: '<' | '<=' | '>=' | '>'
  unit?: FHIRString
  system?: FHIRUri
  code?: FHIRCode
}

export interface Range {
  low?: Quantity
  high?: Quantity
}

export interface Ratio {
  numerator?: Quantity
  denominator?: Quantity
}

export interface Money {
  value?: FHIRDecimal
  currency?: FHIRCode
}

export interface Annotation {
  authorReference?: Reference
  authorString?: FHIRString
  time?: FHIRDateTime
  text: FHIRMarkdown
}

export interface Attachment {
  contentType?: FHIRCode
  language?: FHIRCode
  data?: FHIRBase64Binary
  url?: FHIRUrl
  size?: FHIRUnsignedInt
  hash?: FHIRBase64Binary
  title?: FHIRString
  creation?: FHIRDateTime
}

export interface Narrative {
  status: 'generated' | 'extensions' | 'additional' | 'empty'
  div: FHIRString
}

export interface Dosage {
  sequence?: number
  text?: FHIRString
  timing?: Timing
  route?: CodeableConcept
  method?: CodeableConcept
  doseAndRate?: Array<{
    type?: CodeableConcept
    doseQuantity?: Quantity
    doseRange?: Range
    rateQuantity?: Quantity
    rateRange?: Range
    rateRatio?: Ratio
  }>
  maxDosePerPeriod?: Ratio
  maxDosePerAdministration?: Quantity
  maxDosePerLifetime?: Quantity
}

export interface Timing {
  event?: FHIRDateTime[]
  repeat?: {
    boundsDuration?: Quantity
    boundsPeriod?: Period
    boundsRange?: Range
    count?: FHIRPositiveInt
    countMax?: FHIRPositiveInt
    duration?: FHIRDecimal
    durationMax?: FHIRDecimal
    durationUnit?: 's' | 'min' | 'h' | 'd' | 'wk' | 'mo' | 'a'
    frequency?: FHIRPositiveInt
    frequencyMax?: FHIRPositiveInt
    period?: FHIRDecimal
    periodMax?: FHIRDecimal
    periodUnit?: 's' | 'min' | 'h' | 'd' | 'wk' | 'mo' | 'a'
    dayOfWeek?: FHIRCode[]
    timeOfDay?: string[]
    when?: FHIRCode[]
    offset?: FHIRUnsignedInt
  }
  code?: CodeableConcept
}

export interface Meta {
  versionId?: FHIRString
  lastUpdated?: FHIRInstant
  source?: FHIRUri
  profile?: FHIRUri[]
  tag?: Coding[]
}

/** Base for all FHIR resources */
export interface Resource {
  resourceType: string
  id?: FHIRString
  meta?: Meta
  language?: FHIRCode
  text?: Narrative
}

/** Base for domain resources (most clinical resources) */
export interface DomainResource extends Resource {
  contained?: Resource[]
  extension?: Extension[]
  modifierExtension?: Extension[]
}

export interface Extension {
  url: FHIRUri
  valueString?: FHIRString
  valueBoolean?: FHIRBoolean
  valueCode?: FHIRCode
  valueDate?: FHIRDate
  valueDateTime?: FHIRDateTime
  valueDecimal?: FHIRDecimal
  valueInteger?: number
  valueUri?: FHIRUri
  valueCoding?: Coding
  valueCodeableConcept?: CodeableConcept
  valueQuantity?: Quantity
  valueReference?: Reference
  valuePeriod?: Period
  valueIdentifier?: Identifier
  valueHumanName?: HumanName
  valueAddress?: Address
  valueContactPoint?: ContactPoint
  valueAttachment?: Attachment
  valueMoney?: Money
  valueAnnotation?: Annotation
}

/** Bundle entry */
export interface BundleEntry<T extends Resource = Resource> {
  fullUrl?: FHIRUri
  resource?: T
  search?: { mode?: 'match' | 'include' | 'outcome'; score?: FHIRDecimal }
  request?: { method: 'GET' | 'HEAD' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'; url: FHIRUri }
  response?: { status: FHIRString; location?: FHIRUri; etag?: FHIRString; lastModified?: FHIRInstant }
}

/** Bundle (search results, transactions, etc.) */
export interface Bundle<T extends Resource = Resource> extends Resource {
  resourceType: 'Bundle'
  type: 'searchset' | 'batch' | 'transaction' | 'batch-response' | 'transaction-response' | 'history' | 'document' | 'message' | 'collection'
  total?: FHIRUnsignedInt
  link?: Array<{ relation: FHIRString; url: FHIRUri }>
  entry?: BundleEntry<T>[]
}

/** OperationOutcome — FHIR error/info response */
export interface OperationOutcome extends Resource {
  resourceType: 'OperationOutcome'
  issue: Array<{
    severity: 'fatal' | 'error' | 'warning' | 'information'
    code: FHIRCode
    details?: CodeableConcept
    diagnostics?: FHIRString
    location?: FHIRString[]
    expression?: FHIRString[]
  }>
}
