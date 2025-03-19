//go:build windows && arm64
// +build windows,arm64

package collectors

import (
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/utils"
)

type Windows struct{}

const PowerShellScript = `
<#
.SYNOPSIS
  Collects Windows Application, System, and Security logs from the last 5 minutes and prints each log as a minimal single-line JSON object.
  If no logs are found, no output is produced.

.DESCRIPTION
  1. Retrieves the last 5 minutes of Windows logs (Application, System, Security) using FilterHashtable (no post-fetch filtering).
  2. Transforms TimeCreated into a human-readable format.
  3. Prints each log record as a single-line JSON object with no indentation or extra spacing.
#>

# Suppress any script errors to avoid cluttering the console output
$ErrorActionPreference = 'SilentlyContinue'

# Calculate the start time for filtering
$startTime = (Get-Date).AddSeconds(-30)

# Retrieve logs with filter hashtable
$applicationLogs = Get-WinEvent -FilterHashtable @{ LogName='Application'; StartTime=$startTime }
$systemLogs      = Get-WinEvent -FilterHashtable @{ LogName='System'; StartTime=$startTime }
$securityLogs    = Get-WinEvent -FilterHashtable @{ LogName='Security'; StartTime=$startTime }

# Ensure null collections are treated as empty
if (-not $applicationLogs) { $applicationLogs = @() }
if (-not $systemLogs)      { $systemLogs = @() }
if (-not $securityLogs)    { $securityLogs = @() }

# Combine them
$recentLogs = $applicationLogs + $systemLogs + $securityLogs

# If no logs are found, produce no output
if (-not $recentLogs) {
    return
}

# Transform TimeCreated and output as single-line JSON
$processedLogs = $recentLogs | ForEach-Object {
    $item = $_ | Select-Object *
    $item | Add-Member -MemberType NoteProperty -Name TimeCreated -Value ($_.TimeCreated.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')) -Force
    $item
}

# Print each log record as a single-line, compact JSON object
foreach ($log in $processedLogs) {
    # Convert to JSON with normal indentation
    $json = $log | ConvertTo-Json -Depth 20
    # Remove all newlines and indentation
    $compactJson = $json -replace '(\r?\n\s*)+', ''
    # Print to console (standard output) without additional formatting
    Write-Output $compactJson
}
`

func (w Windows) Install() error {
	path := utils.GetMyPath()
	collectorPath := filepath.Join(path, "collector.ps1")
	err := os.WriteFile(collectorPath, []byte(PowerShellScript), 0644)
	return err
}

func (w Windows) SendSystemLogs() {
	path := utils.GetMyPath()
	collectorPath := filepath.Join(path, "collector.ps1")

	for {
		select {
		case <-time.After(30 * time.Second):
			go func() {
				cmd := exec.Command("Powershell.exe", "-File", collectorPath)

				output, err := cmd.Output()
				if err != nil {
					utils.Logger.ErrorF("error executing powershell script: %v", err)
					return
				}

				logLines := strings.Split(string(output), "\n")

				validatedLogs := make([]string, 0, len(logLines))

				for logLine := range logLines {
					validatedLog, _, err := validations.ValidateString(logLine, false)
					if err != nil {
						utils.Logger.ErrorF("error validating log: %s: %v", logLine, err)
						continue
					}

					validatedLogs = append(validatedLogs, validatedLog)
				}

				logservice.LogQueue <- logservice.LogPipe{
					Src:  string(config.DataTypeWindowsAgent),
					Logs: validatedLogs,
				}
			}()
		}
	}
}

func (w Windows) Uninstall() error {
	path := utils.GetMyPath()
	collectorPath := filepath.Join(path, "collector.ps1")
	err := os.Remove(collectorPath)
	return err
}
