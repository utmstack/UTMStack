package processor

import (
	"regexp"
	"strings"
)

const ipv4Regex = `^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.` +
	`(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.` +
	`(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.` +
	`(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)$`

const ipv6Regex = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|` +
	`([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|` +
	`([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|` +
	`([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|` +
	`[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|` +
	`fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}` +
	`((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|` +
	`(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:` +
	`((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`

var (
	ipv4Pattern = regexp.MustCompile(ipv4Regex)
	ipv6Pattern = regexp.MustCompile(ipv6Regex)
)

func isIP(ip string) bool {
	return ipv4Pattern.MatchString(ip) || ipv6Pattern.MatchString(ip)
}

func getPositionValue(arr []string, position int) string {
	if position < len(arr) {
		return arr[position]
	}
	return ""
}

func buildTrafficLog(awsLog []string) map[string]interface{} {
	return map[string]interface{}{
		"version":     getPositionValue(awsLog, 0),
		"accountId":   getPositionValue(awsLog, 1),
		"interfaceId": getPositionValue(awsLog, 2),
		"src_ip":      getPositionValue(awsLog, 3),
		"dest_ip":     getPositionValue(awsLog, 4),
		"src_port":    getPositionValue(awsLog, 5),
		"dest_port":   getPositionValue(awsLog, 6),
		"protocol":    getPositionValue(awsLog, 7),
		"packets":     getPositionValue(awsLog, 8),
		"bytes":       getPositionValue(awsLog, 9),
		"start":       getPositionValue(awsLog, 10),
		"end":         getPositionValue(awsLog, 11),
		"action":      getPositionValue(awsLog, 12),
		"logStatus":   getPositionValue(awsLog, 13),
	}
}

func buildTPCLog(awsLog []string) map[string]interface{} {
	return map[string]interface{}{
		"version":       getPositionValue(awsLog, 0),
		"vpcId":         getPositionValue(awsLog, 1),
		"subnetId":      getPositionValue(awsLog, 2),
		"instanceId":    getPositionValue(awsLog, 3),
		"interfaceId":   getPositionValue(awsLog, 4),
		"accountId":     getPositionValue(awsLog, 5),
		"type":          getPositionValue(awsLog, 6),
		"src_ip":        getPositionValue(awsLog, 7),
		"dest_ip":       getPositionValue(awsLog, 8),
		"src_port":      getPositionValue(awsLog, 9),
		"dest_port":     getPositionValue(awsLog, 10),
		"packetSrcAddr": getPositionValue(awsLog, 11),
		"packetDstAddr": getPositionValue(awsLog, 12),
		"protocol":      getPositionValue(awsLog, 13),
		"bytes":         getPositionValue(awsLog, 14),
		"packets":       getPositionValue(awsLog, 15),
		"start":         getPositionValue(awsLog, 16),
		"end":           getPositionValue(awsLog, 17),
		"action":        getPositionValue(awsLog, 18),
		"tcpFlags":      getPositionValue(awsLog, 19),
		"logStatus":     getPositionValue(awsLog, 20),
	}
}

func ProcessAwsFlowLog(awsLog string) map[string]interface{} {
	logArr := strings.Fields(awsLog)
	if len(logArr) < 14 {
		return map[string]interface{}{"message": awsLog}
	}
	if logArr[3] != "-" && isIP(logArr[3]) && isIP(logArr[4]) {
		return buildTrafficLog(logArr)
	}
	if logArr[3] == "-" {
		return map[string]interface{}{"message": awsLog}
	}
	return buildTPCLog(logArr)
}
