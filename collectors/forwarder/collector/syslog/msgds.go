package syslog

// MSGDS carries a raw syslog message together with its source address.
// Used internally to pass messages between the network reader and the log emitter.
type MSGDS struct {
	DataSource string
	Message    string
}
