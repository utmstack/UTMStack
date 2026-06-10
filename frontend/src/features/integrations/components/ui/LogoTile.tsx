import { useState } from 'react'
import { cn } from '@/shared/lib/utils'

interface LogoTileProps {
  src: string
  alt: string
  darkInvert?: boolean
  size?: 'sm' | 'md' | 'lg'
}

const sizeStyles = {
  sm: 'h-14 w-14',
  md: 'h-[65%] w-[65%]',
  lg: 'h-24 w-24',
}

function isImageMonochromaticAndDark(img: HTMLImageElement): boolean {
  try {
    const canvas = document.createElement('canvas')
    canvas.width = img.width
    canvas.height = img.height
    const ctx = canvas.getContext('2d')

    if (!ctx) return false

    ctx.drawImage(img, 0, 0)
    const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const data = imageData.data

    let totalR = 0, totalG = 0, totalB = 0, totalBrightness = 0, transparentPixels = 0
    const pixelCount = data.length / 4

    for (let i = 0; i < data.length; i += 4) {
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      const a = data[i + 3]

      if (a < 30) {
        transparentPixels++
        continue
      }

      totalR += r
      totalG += g
      totalB += b
      totalBrightness += (r + g + b) / 3
    }

    const opaquePixels = pixelCount - transparentPixels
    const transparencyRatio = transparentPixels / pixelCount

    const avgR = totalR / (opaquePixels + 1)
    const avgG = totalG / (opaquePixels + 1)
    const avgB = totalB / (opaquePixels + 1)
    const avgBrightness = totalBrightness / (opaquePixels + 1)

    const maxChannelDiff = Math.max(
      Math.abs(avgR - avgG),
      Math.abs(avgG - avgB),
      Math.abs(avgR - avgB)
    )

    const isMonochromatic = maxChannelDiff < 1
    const isDark = avgBrightness < 1

    return isMonochromatic && isDark && transparencyRatio > 0.3
  } catch {
    return false
  }
}

export function LogoTile({ src, alt, darkInvert, size = 'md' }: LogoTileProps) {
  const [shouldInvert, setShouldInvert] = useState(darkInvert ?? false)

  const handleImageLoad = (e: React.SyntheticEvent<HTMLImageElement>) => {
    const img = e.currentTarget
    if (darkInvert) {
      setShouldInvert(isImageMonochromaticAndDark(img))
    }
  }

  const sizeClass = sizeStyles[size]
  const isFixedSize = size === 'sm' || size === 'lg'

  return (
    <div
      className={cn(
        'relative flex shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-muted/40 p-2.5',
        isFixedSize ? sizeClass : sizeClass
      )}
    >
      <img
        src={src}
        alt={alt}
        onLoad={handleImageLoad}
        className={cn(
          'max-h-full max-w-full object-contain',
          shouldInvert && 'dark:brightness-0 dark:invert'
        )}
        loading="lazy"
      />
    </div>
  )
}
