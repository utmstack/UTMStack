package logservice

type Grpc struct {
	UnimplementedLogServiceServer
	OutputService *LogOutputService
}
