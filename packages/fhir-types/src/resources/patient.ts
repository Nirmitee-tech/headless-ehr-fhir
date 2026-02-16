import type { DomainResource, HumanName, Address, ContactPoint, Identifier, CodeableConcept, Reference, Period, Attachment } from '../datatypes'
import type { FHIRDate, FHIRBoolean, FHIRDateTime } from '../primitives'

export interface Patient extends DomainResource {
  resourceType: 'Patient'
  identifier?: Identifier[]
  active?: FHIRBoolean
  name?: HumanName[]
  telecom?: ContactPoint[]
  gender?: 'male' | 'female' | 'other' | 'unknown'
  birthDate?: FHIRDate
  deceasedBoolean?: FHIRBoolean
  deceasedDateTime?: FHIRDateTime
  address?: Address[]
  maritalStatus?: CodeableConcept
  multipleBirthBoolean?: FHIRBoolean
  multipleBirthInteger?: number
  photo?: Attachment[]
  contact?: PatientContact[]
  communication?: PatientCommunication[]
  generalPractitioner?: Reference[]
  managingOrganization?: Reference
  link?: PatientLink[]
}

export interface PatientContact {
  relationship?: CodeableConcept[]
  name?: HumanName
  telecom?: ContactPoint[]
  address?: Address
  gender?: 'male' | 'female' | 'other' | 'unknown'
  organization?: Reference
  period?: Period
}

export interface PatientCommunication {
  language: CodeableConcept
  preferred?: FHIRBoolean
}

export interface PatientLink {
  other: Reference
  type: 'replaced-by' | 'replaces' | 'refer' | 'seealso'
}
