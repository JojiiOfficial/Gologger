package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mkideal/cli"
	clix "github.com/mkideal/cli/ext"
)

type viewT struct {
	cli.Helper
	ConfigFile        string        `cli:"C,config" usage:"Specify the config file" dft:"config.json"`
	Follow            bool          `cli:"f,follow" usage:"follow log content"`
	SincePointInTime  clix.Time     `cli:"t,sincetime" usage:"View logs since a point in time"`
	SinceRelativeTime clix.Duration `cli:"s,since" usage:"View logs since some minutes ago"`
	HostnameFilter    []string      `cli:"H,hostname" usage:"View logs from Specific hostname (negatable with \\! before the first element)"`
	Reverse           bool          `cli:"r,reverse" usage:"View in reversed order" dft:"false"`
	NoColor           bool          `cli:"no-color" usage:"Don't show colors"`
}

var viewCMD = &cli.Command{
	Argv: func() interface{} { return new(viewT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*viewT)
		config, err := checkConfig(argv.ConfigFile)
		if argv.NoColor || os.Getenv("NO_COLOR") == "true" {
			color.NoColor = true
		}
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
		if argv.SincePointInTime.IsSet() && argv.SinceRelativeTime.Seconds() > 0 {
			fmt.Println("Error! You can't set both -s and -t")
			return nil
		}

		if argv.Reverse && argv.Follow {
			fmt.Println("You can't use -t and -r together")
			return nil
		}
		InitFilter(&argv.HostnameFilter, true)

		pullLogs(config, argv)
		return nil
	},
}

//InitFilter split parameter values
func InitFilter(sl *[]string, checkNegation bool) {
	if len(*sl) == 0 {
		*sl = nil
		return
	}
	var e []string
	for _, hn := range *sl {
		if strings.Contains(hn, ",") {
			for _, hh := range strings.Split(hn, ",") {
				if len(hh) == 0 {
					continue
				}
				e = append(e, hh)
			}
		} else {
			if len(hn) == 0 {
				continue
			}
			e = append(e, hn)
		}
	}
	*sl = e
	if checkNegation {
		for i, s := range *sl {
			if i == 0 {
				continue
			}
			if strings.HasPrefix(s, "!") {
				fmt.Println("Error! If you want to negate the filter, use the first element!")
				os.Exit(1)
				return
			}
		}
	}
}

func pullLogs(config *Config, argv *viewT) {
	fetchLogsReques := FetchLogsRequest{}
	fetchLogsReques.Token = config.Token
	fetchLogsReques.Follow = argv.Follow
	fetchLogsReques.Reverse = argv.Reverse
	fetchLogsReques.LogType = 0
	fetchLogsReques.HostnameFilter = argv.HostnameFilter
	if argv.SincePointInTime.IsSet() {
		fetchLogsReques.Since = argv.SincePointInTime.Unix()
		fmt.Println(argv.SincePointInTime.Unix())
	} else if argv.SinceRelativeTime.Seconds() > 0 {
		fetchLogsReques.Since = config.LastView - int64(math.Abs(argv.SinceRelativeTime.Seconds()))
	} else {
		fetchLogsReques.Since = config.LastView
		if config.LastView-3600 > time.Now().Unix() {
			fetchLogsReques.Since = time.Now().Unix() - 3600
		}
	}

	for ok := true; ok; ok = argv.Follow {
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
			viewSyslogEntries(response, argv, !argv.Follow)
			fetchLogsReques.Since = response.Time
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

//GreenBold a green and bold font
var GreenBold = color.New(color.Bold, color.FgHiGreen).SprintFunc()

func viewSyslogEntries(fetchlogResponse *FetchSysLogResponse, argv *viewT, showTimes bool) {
	if len(fetchlogResponse.Logs) == 0 {
		return
	}
	if showTimes {
		firstTime := fetchlogResponse.Logs[0].Date
		lastTime := fetchlogResponse.Logs[len(fetchlogResponse.Logs)-1].Date

		fmt.Println("----->>", GreenBold(parseTime(firstTime)), "------ to ------->>", GreenBold(parseTime(lastTime)))
		fmt.Print("\n")
	}
	for _, logEntry := range fetchlogResponse.Logs {
		fmt.Printf("%s %s %s(%d) %s\n", parseTime(logEntry.Date), logEntry.Hostname, logEntry.Tag, logEntry.PID, logEntry.Message)
	}
}

func parseTime(unix int64) string {
	return time.Unix(unix, 0).Format(time.Stamp)
}
