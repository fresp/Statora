import qrcode from 'qrcode'

export function generateQrCodeDataUrl(text: string): Promise<string> {
  return qrcode.toDataURL(text, { errorCorrectionLevel: 'H' })
}
