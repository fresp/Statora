import React, { useEffect, useState } from 'react'
import { generateQrCodeDataUrl } from '../../utils/qrCodeGenerator'

interface QrCodeDisplayProps {
  text: string
  size?: number
}

const QrCodeDisplay: React.FC<QrCodeDisplayProps> = ({ text, size = 256 }) => {
  const [qrCodeDataUrl, setQrCodeDataUrl] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let isMounted = true

    const generateQr = async () => {
      try {
        setError(null)
        const url = await generateQrCodeDataUrl(text)
        if (isMounted) {
          setQrCodeDataUrl(url)
        }
      } catch (err) {
        console.error('Error generating QR code:', err)
        if (isMounted) {
          setQrCodeDataUrl(null)
          setError('Failed to generate QR code.')
        }
      }
    }

    if (text) {
      void generateQr()
    } else {
      setQrCodeDataUrl(null)
      setError(null)
    }

    return () => {
      isMounted = false
    }
  }, [text])

  if (error) {
    return <div className="text-red-500">{error}</div>
  }

  if (!qrCodeDataUrl) {
    return <div className="text-sm text-gray-500">Loading QR code...</div>
  }

  return (
    <img src={qrCodeDataUrl} alt="QR Code" width={size} height={size} />
  )
}

export default QrCodeDisplay
