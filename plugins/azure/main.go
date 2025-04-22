package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azquery"
	"github.com/google/uuid"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const (
	delayCheck                = 300
	defaultTenant      string = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection        = "https://login.microsoftonline.com/"
	wait                      = 1 * time.Second
)

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go plugins.SendLogsFromChannel()
	}

	for {
		if err := connectionChecker(urlCheckConnection); err != nil {
			_ = catcher.Error("External connection failure detected: %v", err, nil)
		}

		utmConfig := plugins.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		moduleConfig, err := client.GetUTMConfig(enum.AZURE)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = catcher.Error("cannot obtain module configuration", err, nil)
			}

			time.Sleep(time.Second * delayCheck)
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							skip = true
							break
						}
					}

					if !skip {
						azureProcessor := getAzureProcessor(group)
						azureProcessor.pull()
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		}

		time.Sleep(time.Second * delayCheck)
	}
}

type AzureConfig struct {
	GroupName         string
	TenantID          string
	ClientID          string
	ClientSecretValue string
	WorkspaceID       string
}

// Debug Change to real config names
func getAzureProcessor(group types.ModuleGroup) AzureConfig {
	azurePro := AzureConfig{}
	azurePro.GroupName = group.GroupName
	for _, cnf := range group.Configurations {
		switch cnf.ConfName {
		case "Event Hub Shared access policies - Connection string":
			azurePro.TenantID = cnf.ConfValue
		case "Consumer Group Name":
			azurePro.ClientID = cnf.ConfValue
		case "Storage Container Name":
			azurePro.ClientSecretValue = cnf.ConfValue
		case "Storage account connection string with key":
			azurePro.WorkspaceID = cnf.ConfValue
		}
	}
	return azurePro
}

func (ap *AzureConfig) pull() {
	cred, err := azidentity.NewClientSecretCredential(ap.TenantID, ap.ClientID, ap.ClientSecretValue, nil)
	if err != nil {
		_ = catcher.Error("cannot obtain Azure credentials", err, map[string]any{
			"group": ap.GroupName,
		})
		return
	}

	client, err := azquery.NewLogsClient(cred, nil)
	if err != nil {
		_ = catcher.Error("cannot create Logs client", err, map[string]any{
			"group": ap.GroupName,
		})
		return
	}

	res, err := client.QueryWorkspace(
		context.TODO(),
		ap.WorkspaceID,
		azquery.Body{
			Query: to.Ptr("union * | where TimeGenerated >= ago(5m)| order by TimeGenerated desc"),
		},
		nil,
	)
	if err != nil {
		_ = catcher.Error("cannot query Logs", err, map[string]any{
			"group": ap.GroupName,
		})
		return
	}
	if res.Error != nil {
		_ = catcher.Error("cannot query Logs", err, map[string]any{
			"group": ap.GroupName,
		})
		return
	}

	var logs []map[string]any
	for _, table := range res.Tables {
		for _, row := range table.Rows {
			rowMap := make(map[string]any)
			for i, column := range table.Columns {
				if row[i] != nil {
					if str, ok := row[i].(string); ok && str == "" {
						continue
					}
					rowMap[*column.Name] = row[i]
				}
			}
			logs = append(logs, rowMap)
		}
	}

	if len(logs) > 0 {
		for _, log := range logs {
			jsonLog, err := json.Marshal(log)
			if err != nil {
				_ = catcher.Error("cannot encode log to JSON", err, map[string]any{
					"group": ap.GroupName,
				})
				continue
			}
			plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.New().String(),
				TenantId:   defaultTenant,
				DataType:   "azure",
				DataSource: ap.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(jsonLog),
			})
		}
	}
}

func connectionChecker(url string) error {
	checkConn := func() error {
		if err := checkConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}
		return nil
	}

	if err := infiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkConnection(url string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			_ = catcher.Error("cannot close response body", err, nil)
		}
	}()

	return nil
}

func infiniteRetryIfXError(f func() error, exception string) error {
	var xErrorWasLogged bool

	for {
		err := f()
		if err != nil && is(err, exception) {
			if !xErrorWasLogged {
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, nil)
				xErrorWasLogged = true
			}
			time.Sleep(wait)
			continue
		}

		return err
	}
}

func is(e error, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
