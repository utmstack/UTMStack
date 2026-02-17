package dependency

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/http"
)

const (
	updaterBaseName = "utmstack_updater_service"

	// Dependency versions - single source of truth
	UpdaterVersion        = "11.2.3"
	BeatsVersion          = "11.2.3"
	MacosCollectorVersion = "11.2.3"
)

// UpdaterFile returns the updater binary name with the appropriate extension for the current OS.
func UpdaterFile(suffix string) string {
	if runtime.GOOS == "windows" {
		return updaterBaseName + suffix + ".exe"
	}
	return updaterBaseName + suffix
}

// Dependency represents a dependency that the agent needs.
type Dependency struct {
	Name         string                       // Unique identifier
	Version      string                       // Current version in this agent build
	BinaryPath   string                       // Path to check if already exists
	DownloadURL  func(server string) string   // URL template to download from
	DownloadName string                       // Filename to save as (if different from BinaryPath basename)
	Critical     bool                         // If true, failure blocks agent startup
	PostDownload func() error                 // Run after download (e.g., unzip). Can be nil.
	Configure    func() error                 // Run on first install (can be nil)
	Update       func() error                 // Run on version change (can be nil, uses Configure)
	Uninstall    func() error                 // Run when dependency is removed (can be nil)
}

// Exists checks if the dependency binary exists on disk.
func (d *Dependency) Exists() bool {
	return fs.Exists(d.BinaryPath)
}

// InstalledDep represents a dependency that has been installed and tracked.
type InstalledDep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InstalledDeps is the list of installed dependencies (persisted to JSON).
type InstalledDeps struct {
	Dependencies []InstalledDep `json:"dependencies"`
}

func (i *InstalledDeps) Get(name string) *InstalledDep {
	for idx := range i.Dependencies {
		if i.Dependencies[idx].Name == name {
			return &i.Dependencies[idx]
		}
	}
	return nil
}

func (i *InstalledDeps) Add(name, version string) {
	i.Dependencies = append(i.Dependencies, InstalledDep{Name: name, Version: version})
}

func (i *InstalledDeps) Update(name, version string) {
	for idx := range i.Dependencies {
		if i.Dependencies[idx].Name == name {
			i.Dependencies[idx].Version = version
			return
		}
	}
}

func (i *InstalledDeps) Remove(name string) {
	for idx := range i.Dependencies {
		if i.Dependencies[idx].Name == name {
			i.Dependencies = append(i.Dependencies[:idx], i.Dependencies[idx+1:]...)
			return
		}
	}
}

func (i *InstalledDeps) Has(name string) bool {
	return i.Get(name) != nil
}

var (
	installedDepsFile = filepath.Join(fs.GetExecutablePath(), "dependencies.json")
	reconcileMu       sync.Mutex
)

// readInstalledDeps reads the installed dependencies from disk.
func readInstalledDeps() (*InstalledDeps, error) {
	installed := &InstalledDeps{Dependencies: []InstalledDep{}}
	if !fs.Exists(installedDepsFile) {
		return installed, nil
	}
	if err := fs.ReadJSON(installedDepsFile, installed); err != nil {
		return nil, fmt.Errorf("error reading dependencies file: %v", err)
	}
	return installed, nil
}

// writeInstalledDeps writes the installed dependencies to disk.
func writeInstalledDeps(installed *InstalledDeps) error {
	return fs.WriteJSON(installedDepsFile, installed)
}

