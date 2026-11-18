package signatures

import "bytes"

// referenceJPEGProfile limits exact fuzzy-hash decisions to the ordinary
// JFIF and Adobe CMYK layouts covered by the pinned reference's decoder and
// the broad differential corpora. Larger images retain the resizer's
// half-integer rounding guard; other layouts keep decoder uncertainty.
func referenceJPEGProfile(data []byte) bool {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return false
	}
	jfif, adobeCMYK, frame := false, false, false
	components := 0
	for p := 2; p+3 < len(data); {
		if data[p] != 0xff {
			return false
		}
		for p < len(data) && data[p] == 0xff {
			p++
		}
		if p >= len(data) {
			return false
		}
		marker := data[p]
		p++
		if marker == 0xda { // Start of Scan.
			return frame && ((components == 4 && adobeCMYK && !jfif) ||
				(components != 4 && jfif && !adobeCMYK))
		}
		if marker == 0xd9 || marker == 0xd8 || marker == 0x01 ||
			(marker >= 0xd0 && marker <= 0xd7) || p+2 > len(data) {
			return false
		}
		n := int(data[p])<<8 | int(data[p+1])
		if n < 2 || n > len(data)-p {
			return false
		}
		segment := data[p+2 : p+n]
		p += n
		switch marker {
		case 0xe0:
			if bytes.HasPrefix(segment, []byte("JFIF\x00")) {
				jfif = true
			}
		case 0xee:
			// Adobe transform 0 stores inverted CMYK component samples.
			if len(segment) != 12 || !bytes.HasPrefix(segment, []byte("Adobe")) || segment[11] != 0 || adobeCMYK {
				return false
			}
			adobeCMYK = true
		case 0xdd: // Restart interval; the decoder handles MCU resets.
			if len(segment) != 2 {
				return false
			}
		case 0xc0, 0xc2:
			if frame || len(segment) < 6 || segment[0] != 8 {
				return false
			}
			height := int(segment[1])<<8 | int(segment[2])
			width := int(segment[3])<<8 | int(segment[4])
			if width == 0 || height == 0 || width > 2048 || height > 2048 || width*height > 1<<20 {
				return false
			}
			components = int(segment[5])
			if len(segment) != 6+3*components {
				return false
			}
			switch components {
			case 1:
				if segment[6] != 1 || (segment[7] != 0x11 && segment[7] != 0x21 && segment[7] != 0x22) {
					return false
				}
			case 3:
				if segment[6] != 1 || (segment[7] != 0x11 && segment[7] != 0x21 && segment[7] != 0x22) ||
					segment[9] != 2 || segment[10] != 0x11 || segment[12] != 3 || segment[13] != 0x11 {
					return false
				}
			case 4:
				if segment[6] != 'C' || (segment[7] != 0x11 && segment[7] != 0x21 && segment[7] != 0x22) ||
					segment[9] != 'M' || segment[10] != 0x11 ||
					segment[12] != 'Y' || segment[13] != 0x11 ||
					segment[15] != 'K' || segment[16] != 0x11 {
					return false
				}
			default:
				return false
			}
			frame = true
		default:
			// Reject unmodeled frame and hierarchical coding markers.
			if marker >= 0xc0 && marker <= 0xcf && marker != 0xc4 && marker != 0xc8 && marker != 0xcc {
				return false
			}
		}
	}
	return false
}
