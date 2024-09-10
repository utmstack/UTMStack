package agent

import (
	context "context"
	"crypto/tls"
	"net/http"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func CheckGRPCHealth(conn *grpc.ClientConn, ctx context.Context) error {
	healthClient := grpc_health_v1.NewHealthClient(conn)

	for {
		resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			return err
		}

		if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}

func CheckHttpHealth(url string, skipTls bool) {
	client := &http.Client{}
	if skipTls {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	for {
		resp, err := client.Get(url)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			return
		}

		time.Sleep(5 * time.Second)
	}
}