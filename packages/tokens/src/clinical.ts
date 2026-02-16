/** Clinical severity levels — used in AllergyList, CDSHooksCard, alerts */
export const severity = {
  critical: { fg: '#FFFFFF', bg: '#B91C1C', border: '#991B1B', label: 'Critical' },
  high: { fg: '#B91C1C', bg: '#FEF2F2', border: '#FCA5A5', label: 'High' },
  moderate: { fg: '#C2410C', bg: '#FFF7ED', border: '#FDBA74', label: 'Moderate' },
  low: { fg: '#B45309', bg: '#FFFBEB', border: '#FCD34D', label: 'Low' },
  info: { fg: '#1D4ED8', bg: '#EFF6FF', border: '#BFDBFE', label: 'Info' },
} as const

/** FHIR resource status colors — used in ProblemList, MedicationList, CarePlan */
export const status = {
  active: { fg: '#15803D', bg: '#F0FDF4', border: '#BBF7D0', dot: '#22C55E' },
  inactive: { fg: '#6B7280', bg: '#F9FAFB', border: '#E5E7EB', dot: '#9CA3AF' },
  resolved: { fg: '#6B7280', bg: '#F9FAFB', border: '#E5E7EB', dot: '#9CA3AF' },
  draft: { fg: '#B45309', bg: '#FFFBEB', border: '#FDE68A', dot: '#F59E0B' },
  on_hold: { fg: '#EA580C', bg: '#FFF7ED', border: '#FED7AA', dot: '#FB923C' },
  completed: { fg: '#2563EB', bg: '#EFF6FF', border: '#BFDBFE', dot: '#3B82F6' },
  cancelled: { fg: '#6B7280', bg: '#F9FAFB', border: '#E5E7EB', dot: '#9CA3AF' },
  entered_in_error: { fg: '#DC2626', bg: '#FEF2F2', border: '#FECACA', dot: '#EF4444' },
} as const

/** Lab result flag colors — used in LabResults, VitalsPanel */
export const labFlags = {
  normal: { fg: 'inherit', bg: 'transparent', weight: '400', flag: '' },
  abnormal_high: { fg: '#B91C1C', bg: '#FEF2F2', weight: '600', flag: 'H' },
  abnormal_low: { fg: '#1D4ED8', bg: '#EFF6FF', weight: '600', flag: 'L' },
  critical_high: { fg: '#FFFFFF', bg: '#B91C1C', weight: '700', flag: 'HH' },
  critical_low: { fg: '#FFFFFF', bg: '#1E40AF', weight: '700', flag: 'LL' },
} as const

/** Task priority levels */
export const priority = {
  stat: { fg: '#FFFFFF', bg: '#DC2626', label: 'STAT' },
  urgent: { fg: '#C2410C', bg: '#FFF7ED', label: 'Urgent' },
  routine: { fg: '#2563EB', bg: '#EFF6FF', label: 'Routine' },
  elective: { fg: '#6B7280', bg: '#F3F4F6', label: 'Elective' },
} as const

/** Encounter type colors for timeline/calendar */
export const encounterColors = {
  office_visit: '#3B82F6',
  telehealth: '#14B8A6',
  emergency: '#DC2626',
  inpatient: '#9333EA',
  observation: '#F97316',
  procedure: '#16A34A',
  imaging: '#D97706',
  lab: '#0D9488',
} as const

export type SeverityLevel = keyof typeof severity
export type StatusType = keyof typeof status
export type LabFlagType = keyof typeof labFlags
export type PriorityLevel = keyof typeof priority
