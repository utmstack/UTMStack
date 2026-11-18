//go:build linux || darwin

package quarantine

func purgeRemoveFile(_ *purgeFile, d *restoreDirectory, name string) error {
	return d.removeFile(name)
}
