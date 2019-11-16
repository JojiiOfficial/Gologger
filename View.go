package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mkideal/cli"
)

type viewT struct {
	cli.Helper
	ConfigFile     string   `cli:"C,config" usage:"Specify the config file" dft:"config.json"`
	HostnameFilter []string `cli:"H,hostname" usage:"View logs from specific hostname (negatable with \\! before the first element)"`
	MessageFilter  []string `cli:"M,message" usage:"View logs with specific keywords message (negatable with \\! before the first element)"`
	TagFilter      []string `cli:"T,Tag" usage:"View logs from a specific tag (negatable with \\! before the first element)"`
	Since          string   `cli:"s,since" usage:"View logs since a point in time"`
	Until          string   `cli:"u,until" usage:"Show only logs until a secific time"`
	All            bool     `cli:"a,all" usage:"show everything from time 0"`
	FilterOperator bool     `cli:"o,or" usage:"Specify if only one of your filter must match to get an entry 'or' dft: candc" dft:"false"`
	Follow         bool     `cli:"f,follow" usage:"follow log content"`
	Reverse        bool     `cli:"r,reverse" usage:"View in reversed order" dft:"false"`
	NLogs          int      `cli:"n,nums" usage:"Show last n logs (or n logs from -s or -t)"`
	Raw            bool     `cli:"raw" usage:"View logs raw (without counting, ect...)"`
	Yes            bool     `cli:"y,yes" usage:"Dotn't show confirm messages" dft:"false"`
	NoColor        bool     `cli:"no-color" usage:"Don't show colors"`
}

var isDurRegex *regexp.Regexp
var sinceTime, untilTime int64

func genInvalidCombinationErr(mod string, notCompatible ...string) error {
	var e string
	for _, s := range notCompatible {
		if !strings.HasPrefix(s, "-") {
			s = "-" + s
		}
		if len(e) > 0 {
			e += " and " + s
		} else {
			e = s
		}
	}
	return errors.New("can't " + mod + " " + e + " together")
}

func (argv *viewT) Validate(ctx *cli.Context) error {
	if argv.Reverse && argv.Follow {
		return genInvalidCombinationErr("use", "s", "t")
	}

	if argv.All && argv.Follow {
		return genInvalidCombinationErr("use", "f", "a")
	}

	if argv.All && len(argv.Since) > 0 {
		return errors.New("can't view everything and set a starttime at once (-a and -s)")
	}
	if argv.Reverse && argv.Follow {
		return genInvalidCombinationErr("use", "r", "f")
	}

	nLogsSet := argv.NLogs > 0
	if nLogsSet && argv.Follow {
		return genInvalidCombinationErr("use", "f", "n")
	}

	if len(argv.Until) > 0 {
		st, err := parseTimeParam(argv.Until)
		if err != nil {
			return err
		}
		untilTime = int64(st)
	}

	if len(argv.Since) > 0 {
		st, err := parseTimeParam(argv.Since)
		if err != nil {
			return err
		}
		sinceTime = int64(st) - 1
	}

	return nil
}

func parseTimeParam(param string) (uint64, error) {
	param = strings.ToLower(strings.Trim(param, " "))
	if len(param) == 0 {
		return 0, nil
	}
	if isDurRegex == nil {
		isDurRegex, _ = regexp.Compile("(?i)[0-9]+(s|m|h|d|w)$")
	}
	if isDurRegex.MatchString(param) {
		var factor uint64
		var count uint64
		var t string
		timeFactorts := []uint64{1, 60, 60 * 60, 60 * 60 * 24, 60 * 60 * 24 * 7}
		for i, e := range []string{"s", "m", "h", "d", "w"} {
			if strings.HasSuffix(param, e) {
				t = strings.ReplaceAll(param, e, "")
				factor = timeFactorts[i]
				var err error
				count, err = strconv.ParseUint(t, 10, 64)
				count = uint64(math.Abs(float64(count)))
				if err != nil {
					return 0, err
				}
				break
			}
		}

		if count*factor > 18446744073709551615 {
			return 0, errors.New("Overflows uint64")
		}
		return uint64(time.Now().Unix()) - count*factor, nil
	}
	timeFormats := []string{
		time.Stamp,
		time.ANSIC,
		time.RFC822,
		time.RFC822Z,
		time.UnixDate,
	}
	var t time.Time
	var err error
	for _, ti := range timeFormats {
		t, err = time.ParseInLocation(ti, param, time.Now().Location())
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, err
	}
	if t.Year() == 0 {
		t = t.AddDate(time.Now().Year(), 0, 0)
	}
	if time.Now().Sub(t) < 0 {
		return 0, errors.New("Time must be in past")
	}

	return uint64(t.Unix()), nil
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
			fmt.Println("Config created successfully: \"" + getConfFile(argv.ConfigFile) + "\". You need to set \"host\" and \"token\"")
			return nil
		}
		if len(strings.Trim(config.Host, " ")) < 1 || len(strings.Trim(config.Token, " ")) < 1 {
			fmt.Println("You need to fill \"host\" and \"token\" in", getConfFile(argv.ConfigFile))
			return nil
		}

		reader := bufio.NewReader(os.Stdin)

		if argv.All && len(argv.HostnameFilter) == 0 && len(argv.TagFilter) == 0 && len(argv.MessageFilter) == 0 && argv.NLogs == 0 && !argv.Yes {
			y, _ := confirmInput("You didn't set a filter. Do you really want to show everything [y/n]> ", reader)
			if !y {
				return nil
			}
		}

		InitFilter(&argv.HostnameFilter, true)
		InitFilter(&argv.TagFilter, true)
		InitFilter(&argv.MessageFilter, true)

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

