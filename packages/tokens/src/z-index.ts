export const zIndex = {
  hide: -1,
  base: 0,
  dropdown: 10,
  sticky: 20,
  overlay: 30,
  modal: 40,
  popover: 50,
  toast: 60,
  alert: 70,
  max: 99,
} as const

export type ZIndexToken = keyof typeof zIndex
