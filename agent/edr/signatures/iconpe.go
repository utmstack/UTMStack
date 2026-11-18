package signatures

import "encoding/binary"

// Icon extraction, as the reference's debug output shows it: every language
// of the first group-icon ID (resource type 14) is scanned, stopping at the
// next ID; each group entry names an icon (type 3) by id. A PNG
// image and a bitmap outside the reference's size/aspect limits are skipped.
// Checked on the 237 programs with icons of a Windows system directory (all
// 1,852 icon hashes) and on generated programs.
//
// Malformed directory reads terminate the modeled lookup and preserve its
// earlier callbacks. Bitmap offsets use the complete physical object,
// including uint32 header-offset wrap. A file-backed bitmap body that is too
// short for the reference's pixel read is a definite skipped image.

const (
	maxGroupIcons = 100 // ClamAV's default MaxIconsPE entry budget
	maxIconWidth  = 256
	minIconWidth  = 16
)

// programIcon is one parsed icon: its hash where this engine computes it,
// and whether its color and gray areas are unknown.
type programIcon struct {
	hash     iconHash
	grayMode bool
}

// programIcons parses the first group ID's images across its languages. ok is
// false when the group walk or an image is uncertain.
func (m *peMeta) programIcons(data []byte) (icons []programIcon, ok bool) {
	icons, ok, _ = m.programIconsWithWork(data, nil, nil)
	return icons, ok
}

func (m *peMeta) programIconsWithWork(data []byte, work *matchWork, matched func(programIcon) (bool, error)) (icons []programIcon, known bool, err error) {
	le := binary.LittleEndian
	var firstID uint32
	haveGroup, uncertain, foundMatch := false, false, false
	visited := 0
	_, err = m.walkResources(data, 14, ^uint32(0), work, func(name, groupEntry uint32) (bool, error) {
		if haveGroup && name != firstID {
			return true, nil
		}
		haveGroup, firstID = true, name
		group, found := m.resourceData(data, groupEntry)
		if !found || len(group) < 6 {
			return false, nil
		}
		count := min(int(le.Uint16(group[4:])), (len(group)-6)/14)
		for i := 0; i < count; i++ {
			id := uint32(int32(int16(le.Uint16(group[6+14*i+12:]))))
			_, err := m.walkResources(data, 3, id, work, func(_ uint32, entry uint32) (bool, error) {
				img, found := m.iconFileData(data, entry)
				if !found {
					return false, nil
				}
				// Charge bounded bitmap work before allocation/conversion. The
				// hash has at most 32x32 pixels; its area searches cost O(side^4).
				if work != nil && len(img) >= 40 {
					width, height := le.Uint32(img[4:]), le.Uint32(img[8:])/2
					if width >= 16 && width <= 256 && height >= 16 && height <= 256 {
						side := 32
						if width == height {
							side = int(width)
							if width%24 == 0 || width%32 == 0 {
								for side > 32 {
									side /= 2
								}
							} else if width <= 20 {
								side = 16
							} else if width <= 28 {
								side = 24
							} else {
								side = 32
							}
						}
						cost := int64(width)*int64(height) + int64(side)*int64(side)*int64(side)*int64(side)
						if err := work.step(cost); err != nil {
							return true, err
						}
					}
				}
				icon, skip, sure := parseIconImageFile(data, len(data)-len(img))
				if !sure {
					uncertain = true
					return true, nil
				}
				if !skip {
					icons = append(icons, icon)
					if matched != nil {
						hit, e := matched(icon)
						if hit || e != nil {
							foundMatch = hit
							return true, e
						}
					}
				}
				return false, nil
			})
			if err != nil || uncertain || foundMatch {
				return true, err
			}
			visited++
			if visited >= maxGroupIcons {
				return true, nil
			}
		}
		return false, nil
	})
	// A physical read failure ends findres without inventing another icon.
	// The resulting parsed prefix remains usable, including an earlier match.
	return icons, !uncertain && !foundMatch, err
}

// iconFileData returns the remaining mapped file from an icon's bitmap RVA.
// ClamAV's parseicon reads that RVA from the resource data entry, then reads
// the bitmap from the file map without consulting the entry's declared size.
func (m *peMeta) iconFileData(data []byte, entry uint32) ([]byte, bool) {
	at, ok := m.rawAddr(entry)
	if !ok || at+4 > len(data) {
		return nil, false
	}
	start, ok := m.rawAddr(binary.LittleEndian.Uint32(data[at:]))
	if !ok || start > len(data) {
		return nil, false
	}
	return data[start:], true
}

// resourceData reads the bytes a resource data entry points to.
func (m *peMeta) resourceData(data []byte, entry uint32) ([]byte, bool) {
	le := binary.LittleEndian
	at, ok := m.rawAddr(entry)
	if !ok || at+16 > len(data) {
		return nil, false
	}
	start, ok := m.rawAddr(le.Uint32(data[at:]))
	size := uint64(le.Uint32(data[at+4:]))
	if !ok || uint64(start)+size > uint64(len(data)) {
		return nil, false
	}
	return data[start : start+int(size)], true
}

type iconLookupState uint8

const (
	iconAbsent iconLookupState = iota
	iconFound
)

// resourcesByID lists the data entries of every language of the resource
// with the given signed icon id below the first directory of the type.
// An id of -1 selects every name. Duplicate IDs retain callback order;
// non-directory names and nested languages are skipped. A physical read
// failure ends the lookup and retains the already delivered prefix.
func (m *peMeta) resourcesByID(data []byte, typeID, id uint32) (entries []uint32, state iconLookupState) {
	m.walkResources(data, typeID, id, nil, func(_ uint32, entry uint32) (bool, error) { entries = append(entries, entry); return false, nil })
	if len(entries) == 0 {
		return nil, iconAbsent
	}
	return entries, iconFound
}

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// parseIconImage parses an icon image from its file-backed suffix. The
// reference does not use the group entry's declared dimensions or the image
// data entry's declared length. skip is true for an image the reference skips;
// sure is false when this engine cannot be sure what it does.
func parseIconImage(b []byte) (icon programIcon, skip, sure bool) {
	return parseIconImageAt(b, 0)
}

