package netblock

import "strings"

func stripQ(p string) string { return strings.TrimPrefix(p, `\??\`) }
