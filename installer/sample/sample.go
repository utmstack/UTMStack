package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/utmstack/UTMStack/installer/types"
	"github.com/utmstack/UTMStack/installer/utils"
)

func main() {
	// Total memory values in MB.
	totals := []int{16000, 20000, 32000, 48000, 64000, 120000, 200000}

	// Create a sorted copy of the services (sorted by Priority then Name).
	sortedServices := make([]utils.ServiceConfig, len(types.Services))
	copy(sortedServices, types.Services)
	sort.Slice(sortedServices, func(i, j int) bool {
		if sortedServices[i].Priority == sortedServices[j].Priority {
			return sortedServices[i].Name < sortedServices[j].Name
		}
		return sortedServices[i].Priority < sortedServices[j].Priority
	})

	// Create the output CSV file.
	file, err := os.Create("output.csv")
	if err != nil {
		log.Fatalf("Error creating CSV file: %v", err)
	}
	defer file.Close()

	// Create a CSV writer.
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Iterate over each total memory value.
	for _, total := range totals {
		// Write a comment line indicating the total memory.
		comment := fmt.Sprintf("# Total Memory: %d MB", total)
		// Since csv.Writer does not support comment lines,
		// write the comment directly to the file.
		if _, err := file.WriteString(comment + "\n"); err != nil {
			log.Fatalf("Error writing comment: %v", err)
		}

		// Write the header row.
		if err := writer.Write([]string{"Service", "AssignedMemory"}); err != nil {
			log.Fatalf("Error writing header: %v", err)
		}

		// Balance memory using the utils package.
		rsrcs, err := utils.BalanceMemory(types.Services, total)
		if err != nil {
			log.Fatalf("Error balancing memory for total %d MB: %v", total, err)
		}

		// Write each service's allocated memory.
		for _, svc := range sortedServices {
			// Assume that rsrcs[svc.Name] returns a pointer to a ServiceConfig,
			// and that AssignedMemory is an integer field.
			assigned := rsrcs[svc.Name].AssignedMemory
			row := []string{svc.Name, strconv.Itoa(assigned)}
			if err := writer.Write(row); err != nil {
				log.Fatalf("Error writing row: %v", err)
			}
		}

		// Flush the CSV writer so far.
		writer.Flush()
		if err := writer.Error(); err != nil {
			log.Fatalf("Error flushing CSV writer: %v", err)
		}

		// Write an empty line to separate tables.
		if _, err := file.WriteString("\n"); err != nil {
			log.Fatalf("Error writing empty line: %v", err)
		}
	}

	log.Println("CSV file 'output.csv' created successfully")
}
