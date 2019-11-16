package main

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"
)

func parseFetchLogsResponse(src string) (*FetchLogResponse, error) {
	response := FetchLogResponse{}
	err := json.Unmarshal([]byte(src), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func parseTime(unix int64) string {
	return time.Unix(unix, 0).Format(time.Stamp)
}

func mergeLogs(syslogs []SyslogEntry, customLogs []CustomLogEntry) []MergedLog {
	var merged []MergedLog
	for _, log := range syslogs {
		merged = append(merged, syslogToMerged(log))
	}
	for _, log := range customLogs {
		merged = append(merged, custlogToMerged(log))
	}
	sort.Slice(merged, func(p, q int) bool {
		return merged[p].Date < merged[q].Date
	})
	return merged
}

func syslogToMerged(log SyslogEntry) MergedLog {
	msg := log.Message
	if log.LogLevel > 0 {
		msg = "(" + strconv.Itoa(log.LogLevel) + ") " + msg
	}
	tag := log.Tag + "(" + strconv.Itoa(log.PID) + ")"
	return MergedLog{
		Date:     log.Date,
		Hostname: log.Hostname,
		Tag:      tag,
		Message:  msg,
		Count:    log.Count,
	}
}

func custlogToMerged(log CustomLogEntry) MergedLog {
	msg := log.Message
	if len(log.Source) > 0 {
		msg = "[" + log.Source + "] " + msg
	}
	return MergedLog{
		Date:     log.Date,
		Hostname: log.Hostname,
		Tag:      log.Tag,
		Message:  msg,
		Count:    log.Count,
	}
}
