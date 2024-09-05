package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"

	go_sdk "github.com/threatwinds/go-sdk"
	
	"github.com/utmstack/UTMStack/plugins/alerts/correlation"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type correlationServer struct {
	go_sdk.UnimplementedCorrelationServer
}

func main() {
	os.Remove(path.Join(go_sdk.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.alerts_correlation.sock"))

	laddr, err := net.ResolveUnixAddr(
		"unix", path.Join(go_sdk.GetCfg().Env.Workdir,
			"sockets", "com.utmstack.alerts_correlation.sock"))

	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	go_sdk.RegisterCorrelationServer(grpcServer, &correlationServer{})

	if err := grpcServer.Serve(listener); err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *correlationServer) Correlate(ctx context.Context,
	alert *go_sdk.Alert) (*emptypb.Empty, error) {

	correlation.Alert(
		alert.Id,
		alert.Name,
		alert.Severity,
		alert.Description,
		"",
		alert.Category,
		alert.Technique,
		alert.References,
		alert.DataType,
		alert.DataSource,
		getDetails(alert),
	)

	return nil, nil
}

func getDetails(alert *go_sdk.Alert) map[string]string {
	var details = make(map[string]string, 10)
	if len(alert.Events) != 0 {
		details["id"] = alert.Events[0].Id
		details["Protocol"] = alert.Events[0].Protocol
	}

	if alert.Target != nil {
		details["DestinationUser"] = alert.Target.User
		details["DestinationHost"] = alert.Target.Host
		details["DestinationIP"] = alert.Target.Ip
		details["DestinationPort"] = fmt.Sprintf("%d", alert.Target.Port)

		if alert.Target.Geolocation != nil {
			details["DestinationASN"] = fmt.Sprintf("%d", alert.Target.Geolocation.Asn)
			details["DestinationASO"] = alert.Target.Geolocation.Aso
			details["DestinationLat"] = fmt.Sprintf("%f", alert.Target.Geolocation.Latitude)
			details["DestinationLon"] = fmt.Sprintf("%f", alert.Target.Geolocation.Longitude)
			details["DestinationCity"] = alert.Target.Geolocation.City
			details["DestinationCountry"] = alert.Target.Geolocation.Country
			details["DestinationCountryCode"] = alert.Target.Geolocation.CountryCode
			details["DestinationAccuracyRadius"] = fmt.Sprintf("%d", alert.Target.Geolocation.Accuracy)
		}
	}

	if alert.Adversary != nil {
		details["SourceUser"] = alert.Adversary.User
		details["SourceHost"] = alert.Adversary.Host
		details["SourceIP"] = alert.Adversary.Ip
		details["SourcePort"] = fmt.Sprintf("%d", alert.Adversary.Port)

		if alert.Adversary.Geolocation != nil {
			details["SourceASN"] = fmt.Sprintf("%d", alert.Adversary.Geolocation.Asn)
			details["SourceASO"] = alert.Adversary.Geolocation.Aso
			details["SourceLat"] = fmt.Sprintf("%f", alert.Adversary.Geolocation.Latitude)
			details["SourceLon"] = fmt.Sprintf("%f", alert.Adversary.Geolocation.Longitude)
			details["SourceCity"] = alert.Adversary.Geolocation.City
			details["SourceCountry"] = alert.Adversary.Geolocation.Country
			details["SourceCountryCode"] = alert.Adversary.Geolocation.CountryCode
			details["SourceAccuracyRadius"] = fmt.Sprintf("%d", alert.Adversary.Geolocation.Accuracy)
		}
	}

	return details
}
