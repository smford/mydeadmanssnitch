package main

// snitchit.go

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type oneSnitch struct {
	Token       string    `json:"token"`
	Href        string    `json:"href"`
	Name        string    `json:"name"`
	Tags        []string  `json:"tags"`
	Notes       string    `json:"notes,omitempty"`
	Status      string    `json:"status"`
	CheckedInAt time.Time `json:"checked_in_at"`
	CreatedAt   time.Time `json:"created_at"`
	Interval    string    `json:"interval"`
	AlertType   string    `json:"alert_type"`
}

type newSnitch struct {
	Name      string   `json:"name"`
	AlertType string   `json:"alert_type"`
	Interval  string   `json:"interval"`
	Notes     string   `json:"notes"`
	Tags      []string `json:"tags"`
}

const appversion = 0.01

var (
	apikey        string
	defaultsnitch string
	interval      string
	message       string
	name          string
	showsnitches  bool
	silent        bool
	snitch        string
)

func init() {
	flag.Bool("debug", false, "Enable debugging")
	flag.String("alert", "basic", "Alert type: \"basic\" or \"smart\"")
	flag.String("apikey", "", "Deadmanssnitch.com API Key")
	configFile := flag.String("config", "config.yaml", "Configuration file, default = config.yaml")
	flag.Bool("create", false, "Create a snitch")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help")
	flag.String("interval", "", "\"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
	tempmessage := flag.String("message", "", "Mesage to send, default = \"2006-01-02T15:04:05Z07:00\" format")
	flag.String("name", "", "Name of snitch")
	flag.String("notes", "", "Notes")
	configPath := flag.String("path", ".", "Path to configuration file, default = current directory")
	flag.String("pause", "", "Pause a snitch")
	showsnitches = *flag.Bool("show", false, "Show snitches")
	flag.Bool("silent", false, "Be silent")
	flag.String("snitch", "", "Snitch to use")
	flag.String("tags", "", "Tags")
	flag.String("unpause", "", "Unpause a snitch")
	flag.Bool("version", false, "Version")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	if viper.GetBool("version") {
		fmt.Println(appversion)
		os.Exit(0)
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(*configPath)

	if *tempmessage == "" {
		currenttime := time.Now().Format(time.RFC3339)
		message = currenttime
	} else {
		message = *tempmessage
	}

	config := strings.TrimSuffix(*configFile, ".yaml")
	// fmt.Printf("Loading: %s/%s\n", *configPath, *configFile)

	viper.SetConfigName(config)
	err := viper.ReadInConfig()

	if err != nil {
		if !viper.GetBool("silent") {
			fmt.Printf("%s\n", err)
		}
	}

	// fmt.Println("Snitch:", viper.GetString("snitch"))

	if viper.GetString("snitch") == "" {
		snitch = viper.GetString("defaultsnitch")
	} else {
		snitch = viper.GetString("snitch")
	}

	apikey = viper.GetString("apikey")
	silent = viper.GetBool("silent")

}

func main() {
	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}

	if viper.GetBool("show") {
		displaySnitch(snitch)
		os.Exit(0)
	}

	if viper.GetBool("create") {

		var mytags []string

		mytags = append(mytags, strings.Split(viper.GetString("tags"), ",")...)

		fmt.Println("tags:", mytags)

		newsnitch := newSnitch{Name: viper.GetString("name"), Interval: viper.GetString("interval"), AlertType: viper.GetString("alert"), Notes: viper.GetString("notes"), Tags: mytags}

		//newsnitch["interval"] = viper.GetString("interval")
		//newsnitch["name"] = viper.GetString("name")

		createSnitch(newsnitch)
		os.Exit(0)
	}

	if viper.GetString("pause") != "" {
		pauseSnitch(viper.GetString("pause"))
		os.Exit(0)
	}

	if viper.GetString("unpause") != "" {
		message = "Unpausing: " + message
		unpauseSnitch(viper.GetString("unpause"))
		os.Exit(0)
	}

	if !silent {
		fmt.Println("Message:", message)
	}

	if len(snitch) == 0 {
		fmt.Println("ERROR: No snitch defined")
		os.Exit(1)
	}

	sendSnitch(snitch)
}

func displayConfig() {
	fmt.Println("CONFIG: file :", viper.ConfigFileUsed())
	allmysettings := viper.AllSettings()
	var keys []string
	for k := range allmysettings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("CONFIG:", k, ":", allmysettings[k])
	}
}

func sendSnitch(sendsnitch string) {
	client := &http.Client{}
	client.Timeout = time.Second * 15
	sendsnitch = url.QueryEscape(sendsnitch)
	if !silent {
		fmt.Printf("Snitch: https://nosnch.in/%s\n", sendsnitch)
	}
	uri := fmt.Sprintf("https://nosnch.in/%s", sendsnitch)
	data := url.Values{
		"m": []string{message},
	}
	resp, err := client.PostForm(uri, data)
	if err != nil {
		log.Fatalf("client.PosFormt() failed with '%s'\n", err)
	}
	defer resp.Body.Close()

	snitchresponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ioutil.ReadAll() failed with '%s'\n", err)
	}
	if !silent {
		fmt.Printf("Response: %s\n", snitchresponse)
	}
}

