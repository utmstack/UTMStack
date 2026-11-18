package dependency

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/release"
	"github.com/utmstack/UTMStack/shared/svc"
)

// EDRReleasePublicKeyHex is set only by the trusted agent build. An empty or
// malformed value disables native EDR updates; server data cannot set it.
var EDRReleasePublicKeyHex string

type edrServiceState struct{ installed, running bool }

type releaseServiceProbe struct{}

func (releaseServiceProbe) Start(service.Service) error { return nil }
func (releaseServiceProbe) Stop(service.Service) error  { return nil }

type releaseEDRUpdate struct {
	base, archive, downloadDir, backupDir, stageDir string
	names, presentNames, backedUp, promotedNames    []string
	identity                                        release.Identity
	expectedServiceSHA                              string
	state                                           edrServiceState
	promoted                                        bool
	key                                             func() (ed25519.PublicKey, error)
	url                                             func(string) string
	client                                          func(bool) *http.Client
	version                                         func(string) error
	ready                                           func(time.Time) error
	status                                          func() (edrServiceState, error)
	stop, start, install, uninstall                 func() error
}

func nativeEDRDependency(base string) Dependency {
	u := newReleaseEDRUpdate(base)
	return Dependency{
		Name: "edr", Version: EDRVersion,
		BinaryPath:   filepath.Join(base, u.identity.ServiceName),
		DownloadName: filepath.Base(u.archive),
		Download:     u.download,
		Critical:     false,
		Recover:      u.recover,
		CheckExisting: func() bool {
			return u.verifyInstalled() == nil
		},
		PostDownload: u.promote,
		Configure:    u.finish,
		Update:       u.finish,
		Uninstall:    u.uninstall,
	}
}

func edrReleasePublicKey() (ed25519.PublicKey, error) {
	key, err := hex.DecodeString(EDRReleasePublicKeyHex)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("EDR release publisher public key is not pinned in this agent build")
	}
	nonzero := false
	for _, b := range key {
		nonzero = nonzero || b != 0
	}
	if !nonzero {
		return nil, errors.New("EDR release publisher public key is all zero")
	}
	return ed25519.PublicKey(key), nil
}

func newReleaseEDRUpdate(base string) *releaseEDRUpdate {
	identity := release.Identity{
		PackageVersion: EDRVersion, Implementation: buildinfo.ImplementationVersion,
		OS: runtime.GOOS, Architecture: runtime.GOARCH,
		ServiceName: EDRFile(""),
	}
	if runtime.GOOS == "windows" {
		identity.ProviderName = "utmstack_amsi.dll"
	}
	u := &releaseEDRUpdate{
		base: base, archive: filepath.Join(base, identity.ServiceName+".release.incoming.zip"),
		identity: identity, key: edrReleasePublicKey, version: verifyNativeEDR,
		url:     edrReleaseDownloadURL,
		client:  newEDRReleaseClient,
		status:  edrServiceStatus,
		ready:   nil,
		stop:    func() error { return stopEDRReleaseService(filepath.Join(base, identity.ServiceName)) },
		start:   func() error { return svc.Start(edrconfig.ServiceName) },
		install: configureEDR,
		uninstall: func() error {
			s, err := service.New(releaseServiceProbe{}, &service.Config{Name: edrconfig.ServiceName})
			if err != nil {
				return err
			}
			return s.Uninstall()
		},
	}
	u.names = []string{identity.ServiceName}
	if identity.ProviderName != "" {
		u.names = append(u.names, identity.ProviderName)
	}
	u.names = append(u.names, identity.ServiceName+".release.json")
	return u
}

