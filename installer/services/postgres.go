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

	// Creating utm_client table
	createTable := `CREATE TABLE public.utm_client (
		id serial NOT NULL,
		client_name varchar(100) NULL,
		client_domain varchar(100) NULL,
		client_prefix varchar(10) NULL,
		client_mail varchar(100) NULL,
		client_user varchar(50) NULL,
		client_pass varchar(50) NULL,
		client_licence_creation timestamp(0) NULL,
		client_licence_expire timestamp(0) NULL,
		client_licence_id varchar(100) NULL,
		client_licence_verified bool NOT NULL,
		CONSTRAINT utm_client_pkey PRIMARY KEY (id)
	);`
	err = execPsql(containerID, "utmstack", createTable)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	// Insert client data
	err = execPsql(containerID, "utmstack", "INSERT INTO public.utm_client (client_licence_verified) VALUES (false);")
	if err != nil && !strings.Contains(err.Error(), "duplicate key") {
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

func InitPgUserAuditor(_ *config.Config) error {
	containerID, err := getPostgresContainerID()
	if err != nil {
		return err
	}

	// Creating userauditor database
	err = execPsql(containerID, "", "CREATE DATABASE userauditor")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	return nil
}
