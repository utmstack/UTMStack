package ti

// import (
// 	"bufio"
// 	"os"
// 	"path/filepath"

// 	"github.com/threatwinds/go-sdk/catcher"
// )

// func Load() {
// 	catcher.Info("Loading Threat Intelligence Feeds", nil)

// 	var files = []string{
// 		"ip_level1.list",
// 		"ip_level2.list",
// 		"ip_level3.list",
// 	}

// 	for _, file := range files {
// 		var t string

// 		switch file {
// 		case "ip_level1.list":
// 			t = "Low"
// 		case "ip_level2.list":
// 			t = "Medium"
// 		case "ip_level3.list":
// 			t = "High"
// 		default:
// 		}

// 		f, err := os.Open(filepath.Join("/app", file))
// 		if err != nil {
// 			catcher.Error("Could not open file", err, nil)
// 			continue
// 		}

// 		scanner := bufio.NewScanner(f)

// 		for scanner.Scan() {
// 			element := scanner.Text()
// 			if element == "" {
// 				continue
// 			}

// 			blockList[element] = t
// 		}

// 		_ = f.Close()
// 	}

// 	catcher.Info("Threat Intelligence feeds loaded", nil)
// }
