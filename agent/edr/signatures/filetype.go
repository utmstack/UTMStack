package signatures

// FileType is an object type with the reference engine's numeric values.
// The order matters: typing compares these numbers (see typing.go).
type FileType uint16

const (
	TypeAny           FileType = 0
	TypeTextASCII     FileType = 500
	TypeTextUTF8      FileType = 501
	TypeTextUTF16LE   FileType = 502
	TypeTextUTF16BE   FileType = 503
	TypeBinaryData    FileType = 504
	TypeError         FileType = 505
	TypeMSEXE         FileType = 506
	TypeELF           FileType = 507
	TypeMachO         FileType = 508
	TypeMachOUnibin   FileType = 509
	TypePosixTar      FileType = 510
	TypeOldTar        FileType = 511
	TypeCPIOOld       FileType = 512
	TypeCPIOODC       FileType = 513
	TypeCPIONewc      FileType = 514
	TypeCPIOCRC       FileType = 515
	TypeGZ            FileType = 516
	TypeZIP           FileType = 517
	TypeBZ            FileType = 518
	TypeRAR           FileType = 519
	TypeARJ           FileType = 520
	TypeMSSZDD        FileType = 521
	TypeMSOLE2        FileType = 522
	TypeMSCAB         FileType = 523
	TypeMSCHM         FileType = 524
	TypeSIS           FileType = 525
	TypeScrEnc        FileType = 526
	TypeGraphics      FileType = 527
	TypeGIF           FileType = 528
	TypePNG           FileType = 529
	TypeJPEG          FileType = 530
	TypeTIFF          FileType = 531
	TypeRIFF          FileType = 532
	TypeBinHex        FileType = 533
	TypeTNEF          FileType = 534
	TypeCryptFF       FileType = 535
	TypePDF           FileType = 536
	TypeUUEncoded     FileType = 537
	TypeScript        FileType = 538
	TypeHTMLUTF16     FileType = 539
	TypeRTF           FileType = 540
	Type7Z            FileType = 541
	TypeSWF           FileType = 542
	TypeJava          FileType = 543
	TypeXAR           FileType = 544
	TypeXZ            FileType = 545
	TypeOOXMLWord     FileType = 546
	TypeOOXMLPPT      FileType = 547
	TypeOOXMLXL       FileType = 548
	TypeInternal      FileType = 549
	TypeHWP3          FileType = 550
	TypeOOXMLHWP      FileType = 551
	TypePS            FileType = 552
	TypeEGG           FileType = 553
	TypeOneNote       FileType = 554
	TypePythonComp    FileType = 555
	TypeLHALZH        FileType = 556
	TypeAIModel       FileType = 557
	TypePartAny       FileType = 558
	TypePartHFSPlus   FileType = 559
	TypeMBR           FileType = 560
	TypeHTML          FileType = 561
	TypeMail          FileType = 562
	TypeSFX           FileType = 563
	TypeZIPSFX        FileType = 564
	TypeRARSFX        FileType = 565
	Type7ZSFX         FileType = 566
	TypeCABSFX        FileType = 567
	TypeARJSFX        FileType = 568
	TypeEGGSFX        FileType = 569
	TypeNullsoft      FileType = 570
	TypeAutoIt        FileType = 571
	TypeIShieldMSI    FileType = 572
	TypeISO9660       FileType = 573
	TypeDMG           FileType = 574
	TypeGPT           FileType = 575
	TypeAPM           FileType = 576
	TypeXDP           FileType = 577
	TypeXMLWord       FileType = 578
	TypeXMLXL         FileType = 579
	TypeXMLHWP        FileType = 580
	TypeHWPOLE2       FileType = 581
	TypeMHTML         FileType = 582
	TypeLNK           FileType = 583
	TypeUDF           FileType = 584
	TypeALZ           FileType = 585
	TypeOther         FileType = 586
	TypeIgnored       FileType = 587
	firstNumberedType          = TypeTextASCII
)

