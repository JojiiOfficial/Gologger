package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
)

//Config config
type Config struct {
	LastView   int64  `json:"lv"`
	Host       string `json:"host"`
	Token      string `json:"token"`
	IgnoreCert bool   `json:"ignoreCert"`
}

func loadConfig(file string) (*Config, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

//Save saves a config file
func (config *Config) Save(file string) error {
	bts, err := json.Marshal(*config)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, bts, "", "\t")

	return ioutil.WriteFile(file, out.Bytes(), 0600)
}

func getHome() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error retrieving homedir: ", err.Error())
		os.Exit(1)
	}
	return usr.HomeDir
}

func getConfPath() string {
	return getHome() + "/" + ".gologger/"
}

func getConfFile(confName string) string {
	confPath := getConfPath()
	if !strings.HasSuffix(confName, ".json") {
		confName += ".json"
	}
	return confPath + confName
}

func checkConfig(file string) (config *Config, err error) {
	createDirIfNotExist(getConfPath())
	confFile := getConfFile(file)
	if _, err := os.Stat(confFile); err != nil {
		y, _ := confirmInput("Config \""+file+"\" doesn't exists. Do you want to create a new config[y/n/a]> ", bufio.NewReader(os.Stdin))
		if !y {
			os.Exit(0)
			return nil, nil
		}
		_, err = os.Create(confFile)
		if err != nil {
			return nil, err
		}
		(&Config{}).Save(confFile)
		return nil, nil
	}
	return loadConfig(confFile)
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			panic(err)
		}
	}
}
