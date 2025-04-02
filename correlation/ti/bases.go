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
		"ip_blocklist.list",
	}

	for _, file := range files {
		var t string

		switch file {
		case "ip_blocklist.list":
			t = "IP"
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
