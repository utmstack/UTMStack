package modules

import (
	"fmt"
	"os"
	"strings"
)

func WriteCollectorConfig(integrations map[string]Integration, filename string) error {
	fileContent := "{\n    \"integrations\": {\n"
	for name, integration := range integrations {
		fileContent += fmt.Sprintf("        \"%s\": {\n", name)
		if integration.TCP.Port != "" {
			fileContent += fmt.Sprintf("            \"tcp_port\": {\"enabled\": %t, \"value\": \"%s\"},\n", integration.TCP.IsListen, integration.TCP.Port)
		}
		if integration.UDP.Port != "" {
			fileContent += fmt.Sprintf("            \"udp_port\": {\"enabled\": %t, \"value\": \"%s\"},\n", integration.UDP.IsListen, integration.UDP.Port)
		}
		if strings.HasSuffix(fileContent, ",\n") {
			fileContent = fileContent[:len(fileContent)-2] + "\n"
		}
		fileContent += "        },\n"
	}
	if strings.HasSuffix(fileContent, ",\n") {
		fileContent = fileContent[:len(fileContent)-2] + "\n"
	}
	fileContent += "    }\n}\n"

	err := os.WriteFile(filename, []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func WriteCollectorConfigFromModules(mod []Module, filename string) error {
	integrations := make(map[string]Integration)
	for _, m := range mod {
		integrations[string(m.GetDataType())] = Integration{
			TCP: Port{
				IsListen: m.IsPortListen("tcp"),
				Port:     m.GetPort("tcp"),
			},
			UDP: Port{
				IsListen: m.IsPortListen("udp"),
				Port:     m.GetPort("udp"),
			},
		}
	}
	return WriteCollectorConfig(integrations, filename)
}
