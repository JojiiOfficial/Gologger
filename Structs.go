package main

//FetchLogsRequest fetches logs from the server
type FetchLogsRequest struct {
	Token   string `json:"t"`
	Since   int64  `json:"sin"`
	LogType int    `json:"lt"`
}

//FetchSysLogResponse response for fetchlog
type FetchSysLogResponse struct {
	Time int64         `json:"t"`
	Logs []SyslogEntry `json:"lgs"`
}

//SyslogEntry a log entry in the syslog
type SyslogEntry struct {
	Date     int64  `json:"d"`
	Hostname string `json:"h"`
	Tag      string `json:"t"`
	PID      int    `json:"p"`
	LogLevel int    `json:"l"`
	Message  string `json:"m"`
}