func displaySnitch(snitch string) {

	if len(apikey) == 0 {
		fmt.Println("ERROR: No API Key provided")
		os.Exit(1)
	}

	snitch = url.QueryEscape(snitch)
	url := fmt.Sprintf("https://api.deadmanssnitch.com/v1/snitches/%s", snitch)

	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(apikey, "")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	client := &http.Client{}
	client.Timeout = time.Second * 15

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()

	var mysnitches []oneSnitch

	if snitch != "" {
		var singleSnitch oneSnitch
		if err := json.NewDecoder(resp.Body).Decode(&singleSnitch); err != nil {
			log.Println(err)
		}
		mysnitches = append(mysnitches, singleSnitch)
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&mysnitches); err != nil {
			log.Println(err)
		}
	}

	w := new(tabwriter.Writer)
	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 10, 8, 4, '\t', 0)
	defer w.Flush()
	fmt.Fprintf(w, "\n%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", "Snitch", "Name", "Status", "Last CheckIn", "Interval", "Alert Type", "Notes", "Tags")

	for _, onesnitch := range mysnitches {
		if onesnitch.Token != "" {
			fmt.Fprintf(w, "\n%s\t%s\t%s\t%s\t%s\t%s\t%s\t[%s]\n", onesnitch.Token, onesnitch.Name, onesnitch.Status, onesnitch.CheckedInAt.Format("2006-01-02 15:04:05"), onesnitch.Interval, onesnitch.AlertType, onesnitch.Notes, strings.Join(onesnitch.Tags, ","))
		} else {
			fmt.Fprintf(w, "\n%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "ERROR NO SNITCH FOUND", "", "", "", "", "", "", "")
		}
	}

}

func pauseSnitch(snitch string) {
	fmt.Println("Pausing snitch:", snitch)
	actionSnitch(snitch+"/pause", "POST", "")
}

func unpauseSnitch(snitch string) {
	fmt.Println("Unpausing snitch:", snitch)
	sendSnitch(snitch)
}

func actionSnitch(action string, httpaction string, customheader string) {

	if len(apikey) == 0 {
		fmt.Println("ERROR: No API Key provided")
		os.Exit(1)
	}

	fmt.Println("running action:", action, " ", httpaction)
	snitch = url.QueryEscape(snitch)

	// if doing a create, action should not be appended for the url
	//fmt.Printf("action string: %s https://api.deadmanssnitch.com/v1/snitches/%s\n", httpaction, action)

	url := ""

	if !viper.GetBool("debug") {
		url = "https://api.deadmanssnitch.com/v1/snitches"
	} else {
		// nc -l 127.0.0.1 8888
		url = "http://localhost:8888/v1/snitches"
	}

	if len(httpaction) == 0 {
		url = url + "/" + action
	}

	fmt.Println("url:", url)

	bytesaction := []byte(action)

	//req, err := http.NewRequest(httpaction, url, nil)
	req, err := http.NewRequest(httpaction, url, bytes.NewBuffer(bytesaction))
	req.SetBasicAuth(apikey, "")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	if len(customheader) != 0 {
		fmt.Println("Customheader:", customheader)
		req.Header.Add("Content-Type", customheader)
	} else {
		fmt.Println("standard header")
	}

	client := &http.Client{}
	client.Timeout = time.Second * 15

	resp, err := client.Do(req)
	fmt.Println("response=", err)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()

}

func createSnitch(newsnitch newSnitch) {
	fmt.Println("Creating snitch")

	jsonsnitch, _ := json.Marshal(newsnitch)
	fmt.Println(string(jsonsnitch))

	//snitch = url.QueryEscape(snitch)

	fmt.Println("--------------\n", newsnitch, "\n--------------")

	if len(newsnitch.Name) == 0 {
		fmt.Println("snitch name blank")
		os.Exit(1)
	}

	if len(newsnitch.Interval) == 0 {
		fmt.Println("interval blank")
		os.Exit(1)
	}

	newsnitch.Interval = strings.ToLower(newsnitch.Interval)
	newsnitch.AlertType = strings.ToLower(newsnitch.AlertType)

	{
		fmt.Println("checking snitch")
		switch newsnitch.Interval {
		case "15_minute", "30_minute", "hourly", "daily", "weekly", "monthly":
			fmt.Println("valid interval:", newsnitch.Interval)
		default:
			fmt.Println("invalid interval:", newsnitch.Interval)
		}

		switch newsnitch.AlertType {
		case "basic", "smart":
			fmt.Println("valid alert type:", newsnitch.AlertType)
		default:
			fmt.Println("invalid alert type:", newsnitch.AlertType)
		}
	}

	// check if existing snitch exists
	if !existSnitch(newsnitch) {
		fmt.Printf("Snitch %s already exists\n")
	} else {
		fmt.Println("creating snitch")
		//============
		actionSnitch(string(jsonsnitch), "POST", "application/json")
		//===========

	}

}

func existSnitch(snitch newSnitch) bool {
	fmt.Println("checking existence of snitch:", snitch.Name)
	return true
}

func displayHelp() {
	helpmessage := `
snitchit

  --alert [type]                     Alert type: "basic" or "smart"
  --apikey [api key]                 Deadmanssnitch.com API key
  --config [config file]             Configuration file, default = config.yaml
  --create [snitch]                  Create snitch, requires --name and --interval
  --displayconfig                    Display configuration
  --help                             Display help
  --interval [interval window]       "15_minute", "30_minute", "hourly", "daily", "weekly", or "monthly"
  --message [messgage to send]       Message to send, default = "2006-01-02T15:04:05Z07:00" format
  --name [name]                      Name of snitch
  --notes [notes]                    Notes for snitch
  --path [path to config file]       Path to configuration file, default = current directory
  --pause [snitch]                   Pauses a snitch
  --show                             Display all snitches
  --show --snitch [snitch]           Show details for a specific snitch
  --silent                           Be silent
  --snitch [snitch]                  Snitch to use, default = defaultsnitch from config.yaml
  --unpause [snitch]                 Unpause a snitch
  --version                          Version
`
	fmt.Printf("%s", helpmessage)
}