// Reconcile ensures all dependencies are installed and up-to-date.
// This should be called at agent startup, before starting collectors.
func Reconcile(server string, skipCertValidation bool) error {
	reconcileMu.Lock()
	defer reconcileMu.Unlock()

	utils.Logger.Info("Starting dependency reconciliation...")

	installed, err := readInstalledDeps()
	if err != nil {
		return fmt.Errorf("failed to read installed dependencies: %v", err)
	}

	desired := GetDependencies()
	var criticalErrors []error

	// Process each desired dependency
	for _, dep := range desired {
		inst := installed.Get(dep.Name)

		if inst == nil {
			// Not tracked yet
			if dep.Exists() {
				// MIGRATION: Already installed by previous version, just track it
				utils.Logger.Info("Migrating existing dependency: %s (version %s)", dep.Name, dep.Version)
				installed.Add(dep.Name, dep.Version)
			} else {
				// FRESH INSTALL: Download and configure
				utils.Logger.Info("Installing new dependency: %s (version %s)", dep.Name, dep.Version)
				if err := downloadDependency(dep, server, skipCertValidation); err != nil {
					errMsg := fmt.Errorf("failed to download dependency %s: %v", dep.Name, err)
					utils.Logger.ErrorF("%v", errMsg)
					if dep.Critical {
						criticalErrors = append(criticalErrors, errMsg)
					}
					continue
				}

				if dep.Configure != nil {
					if err := dep.Configure(); err != nil {
						errMsg := fmt.Errorf("failed to configure dependency %s: %v", dep.Name, err)
						utils.Logger.ErrorF("%v", errMsg)
						if dep.Critical {
							criticalErrors = append(criticalErrors, errMsg)
						}
						continue
					}
				}
				installed.Add(dep.Name, dep.Version)
				utils.Logger.Info("Dependency %s installed successfully", dep.Name)
			}
		} else if inst.Version != dep.Version {
			// VERSION CHANGED: Download and update
			utils.Logger.Info("Updating dependency: %s (%s -> %s)", dep.Name, inst.Version, dep.Version)
			if err := downloadDependency(dep, server, skipCertValidation); err != nil {
				errMsg := fmt.Errorf("failed to download dependency update %s: %v", dep.Name, err)
				utils.Logger.ErrorF("%v", errMsg)
				if dep.Critical {
					criticalErrors = append(criticalErrors, errMsg)
				}
				continue
			}

			updateFn := dep.Update
			if updateFn == nil {
				updateFn = dep.Configure
			}
			if updateFn != nil {
				if err := updateFn(); err != nil {
					errMsg := fmt.Errorf("failed to update dependency %s: %v", dep.Name, err)
					utils.Logger.ErrorF("%v", errMsg)
					if dep.Critical {
						criticalErrors = append(criticalErrors, errMsg)
					}
					continue
				}
			}
			installed.Update(dep.Name, dep.Version)
			utils.Logger.Info("Dependency %s updated successfully", dep.Name)
		} else {
			// Same version, nothing to do
			utils.Logger.Info("Dependency %s is up to date (version %s)", dep.Name, dep.Version)
		}
	}

	// CLEANUP: Remove dependencies that are no longer needed
	for _, inst := range installed.Dependencies {
		found := false
		for _, dep := range desired {
			if dep.Name == inst.Name {
				found = true
				break
			}
		}
		if !found {
			utils.Logger.Info("Removing obsolete dependency: %s", inst.Name)
			// Find the uninstall function if we have it from old code
			// For now, just remove from tracking
			installed.Remove(inst.Name)
		}
	}

	// Save the updated installed dependencies
	if err := writeInstalledDeps(installed); err != nil {
		return fmt.Errorf("failed to write installed dependencies: %v", err)
	}

	if len(criticalErrors) > 0 {
		return fmt.Errorf("critical dependency errors: %v", criticalErrors)
	}

	utils.Logger.Info("Dependency reconciliation completed")
	return nil
}

// downloadDependency downloads the dependency binary.
func downloadDependency(dep Dependency, server string, skipCertValidation bool) error {
	if dep.DownloadURL == nil {
		return fmt.Errorf("dependency %s has no download URL", dep.Name)
	}

	url := dep.DownloadURL(server)

	// Use DownloadName if specified, otherwise use basename of BinaryPath
	filename := dep.DownloadName
	if filename == "" {
		filename = filepath.Base(dep.BinaryPath)
	}

	destDir := filepath.Dir(dep.BinaryPath)

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", destDir, err)
	}

	if err := http.DownloadFile(url, map[string]string{}, filename, destDir, skipCertValidation); err != nil {
		return err
	}

	// Run post-download hook if defined (e.g., unzip)
	if dep.PostDownload != nil {
		if err := dep.PostDownload(); err != nil {
			return fmt.Errorf("post-download failed: %v", err)
		}
	}

	return nil
}
