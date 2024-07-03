package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"path"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/plugins/events/search"
	"github.com/utmstack/UTMStack/plugins/events/statistics"
	"google.golang.org/grpc"
)

type analysisServer struct {
	plugins.UnimplementedAnalysisServer
}

func main() {
	helpers.NewLogger(helpers.GetCfg().Env.LogLevel)

	os.Remove(path.Join(
		helpers.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.events_analysis.sock"))

	laddr, err := net.ResolveUnixAddr("unix",
		path.Join(helpers.GetCfg().Env.Workdir, "sockets",
			"com.utmstack.events_analysis.sock"))

	if err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterAnalysisServer(grpcServer, &analysisServer{})

	// start goroutines
	go statistics.Update()

	if err := grpcServer.Serve(listener); err != nil {
		helpers.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

type Log struct {
	Id         string                            `json:"id"`
	Timestamp  string                            `json:"@timestamp"`
	DeviceTime string                            `json:"deviceTime"`
	DataType   string                            `json:"dataType"`
	DataSource string                            `json:"dataSource"`
	Log        map[string]map[string]interface{} `json:"logx"`
}

func (p *analysisServer) Analyze(ctx context.Context,
	event *plugins.Event) (*plugins.Alert, error) {

	var l Log

	l.Log = make(map[string]map[string]interface{})

	l.Id = event.GetId()
	l.Timestamp = event.GetTimestamp()
	l.DeviceTime = event.GetDeviceTime()
	l.DataType = event.GetDataType()
	l.DataSource = event.GetDataSource()

	logBytes, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	logBuffer := bytes.NewBuffer(logBytes)

	jLog := logBuffer.String()

	l.Log[l.DataType] = func() map[string]interface{} {
		t := gjson.Get(jLog, "log").Value()
		if t == nil {
			return map[string]interface{}{}
		}
		return t.(map[string]interface{})
	}()

	l.Log[l.DataType]["remote"] = func() map[string]interface{} {
		t := gjson.Get(jLog, "remote").Value()
		if t == nil {
			return map[string]interface{}{}
		}
		return t.(map[string]interface{})
	}()

	l.Log[l.DataType]["local"] = func() map[string]interface{} {
		t := gjson.Get(jLog, "local").Value()
		if t == nil {
			return map[string]interface{}{}
		}
		return t.(map[string]interface{})
	}()

	l.Log[l.DataType]["from"] = func() map[string]interface{} {
		t := gjson.Get(jLog, "from").Value()
		if t == nil {
			return map[string]interface{}{}
		}
		return t.(map[string]interface{})
	}()

	l.Log[l.DataType]["to"] = func() map[string]interface{} {
		t := gjson.Get(jLog, "to").Value()
		if t == nil {
			return map[string]interface{}{}
		}
		return t.(map[string]interface{})
	}()

	l.Log[l.DataType]["tenantId"] = event.GetTenantId()
	l.Log[l.DataType]["tenantName"] = event.GetTenantName()
	l.Log[l.DataType]["protocol"] = event.GetProtocol()
	l.Log[l.DataType]["connectionStatus"] = event.GetConnectionStatus()
	l.Log[l.DataType]["statusCode"] = event.GetStatusCode()
	l.Log[l.DataType]["actionResult"] = event.GetActionResult()

	lJs, _ := json.Marshal(l)

	search.AddToQueue(string(lJs))
	statistics.One(l.DataType, l.DataSource)

	return nil, nil
}
