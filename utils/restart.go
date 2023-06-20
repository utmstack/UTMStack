package utils

func Restart() error {
	if err := RunCmd("init", "6"); err != nil {
		return err
	}

	return nil
}
