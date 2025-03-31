package utils

import (
	"fmt"
	"math"
	"sort"
)

const (
	SYSTEM_RESERVED_MEMORY = 1 * 1000
)

type ServiceConfig struct {
	Name           string
	Priority       int
	AssignedMemory int
	MinMemory      int
	MaxMemory      int
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var priorityWeight = map[int]float64{
	1: 4.0,
	2: 3.0,
	3: 2.0,
}

func getWeight(priority int) float64 {
	if weight, ok := priorityWeight[priority]; ok {
		return weight
	}
	return 1.0
}

func BalanceMemory(services []ServiceConfig, totalSystemMemory int) (map[string]*ServiceConfig, error) {
	if totalSystemMemory <= SYSTEM_RESERVED_MEMORY {
		return nil, fmt.Errorf("total system memory (%dMB) is not greater than reserved memory (%dMB)", totalSystemMemory, SYSTEM_RESERVED_MEMORY)
	}
	availableMemory := totalSystemMemory - SYSTEM_RESERVED_MEMORY

	workingServices := make([]*ServiceConfig, len(services))
	for i := range services {
		s := services[i]
		workingServices[i] = &s
		workingServices[i].AssignedMemory = 0
	}

	totalMinMemoryNeeded := 0
	for _, s := range workingServices {
		if s.MinMemory < 0 {
			return nil, fmt.Errorf("service %s has negative MinMemory: %dMB", s.Name, s.MinMemory)
		}
		if s.MaxMemory != 0 && s.MinMemory > s.MaxMemory {
			return nil, fmt.Errorf("service %s has MinMemory (%dMB) greater than MaxMemory (%dMB)", s.Name, s.MinMemory, s.MaxMemory)
		}
		totalMinMemoryNeeded += s.MinMemory
	}

	if totalMinMemoryNeeded > availableMemory {
		return nil, fmt.Errorf("insufficient memory: Available %dMB < Total Minimum Required %dMB. (System requires %dMB)",
			availableMemory, totalMinMemoryNeeded, totalMinMemoryNeeded+SYSTEM_RESERVED_MEMORY)
	}

	memoryUsed := 0
	for _, s := range workingServices {
		s.AssignedMemory = s.MinMemory
		memoryUsed += s.MinMemory
	}
	remainingMemory := availableMemory - memoryUsed

	for remainingMemory > 0 {
		distributableInThisPass := remainingMemory
		candidates := []*ServiceConfig{}
		totalWeight := 0.0

		for _, s := range workingServices {
			canTakeMore := (s.MaxMemory == 0) || (s.AssignedMemory < s.MaxMemory)
			if canTakeMore {
				candidates = append(candidates, s)
				totalWeight += getWeight(s.Priority)
			}
		}

		if len(candidates) == 0 || totalWeight <= 0 {
			break
		}

		memoryAssignedInPass := 0
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].Priority < candidates[j].Priority
		})

		tempAssignments := make(map[string]int)
		for _, s := range candidates {
			weight := getWeight(s.Priority)
			proportionalShare := float64(distributableInThisPass) * (weight / totalWeight)

			remainingCapacity := math.MaxInt
			if s.MaxMemory != 0 {
				remainingCapacity = s.MaxMemory - s.AssignedMemory
			}

			assignAmount := int(math.Floor(proportionalShare))
			assignAmount = minInt(assignAmount, remainingCapacity)
			assignAmount = maxInt(0, assignAmount)

			tempAssignments[s.Name] = assignAmount
			memoryAssignedInPass += assignAmount
		}

		for _, s := range candidates {
			s.AssignedMemory += tempAssignments[s.Name]
		}

		remainingMemory -= memoryAssignedInPass

		if memoryAssignedInPass == 0 && remainingMemory > 0 {
			break
		}
	}

	if remainingMemory > 0 {
		sort.SliceStable(workingServices, func(i, j int) bool {
			if workingServices[i].Priority != workingServices[j].Priority {
				return workingServices[i].Priority < workingServices[j].Priority
			}
			if workingServices[i].MaxMemory == 0 && workingServices[j].MaxMemory != 0 {
				return true
			}
			if workingServices[i].MaxMemory != 0 && workingServices[j].MaxMemory == 0 {
				return false
			}
			return workingServices[i].Name < workingServices[j].Name
		})

		for remainingMemory > 0 {
			memoryDistributedInSweep := false
			for _, s := range workingServices {
				if remainingMemory == 0 {
					break
				}
				canTakeMore := (s.MaxMemory == 0) || (s.AssignedMemory < s.MaxMemory)
				if canTakeMore {
					s.AssignedMemory++
					remainingMemory--
					memoryDistributedInSweep = true
				}
			}
			if !memoryDistributedInSweep && remainingMemory > 0 {
				fmt.Printf("WARNING: Unable to distribute final %dMB in sweep phase. Memory might be underutilized.\n", remainingMemory)
				break
			}
		}
	}

	finalAssignedMemory := 0
	serviceMap := make(map[string]*ServiceConfig)
	for _, s := range workingServices {
		serviceMap[s.Name] = s
		finalAssignedMemory += s.AssignedMemory
	}

	if finalAssignedMemory != availableMemory {
		fmt.Printf("WARNING: Final memory validation failed. Expected %dMB, Assigned %dMB. Discrepancy: %dMB\n",
			availableMemory, finalAssignedMemory, availableMemory-finalAssignedMemory)
	} else {
		fmt.Printf("Memory distribution successful. Total Assigned: %dMB (Available: %dMB)\n", finalAssignedMemory, availableMemory)
	}

	return serviceMap, nil
}

func GetOddValue(value int) int {
	if value%2 == 0 {
		return value
	}
	return value - 1
}
