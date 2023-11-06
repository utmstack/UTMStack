package rules

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/utmstack/UTMStack/correlation/utils"
)

func ListRulesFiles() []string {
	var files []string
	cnf := utils.GetConfig()
	log.Printf("Listing rules files in %s", cnf.RulesFolder)
	err := filepath.Walk(cnf.RulesFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Could not list rules files: %v", err)
		}

		if filepath.Ext(path) == ".yml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Could not list rules files: %v", err)
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
	TimeLapse int64          `yaml:"timeLapse"`
	MinCount  int          `yaml:"minCount"`
	Save      []SavedField `yaml:"save"`
}

type Search struct {
	Query    string       `yaml:"query"`
	MinCount int          `yaml:"minCount"`
	Save     []SavedField `yaml:"save"`
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
		log.Printf("Reading rules from: %s", file)
		utils.ReadYaml(file, &tmpRules)
		log.Printf("%v rule/s found", len(tmpRules))
		for _, tr := range tmpRules {
			new := true
			for _, r := range rules {
				if r.Name == tr.Name {
					new = false
					log.Printf("Ignoring rule: '%s' from: %s", r.Name, file)
					break
				}
			}
			if new {
				rules = append(rules, tr)
			}
		}
	}

	return rules
}

func RulesChanges(signals chan os.Signal) {
	cnf := utils.GetConfig()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Could not create a new watcher: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Printf("Could not detect changes in ruleset: %v", err)
				}
			case event, ok := <-watcher.Events:
				if !ok {
					log.Printf("Error trying to detect changes in ruleset.")
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name != cnf.RulesFolder+"system/.git/FETCH_HEAD" {
						log.Printf("Changes detected in: %s", event.Name)
						log.Printf("Restarting correlation engine")
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
					log.Printf("Could not list rules folders: %v", err)
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
							log.Printf("Could not start watcher for a rules folder: %v", err)
						}

					}
				}
				return nil
			})
			if err != nil {
				log.Printf("Could not list rules folders: %v", err)
				continue
			}

			time.Sleep(10 * time.Second)
		}
	}()
	<-done
}
