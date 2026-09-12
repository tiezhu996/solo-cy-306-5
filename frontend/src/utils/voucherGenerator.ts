export function generateVoucherNo(): string {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  const date = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}`
  const rand = Math.floor(Math.random() * 10000).toString().padStart(4, '0')
  return `GB${date}${rand}`
}

export function validateVoucherFormat(voucher: string): boolean {
  const v = voucher.trim()
  return v.length >= 8 && v.startsWith('GB')
}
