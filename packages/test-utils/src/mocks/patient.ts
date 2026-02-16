import type { Patient } from '@ehr/fhir-types'

export const mockPatient: Patient = {
  resourceType: 'Patient',
  id: 'test-patient-1',
  meta: { versionId: '1', lastUpdated: '2025-01-15T14:30:00Z' },
  identifier: [
    {
      use: 'usual',
      type: { coding: [{ system: 'http://terminology.hl7.org/CodeSystem/v2-0203', code: 'MR' }], text: 'MRN' },
      system: 'http://hospital.example.org/mrn',
      value: '12345678',
    },
  ],
  active: true,
  name: [
    {
      use: 'official',
      family: 'Smith',
      given: ['John', 'Andrew'],
      prefix: ['Dr.'],
      suffix: ['Jr.'],
    },
  ],
  telecom: [
    { system: 'phone', value: '(555) 123-4567', use: 'mobile' },
    { system: 'email', value: 'john.smith@example.com', use: 'home' },
  ],
  gender: 'male',
  birthDate: '1978-03-15',
  address: [
    {
      use: 'home',
      type: 'physical',
      line: ['123 Main Street', 'Apt 4B'],
      city: 'Springfield',
      state: 'IL',
      postalCode: '62704',
      country: 'US',
    },
  ],
}

export const mockPatientMinimal: Patient = {
  resourceType: 'Patient',
  id: 'test-patient-minimal',
  name: [{ given: ['Jane'], family: 'Doe' }],
  gender: 'female',
}

export const mockPatientEmpty: Patient = {
  resourceType: 'Patient',
  id: 'test-patient-empty',
}
