// Country data and utilities
export const countries = [
  { code: 'KE', name: 'Kenya', dialCode: '+254' },
  { code: 'NG', name: 'Nigeria', dialCode: '+234' },
  { code: 'ZA', name: 'South Africa', dialCode: '+27' },
  { code: 'ZM', name: 'Zambia', dialCode: '+260' },
]

export function formatCountryOption(country: typeof countries[0]) {
  return `${country.name} (${country.dialCode})`
}

export function formatCountrySelectDisplay(country: typeof countries[0] | undefined) {
  if (!country) return 'Select a country'
  return country.name
}

export function findCountryByDialCode(dialCode: string) {
  return countries.find((c) => c.dialCode === dialCode)
}

export function findCountryByCode(code: string) {
  return countries.find((c) => c.code === code)
}
