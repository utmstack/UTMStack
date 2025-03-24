//go:build windows && arm64
// +build windows,arm64

package collectors

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"

	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/utils"
)

type Windows struct{}

func getCollectorsInstances() []Collector {
	var collectors []Collector
	collectors = append(collectors, Windows{})
	return collectors
}

const PowerShellScript = `
<#
.SYNOPSIS
 Collects Windows Application, System, and Security logs from the last 5 minutes, then prints them to the console in a compact, single-line JSON format,
 emulating the field structure that Winlogbeat typically produces.

.DESCRIPTION
 1. Retrieves the last 5 minutes of Windows logs (Application, System, Security) using FilterHashtable (no post-fetch filtering).
 2. Maps event properties to a schema similar to Winlogbeat's, including:
   - @timestamp
   - message
   - event.code
   - event.provider
   - event.kind
   - winlog fields (e.g. record_id, channel, activity_id, etc.)
 3. Prints each log record as a single-line JSON object with no indentation/extra spacing.
 4. If no logs are found, the script produces no output at all.
#>

# Suppress any runtime errors that would clutter the console.
$ErrorActionPreference = 'SilentlyContinue'

# Calculate the start time for filtering
$startTime = (Get-Date).AddSeconds(-30)

# Retrieve logs with filter hashtable
$applicationLogs = Get-WinEvent -FilterHashtable @{ LogName='Application'; StartTime=$startTime }
$systemLogs   = Get-WinEvent -FilterHashtable @{ LogName='System';   StartTime=$startTime }
$securityLogs  = Get-WinEvent -FilterHashtable @{ LogName='Security';  StartTime=$startTime }

# Safeguard against null results
if (-not $applicationLogs) { $applicationLogs = @() }
if (-not $systemLogs)   { $systemLogs   = @() }
if (-not $securityLogs)  { $securityLogs  = @() }

# Combine them
$recentLogs = $applicationLogs + $systemLogs + $securityLogs

# If no logs are found, produce no output at all
if (-not $recentLogs) {
  return
}

# Function to convert the raw Properties array to a dictionary-like object under winlog.event_data
function Convert-PropertiesToEventData {
  param([Object[]] $Properties)

  # If nothing is there, return an empty hashtable
  if (-not $Properties) { return @{} }

  # Winlogbeat places custom fields under winlog.event_data. 
  # Typically, it tries to parse known keys, but we'll do a simple best-effort approach:
  # We'll create paramN = <value> pairs for each array index.
  $eventData = [ordered]@{}

  for ($i = 0; $i -lt $Properties.Count; $i++) {
    $value = $Properties[$i].Value

    # If the property is itself an object with nested fields, we can flatten or store as-is.
    # We'll store as-is for clarity. 
    # We'll name them param1, param2, param3,... unless you'd like more specific field logic.
    $paramName = "param$($i+1)"

    $eventData[$paramName] = $value
  }

  return $eventData
}

# Transform each event into a structure emulating Winlogbeat
foreach ($rawEvent in $recentLogs) {
  # Convert TimeCreated to a universal ISO8601 string
  $timestamp = $rawEvent.TimeCreated.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')

  # Build the top-level document
  $doc = [ordered]@{
    # Matches Winlogbeat's typical top-level timestamp field
    '@timestamp' = $timestamp

    # The main message content from the event
    'message'  = $rawEvent.Message

    # "event" block: minimal example
    'event' = [ordered]@{
      'code'   = $rawEvent.Id    # event_id in Winlogbeat is typically a string or numeric
      'provider' = $rawEvent.ProviderName
      'kind'   = 'event'
      'created'  = $timestamp     # or you could omit if desired
    }

    # "winlog" block: tries to mirror Winlogbeat's structure for Windows
    'winlog' = [ordered]@{
      'record_id'     = $rawEvent.RecordId
      'computer_name'   = $rawEvent.MachineName
      'channel'      = $rawEvent.LogName
      'provider_name'   = $rawEvent.ProviderName
      'provider_guid'   = $rawEvent.ProviderId
      'process' = [ordered]@{
        'pid' = $rawEvent.ProcessId
        'thread' = @{
          'id' = $rawEvent.ThreadId
        }
      }
      'event_id'      = $rawEvent.Id
      'version'      = $rawEvent.Version
      'activity_id'    = $rawEvent.ActivityId
      'related_activity_id'= $rawEvent.RelatedActivityId
      'task'        = $rawEvent.TaskDisplayName
      'opcode'       = $rawEvent.OpcodeDisplayName
      'keywords'      = $rawEvent.KeywordsDisplayNames
      'time_created'    = $timestamp
      # Convert "Properties" into a dictionary for event_data
      'event_data'     = Convert-PropertiesToEventData $rawEvent.Properties
    }
  }

  # Convert our object to JSON (with no extra formatting).
  $json = $doc | ConvertTo-Json -Depth 20

  # Remove all newlines and indentation for a single-line representation
  $compactJson = $json -replace '(\r?\n\s*)+', ''

  # Output the line
  Write-Output $compactJson
}
`

func (w Windows) Install() error {
	return nil
}

func (w Windows) SendSystemLogs() {
	path := utils.GetMyPath()
	collectorPath := filepath.Join(path, "collector.ps1")

	err := os.WriteFile(collectorPath, []byte(PowerShellScript), 0644)
	if err != nil {
		_ = utils.Logger.ErrorF("error writing powershell script: %v", err)
		return
	}

	cmd := exec.Command("Powershell.exe", "-Command", `"Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser -Force"`)
	err = cmd.Run()
	if err != nil {
		_ = utils.Logger.ErrorF("error setting powershell execution policy: %v", err)
		return
	}

	for {
		select {
		case <-time.After(30 * time.Second):
			go func() {
				cmd := exec.Command("Powershell.exe", "-File", collectorPath)

				output, err := cmd.Output()
				if err != nil {
					_ = utils.Logger.ErrorF("error executing powershell script: %v", err)
					return
				}

				utils.Logger.LogF(100, "output: %s", string(output))

				logLines := strings.Split(string(output), "\n")

				validatedLogs := make([]string, 0, len(logLines))

				for _, logLine := range logLines {
					validatedLog, _, err := validations.ValidateString(logLine, false)
					if err != nil {
						utils.Logger.LogF(100, "error validating log: %s: %v", logLine, err)
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