func (u *releaseEDRUpdate) verifyInstalled() error {
	u.expectedServiceSHA = ""
	key, err := u.key()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(u.base, u.identity.ServiceName+".release.json")
	info, err := os.Lstat(manifestPath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > 16<<10 {
		return errors.New("installed EDR release manifest is not a bounded regular file")
	}
	f, err := os.Open(manifestPath)
	if err != nil {
		return err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || opened.Size() != info.Size() {
		f.Close()
		return errors.New("installed EDR release manifest changed before verification")
	}
	data, err := io.ReadAll(io.LimitReader(f, (16<<10)+1))
	closeErr := f.Close()
	if err := errors.Join(err, closeErr); err != nil {
		return err
	}
	manifest, err := release.Verify(data, key, u.identity)
	if err != nil {
		return err
	}
	if err := release.VerifyFiles(u.base, manifest); err != nil {
		return err
	}
	for _, artifact := range manifest.Artifacts {
		if artifact.Name == u.identity.ServiceName {
			u.expectedServiceSHA = artifact.SHA256
			return nil
		}
	}
	return errors.New("signed EDR release has no service executable")
}

// promote authenticates the full archive before stopping any protection. The
// version command runs only after the signed manifest and exact bytes pass.
func (u *releaseEDRUpdate) promote() (err error) {
	if _, err := os.Lstat(u.journalPath()); err == nil {
		return errors.New("previous EDR release transaction must be recovered before promotion")
	} else if !os.IsNotExist(err) {
		return err
	}
	u.presentNames = nil
	u.expectedServiceSHA = ""
	defer func() {
		if _, journalErr := os.Lstat(u.journalPath()); journalErr == nil {
			return // retain every recovery input until finish or a safe rollback
		}
		_ = os.Remove(u.archive)
		if u.downloadDir != "" {
			_ = os.RemoveAll(u.downloadDir)
			u.downloadDir = ""
		}
		if err != nil && u.stageDir != "" {
			_ = os.RemoveAll(u.stageDir)
			u.stageDir = ""
		}
	}()
	key, err := u.key()
	if err != nil {
		return err
	}
	var manifest release.Manifest
	u.stageDir, manifest, err = release.Extract(u.archive, u.base, key, u.identity)
	if err != nil {
		return fmt.Errorf("rejecting native EDR release: %w", err)
	}
	for _, artifact := range manifest.Artifacts {
		if artifact.Name == u.identity.ServiceName {
			u.expectedServiceSHA = artifact.SHA256
		}
	}
	service := filepath.Join(u.stageDir, u.identity.ServiceName)
	if err := prepareEDRExecutable(service); err != nil {
		return err
	}
	if err := u.version(service); err != nil {
		return fmt.Errorf("authenticated native EDR version check: %w", err)
	}
	for _, name := range u.names {
		path := filepath.Join(u.base, name)
		if old, statErr := os.Lstat(path); statErr == nil {
			if !old.Mode().IsRegular() {
				return fmt.Errorf("existing EDR file %q is not regular", name)
			}
			u.presentNames = append(u.presentNames, name)
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	}
	u.state, err = u.status()
	if err != nil {
		return fmt.Errorf("checking EDR service before replacement: %w", err)
	}
	u.backupDir, err = release.MakePrivateTempDir(u.base, ".edr-rollback-")
	if err != nil {
		return err
	}
	if err := u.writeJournal(); err != nil {
		_ = os.Remove(u.backupDir)
		u.backupDir = ""
		return err
	}
	if u.state.running {
		if err := u.stop(); err != nil {
			return errors.Join(fmt.Errorf("stopping EDR for replacement: %w", err), u.rollback())
		}
	}
	for _, name := range u.names {
		from := filepath.Join(u.base, name)
		to := filepath.Join(u.backupDir, name)
		if err := os.Rename(from, to); err == nil {
			u.backedUp = append(u.backedUp, name)
		} else if !os.IsNotExist(err) {
			return errors.Join(fmt.Errorf("preserving EDR file %q: %w", name, err), u.rollback())
		}
	}
	for _, name := range u.names {
		stagedName := name
		if name == u.identity.ServiceName+".release.json" {
			stagedName = "release.json"
		}
		from := filepath.Join(u.stageDir, stagedName)
		to := filepath.Join(u.base, name)
		if err := os.Rename(from, to); err != nil {
			return errors.Join(fmt.Errorf("promoting EDR file %q: %w", name, err), u.rollback())
		}
		u.promotedNames = append(u.promotedNames, name)
	}
	u.promoted = true
	return nil
}

func (u *releaseEDRUpdate) finish() error {
	if err := u.verifyInstalled(); err != nil {
		return errors.Join(fmt.Errorf("installed EDR release verification: %w", err), u.rollback())
	}
	for _, name := range u.names[:len(u.names)-1] {
		if err := release.PreparePublishedFile(filepath.Join(u.base, name)); err != nil {
			return errors.Join(fmt.Errorf("published EDR access: %w", err), u.rollback())
		}
	}
	if err := u.verifyInstalled(); err != nil {
		return errors.Join(err, u.rollback())
	}
	if !u.promoted {
		state, err := u.status()
		if err != nil {
			return err
		}
		u.state = state
	}
	installedNow := false
	if !u.state.installed {
		if err := u.install(); err != nil {
			return errors.Join(fmt.Errorf("installing native EDR service: %w", err), u.rollback())
		}
		state, err := u.status()
		if err != nil || !state.installed {
			return errors.Join(errors.New("replacement EDR service did not report installed"), err, u.rollback())
		}
		installedNow = true
	}
	abort := func(err error) error {
		if u.promoted {
			return errors.Join(err, u.rollback())
		}
		if installedNow {
			return errors.Join(err, u.stop(), u.uninstall())
		}
		return err
	}
	if u.promoted && u.state.running || installedNow {
		startAttempt := time.Now().UTC()
		if err := u.start(); err != nil {
			return abort(fmt.Errorf("starting replacement EDR service: %w", err))
		}
		state, err := u.status()
		if err != nil || !state.running {
			return abort(errors.Join(errors.New("replacement EDR service did not report running"), err))
		}
		ready := u.ready
		if ready == nil {
			ready = u.waitForReady
		}
		if err := ready(startAttempt); err != nil {
			return abort(fmt.Errorf("replacement EDR service did not become ready: %w", err))
		}
		state, err = u.status()
		if err != nil || !state.running {
			return abort(errors.Join(errors.New("replacement EDR service stopped before commit"), err))
		}
	}
	if u.promoted {
		if err := os.Remove(u.journalPath()); err != nil {
			return fmt.Errorf("committing native EDR release journal: %w", err)
		}
	}
	u.promoted = false
	if u.backupDir != "" {
		if err := os.RemoveAll(u.backupDir); err != nil {
			return fmt.Errorf("removing native EDR rollback files: %w", err)
		}
		u.backupDir = ""
	}
	if u.stageDir != "" {
		_ = os.RemoveAll(u.stageDir)
		u.stageDir = ""
	}
	if u.downloadDir != "" {
		_ = os.RemoveAll(u.downloadDir)
		u.downloadDir = ""
	}
	u.backedUp, u.promotedNames, u.presentNames = nil, nil, nil
	return nil
}

type edrRuntimeStatus struct {
	InstanceID           string `json:"instance_id"`
	ProcessID            int    `json:"process_id"`
	StartedAt            string `json:"started_at"`
	Running              bool   `json:"running"`
	EngineHealthy        bool   `json:"engine_healthy"`
	EngineImplementation string `json:"engine_implementation"`
	ServiceSHA256        string `json:"service_executable_sha256"`
	SensorAMSI           bool   `json:"sensor_amsi"`
	AMSILocalReady       bool   `json:"amsi_local_ready"`
}

const edrReadyTimeout = 90 * time.Second

func (u *releaseEDRUpdate) waitForReady(startAttempt time.Time) error {
	deadline := time.NewTimer(edrReadyTimeout)
	defer deadline.Stop()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		err := u.checkReady(startAttempt)
		if err == nil {
			return nil
		}
		select {
		case <-deadline.C:
			return err
		case <-tick.C:
		}
	}
}

func (u *releaseEDRUpdate) checkReady(startAttempt time.Time) error {
	path := filepath.Join(u.base, "status.json")
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > 1<<20 {
		return errors.New("EDR status is not a bounded regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || opened.Size() != info.Size() {
		f.Close()
		return errors.New("EDR status changed during opening")
	}
	data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	closeErr := f.Close()
	if err := errors.Join(err, closeErr); err != nil {
		return err
	}
	if len(data) > 1<<20 {
		return errors.New("EDR status exceeds size bound")
	}
	var status edrRuntimeStatus
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&status); err != nil {
		return fmt.Errorf("decoding EDR status: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return errors.New("EDR status has trailing data")
	}
	started, err := time.Parse(time.RFC3339Nano, status.StartedAt)
	if err != nil || status.InstanceID == "" || status.ProcessID <= 0 || started.Before(startAttempt) {
		return errors.New("EDR status is stale or has no process identity")
	}
	if !status.Running || !status.EngineHealthy || status.EngineImplementation != u.identity.Implementation || u.expectedServiceSHA == "" || status.ServiceSHA256 != u.expectedServiceSHA {
		return errors.New("EDR status does not report a running healthy release")
	}
	if status.SensorAMSI && !status.AMSILocalReady {
		return errors.New("EDR AMSI registration or local pipe is not ready")
	}
	return checkEDRPlatformReadiness(u.base, data)
}

// rollback never removes a prior file unless that exact file is already in
// the private backup directory. A failed restore retains the backup for repair.
func (u *releaseEDRUpdate) rollback() error {
	return u.recover()
}
