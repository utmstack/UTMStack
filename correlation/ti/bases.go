package ti

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
)

func Load() {
	log.Printf("Loading Threat Intelligence Feeds")

	var files = []string{
		"ip_level1.list",
		"ip_level2.list",
		"ip_level3.list",
	}

	for _, file := range files {
		var t string

		switch file {
		case "ip_level1.list":
			t = "Low"
		case "ip_level2.list":
			t = "Medium"
		case "ip_level3.list":
			t = "High"
		default:
		}

		f, err := os.Open(filepath.Join("/app", file))
		if err != nil {
			log.Printf("Could not open file: %v", err)
			continue
		}

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			element := scanner.Text()
			if element == "" {
				continue
			}

			blockList[element] = t
		}

		_ = f.Close()
	}

	log.Printf("Threat Intelligence feeds loaded")
}
