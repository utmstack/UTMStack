package signatures

import (
	"encoding/hex"
	"errors"
)

// A file-type database (.ftm) teaches the engine how to type objects. Each
// line is MagicType:Offset:HexSig:Name:RequiredType:DetectedType with an
// optional functionality window. Magic type 0 is a fixed byte comparison at
// an offset, type 1 a type pattern with the full body and offset syntax, and
// type 4 a partition magic; other magic types are skipped, as the reference
// skips them.

// ftmMagic is a fixed byte comparison at an offset.
type ftmMagic struct {
	offset int
	magic  []byte
	typ    FileType
}

// ftmPattern is a type pattern. It applies when the running type is rtype or
// rtype is ANY.
type ftmPattern struct {
	id         int // in the database's pattern table
	rtype, typ FileType
}

type fileTypes struct {
	// magics are tried in order. The reference prepends each line, so the
	// last line of a file is tried first.
	magics     []ftmMagic
	patterns   []ftmPattern
	partitions []ftmMagic
}

// SkipFileTypeKind is a file-type line of a magic type the reference does
// not load.
const SkipFileTypeKind = "file-type-kind"

func (d *Database) addFileTypeLine(line string) error {
	p := cStrtokenize(line, ':', 9)
	if len(p) < 6 || len(p) > 8 {
		return errors.New("invalid file type signature fields")
	}
	if len(p) > 6 {
		var maximum *string
		if len(p) == 8 {
			maximum = &p[7]
		}
		if err := d.fieldWindow(p[6], maximum); err != nil {
			return err
		}
	}
	rtype, knownRequired := ParseFileType(p[4])
	typ, knownDetected := ParseFileType(p[5])
	if !knownRequired || !knownDetected {
		return errors.New("unknown file type in file type signature")
	}
	switch kind := cAtoi(p[0]); kind {
	case 0, 4:
		if !isNumber(p[1]) {
			return errors.New("invalid file type magic offset")
		}
		magic, err := hex.DecodeString(p[2])
		if err != nil || len(magic) == 0 {
			return errors.New("invalid file type magic")
		}
		entry := ftmMagic{offset: int(cAtoi(p[1])), magic: magic, typ: typ}
		if kind == 0 {
			d.ftypes.magics = append([]ftmMagic{entry}, d.ftypes.magics...)
		} else {
			d.ftypes.partitions = append([]ftmMagic{entry}, d.ftypes.partitions...)
		}
	case 1:
		// Type patterns share the pattern table, so a raw scan's single
		// anchor pass finds them with the rules' patterns.
		ids, err := d.compilePatterns(0, []subsignature{{offset: p[1], body: p[2]}})
		if err != nil {
			return err
		}
		d.ftypes.patterns = append(d.ftypes.patterns, ftmPattern{id: ids[0], rtype: rtype, typ: typ})
	default:
		return skip(SkipFileTypeKind)
	}
	d.report.FileTypes++
	return nil
}

// compareMagic is the first typing step: the first magic entry whose bytes
// lie inside buf and equal it gives the type.
func (d *Database) compareMagic(buf []byte) (FileType, bool) {
	for _, m := range d.ftypes.magics {
		if m.offset+len(m.magic) <= len(buf) && string(buf[m.offset:m.offset+len(m.magic)]) == string(m.magic) {
			return m.typ, true
		}
	}
	return 0, false
}

// PartitionType applies the .ftm partition-magic table to a disk partition
// window. The reference scans an unmatched window as PART_ANY.
func (d *Database) PartitionType(buf []byte) FileType {
	for _, m := range d.ftypes.partitions {
		if m.offset >= 0 && m.offset <= len(buf) && len(m.magic) <= len(buf)-m.offset && string(buf[m.offset:m.offset+len(m.magic)]) == string(m.magic) {
			return m.typ
		}
	}
	return TypePartAny
}

// useBuiltInFileTypes adds the built-in table when no magic entry loaded, as
// the reference loads its internal table when its magic list is empty. The
// table is written here from public format magic numbers; it only matters
// for databases without a file-type database, such as local rules or
// hash-only feeds.
func (d *Database) useBuiltInFileTypes() {
	if len(d.ftypes.magics) > 0 {
		return
	}
	magics, patterns := builtInFileTypes()
	d.ftypes.magics = magics
	for _, p := range patterns {
		ids, err := d.compilePatterns(0, []subsignature{{offset: p.offset, body: p.body}})
		if err != nil {
			panic("UTMStack EDR built-in file type pattern: " + err.Error())
		}
		d.ftypes.patterns = append(d.ftypes.patterns, ftmPattern{id: ids[0], rtype: TypeAny, typ: p.typ})
	}
}

type builtInPattern struct {
	offset, body string
	typ          FileType
}

