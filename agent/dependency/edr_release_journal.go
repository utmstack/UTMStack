package dependency

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const edrJournalFile = ".edr-release-transaction.json"

type edrJournal struct {
	Schema       int      `json:"schema"`
	BackupDir    string   `json:"backup_dir"`
	StageDir     string   `json:"stage_dir"`
	DownloadDir  string   `json:"download_dir"`
	Present      []string `json:"present"`
	WasRunning   bool     `json:"was_running"`
	WasInstalled bool     `json:"was_installed"`
}

func (u *releaseEDRUpdate) journalPath() string { return filepath.Join(u.base, edrJournalFile) }

func (u *releaseEDRUpdate) writeJournal() error {
	path := u.journalPath()
	if _, err := os.Lstat(path); err == nil {
		return errors.New("previous EDR release transaction has not been recovered")
	} else if !os.IsNotExist(err) {
		return err
	}
	data, err := json.Marshal(edrJournal{Schema: 1, BackupDir: u.backupDir, StageDir: u.stageDir, DownloadDir: u.downloadDir, Present: u.presentNames, WasRunning: u.state.running, WasInstalled: u.state.installed})
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(u.base, ".edr-journal-")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func (u *releaseEDRUpdate) readJournal() (edrJournal, error) {
	var j edrJournal
	info, err := os.Lstat(u.journalPath())
	if err != nil {
		return j, err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > 8192 {
		return j, errors.New("EDR release transaction journal is not a bounded regular file")
	}
	f, err := os.Open(u.journalPath())
	if err != nil {
		return j, err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || opened.Size() != info.Size() {
		f.Close()
		return j, errors.New("EDR release journal changed before verification")
	}
	data, err := io.ReadAll(io.LimitReader(f, 8193))
	closeErr := f.Close()
	if err := errors.Join(err, closeErr); err != nil {
		return j, err
	}
	if len(data) > 8192 {
		return j, errors.New("EDR release journal exceeds bounds")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&j); err != nil {
		return j, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return j, errors.New("EDR release journal has trailing data")
	}
	if j.Schema != 1 || !privateEDRDir(u.base, j.BackupDir, ".edr-rollback-", true) || !privateEDRDir(u.base, j.StageDir, ".edr-release-", false) || j.DownloadDir != "" && !privateEDRDir(u.base, j.DownloadDir, ".edr-download-", false) {
		return j, errors.New("EDR release journal has invalid private paths")
	}
	allowed := make(map[string]bool, len(u.names))
	for _, name := range u.names {
		allowed[name] = true
	}
	seen := make(map[string]bool, len(j.Present))
	for _, name := range j.Present {
		if !allowed[name] || seen[name] {
			return j, errors.New("EDR release journal has invalid prior file list")
		}
		seen[name] = true
	}
	return j, nil
}

func privateEDRDir(base, dir, prefix string, required bool) bool {
	if !filepath.IsAbs(base) || !filepath.IsAbs(dir) || filepath.Dir(dir) != filepath.Clean(base) || !strings.HasPrefix(filepath.Base(dir), prefix) {
		return false
	}
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return !required
	}
	return err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

// recover rolls an interrupted transaction back to the exact pre-update file
// set. It never uses server bytes or executes a downloaded file.
func (u *releaseEDRUpdate) recover() error {
	j, err := u.readJournal()
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading EDR release transaction: %w", err)
	}
	state, err := u.status()
	if err != nil {
		return err
	}
	if state.running {
		if err := u.stop(); err != nil {
			return fmt.Errorf("stopping interrupted EDR release: %w", err)
		}
	}
	if !j.WasInstalled && state.installed {
		if err := u.uninstall(); err != nil {
			return fmt.Errorf("uninstalling interrupted new EDR service: %w", err)
		}
	} else if j.WasInstalled && !state.installed {
		return errors.New("previous EDR service registration disappeared during update")
	}
	present := make(map[string]bool, len(j.Present))
	for _, name := range j.Present {
		present[name] = true
	}
	for _, name := range u.names {
		backup := filepath.Join(j.BackupDir, name)
		target := filepath.Join(u.base, name)
		old, err := os.Lstat(backup)
		if err == nil {
			if !old.Mode().IsRegular() || !present[name] {
				return fmt.Errorf("EDR rollback file %q is inconsistent", name)
			}
			if err := removeRegularEDRFile(target); err != nil {
				return err
			}
			if err := os.Rename(backup, target); err != nil {
				return fmt.Errorf("restoring EDR file %q: %w", name, err)
			}
		} else if os.IsNotExist(err) {
			if present[name] {
				if info, err := os.Lstat(target); err != nil || !info.Mode().IsRegular() {
					return fmt.Errorf("prior EDR file %q missing from target and rollback", name)
				}
			} else if err := removeRegularEDRFile(target); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if j.WasRunning {
		if err := u.start(); err != nil {
			return fmt.Errorf("restarting prior EDR after recovery: %w", err)
		}
	}
	if err := os.Remove(u.journalPath()); err != nil {
		return err
	}
	for _, dir := range []string{j.BackupDir, j.StageDir, j.DownloadDir} {
		if dir != "" {
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
		}
	}
	u.backupDir, u.stageDir, u.downloadDir = "", "", ""
	u.backedUp, u.promotedNames, u.presentNames = nil, nil, nil
	u.promoted = false
	return nil
}

func removeRegularEDRFile(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("EDR transaction target %q is not a regular file", path)
	}
	return os.Remove(path)
}
