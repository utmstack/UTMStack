package utils

func Mode(mode string, options map[string]any) any {
	for k, v := range options {
		if k == mode {
			return v
		}
	}

	return nil
}
