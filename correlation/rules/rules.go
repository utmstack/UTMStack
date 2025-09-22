package rules

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/correlation/utils"
)

func ListRulesFiles() []string {
	var files []string
	cnf := utils.GetConfig()
	catcher.Info("Listing rules files", map[string]any{"folder": cnf.RulesFolder})
	err := filepath.Walk(cnf.RulesFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			catcher.Error("Could not list rules files", err, nil)
		}

		if filepath.Ext(path) == ".yml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		catcher.Error("Could not list rules files", err, nil)
	}

	return files
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
	AllOf     []AllOf            `yaml:"allOf"`
	OneOf     []OneOf            `yaml:"oneOf"`
	TimeLapse int64              `yaml:"timeLapse"`
	MinCount  int                `yaml:"minCount"`
	Save      []utils.SavedField `yaml:"save"`
}

type Search struct {
	Query    string             `yaml:"query"`
	MinCount int                `yaml:"minCount"`
	Save     []utils.SavedField `yaml:"save"`
}

type Rule struct {
	Name        string   `yaml:"name"`
	Severity    string   `yaml:"severity"`
	Description string   `yaml:"description"`
	Solution    string   `yaml:"solution"`
	Category    string   `yaml:"category"`
	Tactic      string   `yaml:"tactic"`
	DataTypes   []string `yaml:"dataTypes"`
	Reference   []string `yaml:"reference"`
	Frequency   int64    `yaml:"frequency"`
	Cache       []Cache  `yaml:"cache"`
	Search      []Search `yaml:"search"`
}

func GetRules() []Rule {
	var tmpRules []Rule
	var rules []Rule

	for _, file := range ListRulesFiles() {
		catcher.Info("Reading rules from", map[string]any{"file": file})
		utils.ReadYaml(file, &tmpRules)
		catcher.Info("rule/s found", map[string]any{"count": len(tmpRules)})
		for _, tr := range tmpRules {
			n := true
			for _, r := range rules {
				if r.Name == tr.Name {
					n = false
					catcher.Info("Ignoring rule", map[string]any{"name": r.Name, "file": file})
					break
				}
			}
			if n {
				rules = append(rules, tr)
			}
		}
	}

	return rules
}

func Changes(signals chan os.Signal) {
	cnf := utils.GetConfig()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		catcher.Error("Could not create a new watcher", err, nil)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case err, ok := <-watcher.Errors:
				if !ok {
					catcher.Error("Could not detect changes in ruleset", err, nil)
				}
			case event, ok := <-watcher.Events:
				if !ok {
					catcher.Error("Error trying to detect changes in ruleset", err, nil)
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name != cnf.RulesFolder+"system/.git/FETCH_HEAD" {
						catcher.Info("Changes detected in", map[string]any{"file": event.Name})
						catcher.Info("Restarting correlation engine", nil)
						signals <- os.Interrupt
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
					catcher.Error("Could not list rules folders", err, nil)
				}
				n := true
				if info.IsDir() {
					for _, folder := range folders {
						if path == folder {
							n = false
							break
						}
					}
					if n {
						folders = append(folders, path)
						if err := watcher.Add(path); err != nil {
							catcher.Error("Could not start watcher for a rules folder", err, nil)
						}

					}
				}
				return nil
			})
			if err != nil {
				catcher.Error("Could not list rules folders", err, nil)
				continue
			}

			time.Sleep(10 * time.Second)
		}
	}()
	<-done
}
