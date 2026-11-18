package signatures

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

// addPasswordLine loads ClamAV's archive-password format. Only traditional
// ZIP encryption can use these passwords; other container selectors remain
// harmless data until a corresponding archive decryptor is available.
func (d *Database) addPasswordLine(line, _ string) error {
	parts := strings.SplitN(line, ";", 4)
	if len(parts) != 4 {
		return errors.New("invalid archive password fields")
	}
	name, description, storage, value := parts[0], parts[1], parts[2], parts[3]
	if err := validName(name); err != nil {
		return err
	}
	// Unlike other signature formats, the reference hashes only the password
	// name for an .ign/.ign2 MD5 entry.
	if d.ignore.ignored(name, name) {
		return skip(SkipIgnoreList)
	}
	if description == "" {
		return errors.New("archive password target description is empty")
	}
	t, err := parseTargetDescription(description)
	if err != nil {
		return err
	}
	if t.engine != nil && (t.engine.min > int64(d.functionality()) || t.engine.max < int64(d.functionality())) {
		return skip(SkipLevel)
	}
	if t.target != nil && *t.target >= maxTargets {
		return skip(SkipTarget)
	}
	if t.iconGroup && t.target != nil && *t.target != 1 {
		return errors.New("archive password icon group requires target 1")
	}
	if (t.entry != nil || t.sections != nil) && (t.target == nil || *t.target != 1 && *t.target != 6 && *t.target != 9) {
		return errors.New("archive password executable descriptor requires executable target")
	}
	if t.container != nil && *t.container != TypeAny && *t.container != TypeZIP && *t.container != TypeRAR {
		return errors.New("invalid archive password container")
	}
	var password []byte
	if !isNumber(storage) {
		return errors.New("invalid archive password storage type")
	}
	code, err := strconv.Atoi(storage)
	if err != nil {
		return errors.New("invalid archive password storage type")
	}
	switch code {
	case 0:
		password = []byte(value)
	case 1:
		password, err = hex.DecodeString(value)
		if err != nil {
			return errors.New("invalid hexadecimal archive password")
		}
	default:
		return skip(SkipPasswordStorage)
	}
	// A ZIP-specific password outranks a generic password even if the latter
	// appeared earlier in the source. A RAR-specific entry is selected by the
	// loader but ClamAV's ZIP decryptor never attempts it.
	group := 1
	if t.container != nil {
		switch *t.container {
		case TypeZIP:
			group = 0
		case TypeAny:
		default:
			return nil
		}
	}
	d.zipPasswords[group] = append(d.zipPasswords[group], password)
	return nil
}