//TimeIn time in location
func TimeIn(t time.Time, name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	if err != nil {
		panic(err)
	}
	return t
}

func pullLogs(config *Config, argv *viewT) {
	fetchLogsReques := FetchLogsRequest{}
	fetchLogsReques.Token = config.Token
	fetchLogsReques.Follow = argv.Follow
	fetchLogsReques.Reverse = argv.Reverse
	fetchLogsReques.LogType = 0
	if len(argv.MessageFilter) > 0 {
		fetchLogsReques.MessageFilter = argv.MessageFilter
	}
	if len(argv.HostnameFilter) > 0 {
		fetchLogsReques.HostnameFilter = argv.HostnameFilter
	}
	if len(argv.TagFilter) > 0 {
		fetchLogsReques.TagFilter = argv.TagFilter
	}
	if argv.NLogs > 0 {
		fetchLogsReques.Limit = argv.NLogs
	}
	if argv.FilterOperator {
		fetchLogsReques.FilterOperator = argv.FilterOperator
	}
	if untilTime > 0 {
		fetchLogsReques.Until = untilTime
	}
	if sinceTime > 0 {
		fetchLogsReques.Since = sinceTime
	} else {
		fetchLogsReques.Since = config.LastView
		if config.LastView-3600 > time.Now().Unix() {
			fetchLogsReques.Since = time.Now().Unix() - 3600
		}
	}
	if argv.All {
		fetchLogsReques.Since = 0
	}

	for ok := true; ok; ok = argv.Follow {
		timeout := 5 * time.Second
		if argv.Follow {
			timeout = 2 * time.Minute
		}
		d, err := json.Marshal(fetchLogsReques)
		if err != nil {
			fmt.Println("Error creating json: " + err.Error())
			return
		}
		res, err := request(config.Host, "glog/fetch", d, config.IgnoreCert, timeout)
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
			if len(response.Logs) == 0 && !argv.Follow {
				fmt.Println("No new log since", parseTime(fetchLogsReques.Since))
			} else {
				viewSyslogEntries(response, argv, !argv.Follow)
			}

			fetchLogsReques.Since = response.Time

			//Don't save if everything was fetched or if following
			if !argv.All {
				config.LastView = response.Time
				config.Save(getConfFile(argv.ConfigFile))
			}
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

//Yellow fg color
var Yellow = color.New(color.FgHiBlue).SprintFunc()

func viewSyslogEntries(fetchlogResponse *FetchSysLogResponse, argv *viewT, showTimes bool) {
	if showTimes {
		firstTime := fetchlogResponse.Logs[0].Date
		lastTime := fetchlogResponse.Logs[len(fetchlogResponse.Logs)-1].Date

		fmt.Println("----->>", GreenBold(parseTime(firstTime)), "------ to ------->>", GreenBold(parseTime(lastTime)))
		fmt.Print("\n")
	}
	for _, logEntry := range fetchlogResponse.Logs {
		if logEntry.Count > 1 && !argv.Raw {
			fmt.Printf("%s %s %s(%d) %s%s\n", parseTime(logEntry.Date), logEntry.Hostname, logEntry.Tag, logEntry.PID, logEntry.Message, Yellow("(", logEntry.Count, "x)"))
		} else {
			for i := 0; i < logEntry.Count; i++ {
				fmt.Printf("%s %s %s(%d) %s\n", parseTime(logEntry.Date), logEntry.Hostname, logEntry.Tag, logEntry.PID, logEntry.Message)
			}
		}
	}
}

func parseTime(unix int64) string {
	return time.Unix(unix, 0).Format(time.Stamp)
}
