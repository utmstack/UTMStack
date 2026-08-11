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

// defaultTenantID is the platform tenant, whose administrator is the
// instance's. It is a fixed id: the backend seeds it and nothing renames it.
const defaultTenantID = "ce66672c-e36d-4761-a8c8-90058fee1a24"

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

	// The instance's administrator: the platform tenant's oldest active admin.
	// "admin@localhost" is the placeholder the backend creates when nobody
	// supplied an address, so it is not an answer.
	query := `SELECT u.email FROM "user" u ` +
		`JOIN user_role ur ON ur.user_id = u.id ` +
		`JOIN role r ON r.id = ur.role_id AND r.name = 'ROLE_ADMIN' ` +
		`WHERE u.tenant_id = '` + defaultTenantID + `' ` +
		`AND u.status = 'active' AND u.email <> 'admin@localhost' ` +
		`ORDER BY u.created_at LIMIT 1`
	output, err := utils.RunCmdWithOutput("docker", "exec", containerID, "psql", "-U", "postgres", "-d", "utmstack", "-t", "-c", query)
	if err != nil {
		return "", fmt.Errorf("error executing query: %v", err)
	}

	if len(output) == 0 {
		return "", nil
	}

	return output[0], nil
}
