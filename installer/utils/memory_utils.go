package utils

import (
	"fmt"
	"sort"
)

const (
	SYSTEM_RESERVED_MEMORY = 1.4 * 1000
)

type ServiceConfig struct {
	Name           string
	Priority       int
	AssignedMemory int
	MinMemory      int
	MaxMemory      int
}

type serviceLevel struct {
	Priority int
	Children []*ServiceConfig
}

func (s *serviceLevel) balanceMemory(targetMemory int) (int, error) {
	totalMemory := targetMemory
	usedMemory := 0

	for _, child := range s.Children {
		if child.MinMemory > totalMemory {
			return usedMemory, fmt.Errorf("not enough memory to satisfy minimum for service: %s", child.Name)
		}
		child.AssignedMemory = child.MinMemory
		totalMemory -= child.MinMemory
		usedMemory += child.MinMemory
	}

	averageMemory := totalMemory / len(s.Children)

	for _, child := range s.Children {
		assignableMemory := min(averageMemory, child.MaxMemory-child.AssignedMemory)
		child.AssignedMemory += assignableMemory
		totalMemory -= assignableMemory
		usedMemory += assignableMemory
	}

	noLimitChildren := []*ServiceConfig{}
	for _, child := range s.Children {
		if child.MaxMemory == 0 {
			noLimitChildren = append(noLimitChildren, child)
		}
	}
	if len(noLimitChildren) > 0 {
		memoryPerChild := totalMemory / len(noLimitChildren)
		for _, child := range noLimitChildren {
			child.AssignedMemory += memoryPerChild
			totalMemory -= memoryPerChild
			usedMemory += memoryPerChild
		}
	}

	return usedMemory, nil
}

func (s *serviceLevel) getMinimum() int {
	totalMinMemory := 0
	for _, child := range s.Children {
		totalMinMemory += child.MinMemory
	}

	return totalMinMemory
}

func balanceMemoryAcrossTrees(trees []*serviceLevel, totalMemory int) error {
	totalMinMemory := 0
	for _, tree := range trees {
		totalMinMemory += tree.getMinimum()
	}
	if totalMemory < totalMinMemory {
		return fmt.Errorf("your system does not have the minimum required memory: %dMB", totalMinMemory+SYSTEM_RESERVED_MEMORY)
	}

	sort.Slice(trees, func(i, j int) bool {
		return trees[i].Priority < trees[j].Priority
	})

	targetPercentages := []float64{0.7, 0.2, 0.1}

	for i, tree := range trees {
		targetMemory := int(float64(totalMemory-totalMinMemory) * targetPercentages[i])
		if targetMemory > totalMemory {
			targetMemory = totalMemory
		}

		_, err := tree.balanceMemory(tree.getMinimum() + targetMemory)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}

	return nil
}

func createServiceLevels(services []ServiceConfig) []*serviceLevel {
	levelTrees := make(map[int]*serviceLevel)

	for _, sv := range services {
		sv := sv
		if _, ok := levelTrees[sv.Priority]; !ok {
			levelTrees[sv.Priority] = &serviceLevel{Priority: sv.Priority, Children: []*ServiceConfig{}}
		}
		levelTrees[sv.Priority].Children = append(levelTrees[sv.Priority].Children, &sv)
	}

	trees := []*serviceLevel{}
	for _, tree := range levelTrees {
		trees = append(trees, tree)
	}

	return trees
}

func BalanceMemory(services []ServiceConfig, totalMemory int) (map[string]*ServiceConfig, error) {
	trees := createServiceLevels(services)
	err := balanceMemoryAcrossTrees(trees, totalMemory)
	if err != nil {
		return nil, fmt.Errorf("error distributing memory: %v", err)
	}

	serviceMap := make(map[string]*ServiceConfig)
	for _, tree := range trees {
		for _, service := range tree.Children {
			serviceMap[service.Name] = service
		}
	}

	return serviceMap, nil
}

func GetOddValue(value int) int {
	if value%2 == 0 {
		return value
	}
	return value - 1
}