func builtInFileTypes() ([]ftmMagic, []builtInPattern) {
	magic := func(offset int, value string, typ FileType) ftmMagic {
		return ftmMagic{offset: offset, magic: []byte(value), typ: typ}
	}
	magics := []ftmMagic{
		magic(0, "MZ", TypeMSEXE),
		magic(0, "\xb6\xb9\xac\xae\xfe\xff\xff\xff", TypeCryptFF),
		magic(0, "\x7fELF", TypeELF),
		magic(0, "\xfe\xed\xfa\xce", TypeMachO), magic(0, "\xce\xfa\xed\xfe", TypeMachO),
		magic(0, "\xfe\xed\xfa\xcf", TypeMachO), magic(0, "\xcf\xfa\xed\xfe", TypeMachO),
		magic(0, "\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1", TypeMSOLE2),
		// MS-ONESTORE's file-header GUID identifies
		// OneNote sections even with a hash-only database. Without this,
		// attachment-only rules could silently miss the nested object.
		magic(0, "\xe4\x52\x5c\x7b\x8c\xd8\xa7\x4d\xae\xb1\x53\x78\xd0\x29\x96\xd3", TypeOneNote),
		magic(0, "PK\x03\x04", TypeZIP), magic(0, "PK\x05\x06", TypeZIP),
		magic(0, "%PDF-", TypePDF),
		magic(0, "FWS", TypeSWF), magic(0, "CWS", TypeSWF), magic(0, "ZWS", TypeSWF),
		magic(0, "\x89PNG", TypePNG),
		magic(0, "BM", TypeGraphics),
		magic(0, "GIF87a", TypeGIF), magic(0, "GIF89a", TypeGIF),
		magic(0, "\xff\xd8\xff", TypeJPEG),
		magic(0, "II*\x00", TypeTIFF), magic(0, "MM\x00*", TypeTIFF),
		magic(0, "\x1f\x8b", TypeGZ),
		magic(0, "BZh", TypeBZ),
		magic(0, "\xfd7zXZ\x00", TypeXZ),
		magic(0, "Rar!\x1a\x07", TypeRAR),
		magic(0, "7z\xbc\xaf\x27\x1c", Type7Z),
		magic(0, "MSCF", TypeMSCAB),
		magic(0, "ITSF", TypeMSCHM),
		magic(0, "SZDD\x88\xf0\x27\x33", TypeMSSZDD),
		magic(0, "\x78\x9f\x3e\x22", TypeTNEF),
		magic(0, "{\\rtf", TypeRTF),
		magic(0, "{\n  \"Magic\":\"CLAMJSONv0\"", TypeInternal),
		magic(0, "070707", TypeCPIOODC), magic(0, "070701", TypeCPIONewc), magic(0, "070702", TypeCPIOCRC),
		magic(0, "#@~^", TypeScrEnc),
		magic(0, "(This file must be converted with BinHex", TypeBinHex),
		magic(0, "Received: ", TypeMail), magic(0, "Return-Path: ", TypeMail), magic(0, "Delivered-To: ", TypeMail),
		magic(0, "MIME-Version: ", TypeMail), magic(0, "X-Mozilla-Status: ", TypeMail),
	}
	// Later entries are tried first, as for a loaded file.
	for i, j := 0, len(magics)-1; i < j; i, j = i+1, j-1 {
		magics[i], magics[j] = magics[j], magics[i]
	}
	return magics, []builtInPattern{
		{"0,1024", "4d5a{60-300}50450000", TypeMSEXE},
		{"0,1024", "255044462d3?2e", TypePDF},
		{"0", "4552{510}504d0000", TypeAPM},
		{"0", "efbbbf3c3f786d6c2076657273696f6e3d22312e3022*3c4857504d4c", TypeXMLHWP},
		// ClamAV's level-200 fallback recognizes the common compiled-Python
		// header family without requiring an external .ftm file.
		{"0", "??0d0d0a", TypePythonComp},
		{"0", "cafebabe0000000?", TypeMachOUnibin}, {"0", "cafebabe0000001?", TypeMachOUnibin},
		{"0", "cafebabe0000002?", TypeJava}, {"0", "cafebabe0000003?", TypeJava},
		// Match filetypes_int.h's fallback HTML set. The closing '>' on
		// html/head is significant: '<header' is not an HTML type marker.
		{"*", "3c4120*(68|48)(72|52)4546", TypeHTML},
		{"*", "3c4120*(68|48)(72|52)6566", TypeHTML},
		{"*", "3c484541443e", TypeHTML}, {"*", "3c48544d4c3e", TypeHTML},
		{"*", "3c486561643e", TypeHTML}, {"*", "3c48746d6c3e", TypeHTML},
		{"*", "3c494652414d45", TypeHTML}, {"*", "3c494d47", TypeHTML},
		{"*", "3c496d67", TypeHTML}, {"*", "3c4f424a454354", TypeHTML},
		{"*", "3c4f626a656374", TypeHTML}, {"*", "3c534352495054", TypeHTML},
		{"*", "3c536372697074", TypeHTML}, {"*", "3c5441424c45", TypeHTML},
		{"*", "3c6120*(68|48)(72|52)4546", TypeHTML},
		{"*", "3c6120*(68|48)(72|52)6566", TypeHTML},
		{"*", "3c686561643e", TypeHTML}, {"*", "3c68746d6c3e", TypeHTML},
		{"*", "3c696672616d65", TypeHTML}, {"*", "3c696d67", TypeHTML},
		{"*", "3c6f626a656374", TypeHTML}, {"*", "3c736372697074", TypeHTML},
		{"*", "3c7461626c65", TypeHTML},
		{"0", "3c2540204c414e4755414745203d", TypeHTML},
		// Archives inside other objects (the reference finds a 7-Zip archive
		// after other bytes with only a marker database loaded).
		{"*", "504b0304", TypeZIPSFX}, {"*", "526172211a0700", TypeRARSFX},
		{"*", "60ea{7}0002", TypeARJSFX}, {"*", "60ea{7}0102", TypeARJSFX}, {"*", "60ea{7}0202", TypeARJSFX},
		{"*", "4d53434600000000", TypeCABSFX}, {"*", "377abcaf271c", Type7ZSFX},
	}
}
