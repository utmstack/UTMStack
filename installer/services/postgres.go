package services

import (
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

func getPostgresContainerID() (string, error) {
	containerIDs, err := utils.RunCmdWithOutput("docker", "ps", "-q", "-f", "name=utmstack_postgres")
	if err != nil {
		return "", fmt.Errorf("error getting postgres container: %v", err)
	}
	if len(containerIDs) == 0 {
		return "", fmt.Errorf("postgres container not found")
	}
	return containerIDs[0], nil
}

func execPsql(containerID, database, query string) error {
	args := []string{"exec", containerID, "psql", "-U", "postgres"}
	if database != "" {
		args = append(args, "-d", database)
	}
	args = append(args, "-c", query)

	_, err := utils.RunCmdWithOutput("docker", args...)
	return err
}

func InitPgUtmstack(_ *config.Config) error {
	containerID, err := getPostgresContainerID()
	if err != nil {
		return err
	}

	// Creating utmstack database
	err = execPsql(containerID, "", "CREATE DATABASE utmstack")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	// Creating agentmanager database
	err = execPsql(containerID, "", "CREATE DATABASE agentmanager")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	return nil
}

func GetAdminEmail() (string, error) {
	// Get postgres container ID
	containerIDs, err := utils.RunCmdWithOutput("docker", "ps", "-q", "-f", "name=utmstack_postgres")
	if err != nil {
		return "", fmt.Errorf("error getting postgres container: %v", err)
	}

	if len(containerIDs) == 0 {
		return "", fmt.Errorf("postgres container not found")
	}

	containerID := containerIDs[0]

	// Execute query inside the container
	query := "SELECT email FROM jhi_user WHERE login = 'admin' AND created_by = 'system' AND email != 'admin@localhost' LIMIT 1"
	output, err := utils.RunCmdWithOutput("docker", "exec", containerID, "psql", "-U", "postgres", "-d", "utmstack", "-t", "-c", query)
	if err != nil {
		return "", fmt.Errorf("error executing query: %v", err)
	}

	if len(output) == 0 {
		return "", nil
	}

	return output[0], nil
}
