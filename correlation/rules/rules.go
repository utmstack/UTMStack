package rules

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/utmstack/UTMStack/correlation/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/quantfall/holmes"
)

var mu = &sync.Mutex{}
var h = holmes.New(utils.GetConfig().ErrorLevel, "RULES")

func ListRulesFiles() []string {
	var files []string
	cnf := utils.GetConfig()
	h.Info("Listing rules files in %s", cnf.RulesFolder)
	err := filepath.Walk(cnf.RulesFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			h.Error("Could not list rules files: %v", err)
		}

		if filepath.Ext(path) == ".yml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		h.Error("Could not list rules files: %v", err)
	}

	return files
}

type SavedField struct {
	Field string `yaml:"field"`
	Alias string `yaml:"alias"`
}

type AllOf struct {
	Field    string `yaml:"field"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
}

type OneOf struct {
	Field    string `yaml:"field"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
}

type Cache struct {
	AllOf     []AllOf      `yaml:"allOf"`
	OneOf     []OneOf      `yaml:"oneOf"`
	TimeLapse int          `yaml:"timeLapse"`
	MinCount  int          `yaml:"minCount"`
	Save      []SavedField `yaml:"save"`
}

type Search struct {
	Query    string       `yaml:"query"`
	MinCount int          `yaml:"minCount"`
	Save     []SavedField `yaml:"save"`
}

type Rule struct {
	Name        string        `yaml:"name"`
	Severity    string        `yaml:"severity"`
	Description string        `yaml:"description"`
	Solution    string        `yaml:"solution"`
	Category    string        `yaml:"category"`
	Tactic      string        `yaml:"tactic"`
	Reference   []string      `yaml:"reference"`
	Frequency   time.Duration `yaml:"frequency"`
	Cache       []Cache       `yaml:"cache"`
	Search      []Search      `yaml:"search"`
}

func GetRules() []Rule {
	var tmpRules []Rule
	var rules []Rule

	for _, file := range ListRulesFiles() {
		h.Debug("Reading rules from: %s", file)
		utils.ReadYaml(file, &tmpRules)
		h.Debug("%v rule/s found", len(tmpRules))
		for _, tr := range tmpRules {
			new := true
			for _, r := range rules {
				if r.Name == tr.Name {
					new = false
					h.Info("Ignoring rule: '%s' from: %s", r.Name, file)
					break
				}
			}
			if new {
				rules = append(rules, tr)
			}
		}

	}
	h.Debug("Rules to load: %v", len(rules))
	return rules
}

func RulesChanges(signals chan os.Signal) {
	cnf := utils.GetConfig()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		h.Error("Could not create a new watcher: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case err, ok := <-watcher.Errors:
				if !ok {
					h.Error("Could not detect changes in ruleset: %v", err)
				}
			case event, ok := <-watcher.Events:
				if !ok {
					h.Error("Error trying to detect changes in ruleset.")
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name != cnf.RulesFolder+"system/.git/FETCH_HEAD" {
						h.Debug("Changes detected in: %s", event.Name)
						mu.Lock()
						h.Info("Restarting correlation engine")
						signals <- syscall.SIGTERM
						mu.Unlock()
					}
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		var folders []string
		for {
			err := filepath.Walk(cnf.RulesFolder, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					h.Error("Could not list rules folders: %v", err)
				}
				new := true
				if info.IsDir() {
					for _, folder := range folders {
						if path == folder {
							new = false
							break
						}
					}
					if new {
						folders = append(folders, path)
						if err := watcher.Add(path); err != nil {
							h.Error("Could not start watcher for a rules folder: %v", err)
						}

					}
				}
				return nil
			})
			if err != nil {
				h.Error("Could not list rules folders: %v", err)
				continue
			}

			time.Sleep(10 * time.Second)
		}
	}()
	<-done
}
