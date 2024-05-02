package logservice

import (
	"github.com/utmstack/UTMStack/log-auth-proxy/config"
)

type LogData struct {
	logType config.LogType
	data    string
}

type LogService struct {
	logChans map[config.LogType]chan LogData
}

//func NewLogQueueService(connections *connection.SingletonConnections) *LogService {
//	//logTypes := []model.LogType{model.Winlogbeat, model.Filebeat, model.Syslog, model.Http, model.TCP}
//	logChans := make(map[model.LogType]chan LogData)
//
//	service := &LogService{
//		Connections: connections,
//		logChans:    logChans,
//	}
//
//	//for _, logType := range logTypes {
//	//	logChans[logType] = make(chan LogData, 1000)
//	//	go service.writeToFile(logType)
//	//	go service.sendToTCP(logType)
//	//}
//
//	return service
//}

// func (l *LogService) writeToFile(logType model.LogType) {
// 	for logData := range l.logChans[logType] {
// 		file, err := os.OpenFile(string(logData.logType)+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 		if _, err := file.Write([]byte(logData.data + "\n")); err != nil {
// 			log.Println(err)
// 			file.Close()
// 			continue
// 		}
// 		if err := file.Close(); err != nil {
// 			log.Println(err)
// 		}
// 	}
// }

//func (l *LogService) sendToTCP(logType model.LogType) {
//	for logData := range l.logChans[logType] {
//		l.Connections.SendLog(logData.logType, logData.data)
//	}
//}

func (l *LogService) LogData(logType config.LogType, data string) {
	l.logChans[logType] <- LogData{logType: logType, data: data}
}

func (l *LogService) LogBulkData(logType config.LogType, data []string) {
	for _, row := range data {
		l.logChans[logType] <- LogData{logType: logType, data: row}
	}
}
