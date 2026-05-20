/**
 * Pure date utilities. No side effects. No external libraries.
 * All dates are local-time YYYY-MM-DD strings.
 */

/** Returns today's date as "YYYY-MM-DD" in local time. */
export function today(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/**
 * Formats a YYYY-MM-DD string for display.
 * Special-cases "Today" and "Yesterday"; falls back to locale-formatted date.
 */
export function formatDisplay(date: string): string {
  const t = today()
  if (date === t) return 'Today'

  const d = new Date()
  d.setDate(d.getDate() - 1)
  const yest = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  if (date === yest) return 'Yesterday'

  return new Intl.DateTimeFormat(navigator.language, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  }).format(new Date(date + 'T00:00:00'))
}

/** Returns the date one day before the given YYYY-MM-DD string. */
export function prevDate(date: string): string {
  const d = new Date(date + 'T00:00:00')
  d.setDate(d.getDate() - 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

/** Returns the date one day after the given YYYY-MM-DD string. */
export function nextDate(date: string): string {
  const d = new Date(date + 'T00:00:00')
  d.setDate(d.getDate() + 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

/** Returns true when the given date string equals today. */
export function isToday(date: string): boolean {
  return date === today()
}

/** Returns true when the given date string is after today. */
export function isFuture(date: string): boolean {
  return date > today()
}

/** Returns true when the given date string is before today. */
export function isPast(date: string): boolean {
  return date < today()
}