// typeNames spells the types from TypeTextASCII on, in value order. ERROR
// and OTHER have report names but no rule names: the reference's name table
// has no entry for them.
var typeNames = [...]string{
	"TEXT_ASCII", "TEXT_UTF8", "TEXT_UTF16LE", "TEXT_UTF16BE", "BINARY_DATA", "ERROR", "MSEXE", "ELF", "MACHO",
	"MACHO_UNIBIN", "POSIX_TAR", "OLD_TAR", "CPIO_OLD", "CPIO_ODC", "CPIO_NEWC", "CPIO_CRC", "GZ", "ZIP", "BZ",
	"RAR", "ARJ", "MSSZDD", "MSOLE2", "MSCAB", "MSCHM", "SIS", "SCRENC", "GRAPHICS", "GIF", "PNG", "JPEG", "TIFF",
	"RIFF", "BINHEX", "TNEF", "CRYPTFF", "PDF", "UUENCODED", "SCRIPT", "HTML_UTF16", "RTF", "7Z", "SWF", "JAVA",
	"XAR", "XZ", "OOXML_WORD", "OOXML_PPT", "OOXML_XL", "INTERNAL", "HWP3", "OOXML_HWP", "PS", "EGG", "ONENOTE",
	"PYTHON_COMPILED", "LHA_LZH", "AI_MODEL", "PART_ANY", "PART_HFSPLUS", "MBR", "HTML", "MAIL", "SFX", "ZIPSFX", "RARSFX",
	"7ZSFX", "CABSFX", "ARJSFX", "EGGSFX", "NULSFT", "AUTOIT", "ISHIELD_MSI", "ISO9660", "DMG", "GPT", "APM", "XDP",
	"XML_WORD", "XML_XL", "XML_HWP", "HWPOLE2", "MHTML", "LNK", "UDF", "ALZ", "OTHER", "IGNORED",
}

// fileTypeNames are the names the reference accepts in rule attributes
// (Container, HandlerType, Intermediates) and .ftm/.cdb rules, compared
// exactly. An unknown name makes the reference skip a logical rule.
var fileTypeNames = func() map[string]FileType {
	names := map[string]FileType{"CL_TYPE_ANY": TypeAny}
	for i, n := range typeNames {
		if t := firstNumberedType + FileType(i); t != TypeError && t != TypeOther {
			names["CL_TYPE_"+n] = t
		}
	}
	return names
}()

// ParseFileType reads a rule's type name, spelled exactly as the reference
// spells it.
func ParseFileType(name string) (FileType, bool) {
	t, ok := fileTypeNames[name]
	return t, ok
}

func knownFileType(name string) bool {
	_, ok := fileTypeNames[name]
	return ok
}

// String is the type's name without the reference's prefix, for reports and
// events; values without a name give "".
func (t FileType) String() string {
	switch {
	case t == TypeAny:
		return "ANY"
	case t >= firstNumberedType && int(t-firstNumberedType) < len(typeNames):
		return typeNames[t-firstNumberedType]
	}
	return ""
}

// targetOf is the reference's target table: the typed rule root that objects
// of type t are matched against, besides root 0. Types without one give 0.
func targetOf(t FileType) int {
	switch t {
	case TypeMSEXE:
		return 1
	case TypeMSOLE2:
		return 2
	case TypeHTML:
		return 3
	case TypeMail:
		return 4
	case TypeGraphics, TypeGIF, TypePNG, TypeJPEG, TypeTIFF:
		return 5
	case TypeELF:
		return 6
	case TypeTextASCII:
		return 7
	case TypeError:
		return 8
	case TypeMachO, TypeMachOUnibin:
		return 9
	case TypePDF:
		return 10
	case TypeSWF:
		return 11
	case TypeJava:
		return 12
	case TypeInternal:
		return 13
	case TypeOther:
		return 14
	}
	return 0
}
