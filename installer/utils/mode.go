package utils

func Mode(mode string, options map[string]interface{}) interface{}{
	for k,v := range options{
		if k == mode{
			return v
		}
	}

	return nil
}