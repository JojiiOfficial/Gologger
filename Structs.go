package main

//FetchLogsRequest fetches logs from the server
type FetchLogsRequest struct {
	Token          string   `json:"t"`
	Since          int64    `json:"sin"`
	Until          int64    `json:"unt"`
	LogType        int      `json:"lt"`
	Follow         bool     `json:"foll"`
	HostnameFilter []string `json:"hnf,omitempty"`
	MessageFilter  []string `json:"mf,omitempty"`
	TagFilter      []string `json:"tf,omitempty"`
	FilterOperator bool     `json:"fi,omitempty"`
	Limit          int      `json:"lm,omitempty"`
}

//FetchLogResponse response for fetchlog
type FetchLogResponse struct {
	Time       int64            `json:"t"`
	SysLogs    []SyslogEntry    `json:"slg,omitempty"`
	CustomLogs []CustomLogEntry `json:"clg,omitempty"`
}

//SyslogEntry a log entry in the syslog
type SyslogEntry struct {
	Date     int64  `json:"d"`
	Hostname string `json:"h"`
	Tag      string `json:"t"`
	PID      int    `json:"p"`
	LogLevel int    `json:"l"`
	Message  string `json:"m"`
	Count    int    `json:"c"`
}

//CustomLogEntry a log entry from a custom file
type CustomLogEntry struct {
	Date     int64  `json:"d" mapstructure:"d"`
	Message  string `json:"m" mapstructure:"m"`
	Tag      string `json:"t,omitempty" mapstructure:"t"`
	Source   string `json:"s" mapstructure:"s"`
	Hostname string `json:"h" mapstructure:"h"`
	Count    int    `json:"c" mapstructure:"c"`
}

//MergedLog all logs together
type MergedLog struct {
	Date     int64
	Hostname string
	Tag      string
	Message  string
	Count    int
}