// The reference advances a uint32 file offset past the declared bitmap
// header. This suffix-only helper cannot access earlier physical bytes;
// production uses parseIconImageFile with the complete mapped object.
func parseIconImageAt(b []byte, fileOffset uint64) (icon programIcon, skip, sure bool) {
	if len(b) < 40 {
		return icon, true, true
	}
	size := uint64(binary.LittleEndian.Uint32(b))
	if size > uint64(len(b)) && fileOffset > uint64(^uint32(0))-size {
		return icon, false, false
	}
	return parseIconBitmap(b, b, size)
}

// parseIconImageFile retains the complete physical object. ClamAV advances a
// uint32 file offset, so an oversized header can wrap to earlier palette or
// pixel bytes. The mapped file, not the bitmap suffix, bounds those reads.
func parseIconImageFile(data []byte, at int) (icon programIcon, skip, sure bool) {
	if at < 0 || at > len(data)-40 || uint64(at) > uint64(^uint32(0)) {
		return icon, true, true
	}
	header := data[at : at+40]
	size := binary.LittleEndian.Uint32(header)
	if size < 40 {
		return icon, true, true
	}
	pixels := uint32(at) + size
	return parseIconBitmap(header, data, uint64(pixels))
}

func parseIconBitmap(header, b []byte, paletteAt uint64) (icon programIcon, skip, sure bool) {
	le := binary.LittleEndian
	if len(header) < 40 || le.Uint32(header) < 40 {
		return icon, true, true
	}
	w, h := uint64(le.Uint32(header[4:])), uint64(le.Uint32(header[8:]))/2
	if w < minIconWidth || w > maxIconWidth || h < minIconWidth || h > maxIconWidth || w < h*3/4 || h < w*3/4 {
		return icon, true, true
	}
	width, height := int(w), int(h)
	bpp := int(le.Uint16(header[14:]))
	palette := 0
	switch bpp {
	case 1, 4, 8:
		palette = 1 << bpp
	case 16, 24, 32:
	default:
		return icon, true, true
	}
	if paletteAt > uint64(len(b)) || uint64(palette*4) > uint64(len(b))-paletteAt {
		return icon, true, true
	}
	stride := (width*bpp + 31) / 32 * 4
	maskStride := (width + 31) / 32 * 4
	pixels := int(uint32(paletteAt) + uint32(4*palette))
	mask := pixels + stride*height
	if pixels > len(b) || bpp != 32 && uint64(mask)+uint64(maskStride*height) > uint64(len(b)) || mask > len(b) {
		return icon, true, true
	}
	px := make([]iconPixel, width*height)
	anyAlpha := false
	if bpp == 32 {
		for y := 0; y < height && !anyAlpha; y++ {
			row := b[pixels+stride*y:]
			for x := 0; x < width; x++ {
				if row[4*x+3] != 0 {
					anyAlpha = true
					break
				}
			}
		}
	}
	// ClamAV treats an all-zero 32-bit alpha channel as 24-bit color:
	// apply the AND mask if complete, otherwise force every pixel opaque.
	useMask := bpp != 32 || !anyAlpha && mask+maskStride*height <= len(b)
	for y := 0; y < height; y++ {
		row := b[pixels+stride*(height-1-y):]
		for x := 0; x < width; x++ {
			var r, g, bl, a int
			switch bpp {
			case 1, 4, 8:
				bit := x * bpp
				i := int(row[bit/8]>>(8-bpp-bit%8)) & (palette - 1)
				p := b[int(paletteAt)+4*i:]
				bl, g, r, a = int(p[0]), int(p[1]), int(p[2]), int(p[3])
			case 16:
				// Match ClamAV 1.5.4's 16-bit icon conversion, including
				// its shifts into the 32-bit color word. These differ from
				// a conventional RGB555 decoder.
				lo, hi := uint32(row[2*x]), uint32(row[2*x+1])
				blue := lo & 0x1f
				green := (lo >> 5) | ((hi & 3) << 3)
				red := hi & 0xfc
				c := ((red<<3 | red>>2) << 17) |
					((green<<3 | green>>2) << 11) | (blue<<3 | blue>>2)
				bl, g, r, a = int(c&0xff), int((c>>8)&0xff), int((c>>16)&0xff), int(c>>24)
			case 24:
				bl, g, r = int(row[3*x]), int(row[3*x+1]), int(row[3*x+2])
			case 32:
				bl, g, r, a = int(row[4*x]), int(row[4*x+1]), int(row[4*x+2]), int(row[4*x+3])
			}
			if useMask {
				if b[mask+maskStride*(height-1-y)+x/8]>>(7-x%8)&1 == 0 {
					a = 255
				}
				// A set AND bit retains the decoded alpha, including a
				// palette entry's high byte and the reference's RGB16 bits.
			} else if bpp == 32 && !anyAlpha {
				a = 255
			}
			px[y*width+x] = iconPixel{
				(r*a + 255*(255-a)) / 255,
				(g*a + 255*(255-a)) / 255,
				(bl*a + 255*(255-a)) / 255,
			}
		}
	}
	var scaled []iconPixel
	var sw int
	if width == height {
		scaled, sw = iconScale(px, width)
	} else {
		scaled, sw = iconScaleRect(px, width, height)
	}
	icon.hash, icon.grayMode = iconMeasures(scaled, sw)
	return icon, false, true
}
