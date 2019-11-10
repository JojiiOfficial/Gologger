package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mkideal/cli"
)

type viewT struct {
	cli.Helper
	ConfigFile string `cli:"C,config" usage:"Specify the config file" dft:"config.json"`
	Follow     bool   `cli:"f,follow" usage:"follow log content"`
}

var viewCMD = &cli.Command{
	Argv: func() interface{} { return new(viewT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*viewT)
		config, err := checkConfig(argv.ConfigFile)
		if err != nil {
			fmt.Println("Error creating config:", err.Error())
			return nil
		}
		if err == nil && config == nil {
			fmt.Println("Config created successfully: \"" + getConfFile(argv.ConfigFile) + "\". You neet to set \"host\" and \"token\"")
			return nil
		}
		if len(strings.Trim(config.Host, " ")) < 1 || len(strings.Trim(config.Token, " ")) < 1 {
			fmt.Println("You need to fill \"host\" and \"token\" in", getConfFile(argv.ConfigFile))
			return nil
		}
		pullLogs(config, argv)
		return nil
	},
}

func pullLogs(config *Config, argv *viewT) {
	fetchLogsReques := FetchLogsRequest{}
	fetchLogsReques.Token = config.Token
	fetchLogsReques.Follow = argv.Follow
	fetchLogsReques.LogType = 0
	fetchLogsReques.Since = config.LastView
	if config.LastView-3600 > time.Now().Unix() {
		fetchLogsReques.Since = time.Now().Unix() - 3600
	}

	for ok := true; ok; ok = argv.Follow {
		fetchLogsReques.Follow = argv.Follow
		fetchLogsReques.Since = config.LastView
		timeout := 0 * time.Second
		if argv.Follow {
			timeout = 5 * time.Minute
		}
		d, err := json.Marshal(fetchLogsReques)
		if err != nil {
			fmt.Println("Error creating json: " + err.Error())
			return
		}
		res, err := request(config.Host, "fetch", d, config.IgnoreCert, timeout)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if fetchLogsReques.LogType == 0 {
			response, err := parseSyslogResponse(res)
			if err != nil {
				fmt.Println("Error fetching: " + err.Error())
				return
			}
			config.LastView = response.Time
			config.Save(getConfFile(argv.ConfigFile))
			viewSyslogEntries(response, argv)
		} else {
			return
		}
	}
}

func parseSyslogResponse(src string) (*FetchSysLogResponse, error) {
	response := FetchSysLogResponse{}
	err := json.Unmarshal([]byte(src), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func viewSyslogEntries(fetchlogResponse *FetchSysLogResponse, argv *viewT) {
	for _, logEntry := range fetchlogResponse.Logs {
		fmt.Printf("%s %s %s %s\n", parseTime(logEntry.Date), logEntry.Hostname, logEntry.Tag, logEntry.Message)
	}
}

func parseTime(unix int64) string {
	return time.Unix(unix, 0).Format(time.Stamp)
}
