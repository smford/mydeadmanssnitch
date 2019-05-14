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
	Href        string    `json:"href,omitempty"`
	Name        string    `json:"name,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Status      string    `json:"status,omitempty"`
	CheckedInAt time.Time `json:"checked_in_at,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	Interval    string    `json:"interval,omitempty"`
	AlertType   string    `json:"alert_type,omitempty"`
}

type newSnitch struct {
	Name      string   `json:"name"`
	AlertType string   `json:"alert_type"`
	Interval  string   `json:"interval"`
	Notes     string   `json:"notes"`
	Tags      []string `json:"tags"`
}

type udSnitch struct {
	Name      string   `json:"name,omitempty"`
	AlertType string   `json:"alert_type,omitempty"`
	Interval  string   `json:"interval,omitempty"`
	Notes     string   `json:"notes,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type dmsResp struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

const appversion = 0.02

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
	flag.String("alert", "basic", "Alert type: \"basic\" or \"smart\"")
	flag.String("apikey", "", "Deadmanssnitch.com API Key")
	configFile := flag.String("config", "config.yaml", "Configuration file, default = config.yaml")
	flag.Bool("create", false, "Create snitch, requires --name and --interval, optional --tags & --notes")
	flag.Bool("debug", false, "Enable debug mode")
	flag.String("delete", "", "Delete a snitch")
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
	flag.String("update", "", "Update a snitch")
	flag.Bool("verbose", false, "Increase verbosity")
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

	if viper.GetString("snitch") == "" {
		snitch = viper.GetString("defaultsnitch")
	} else {
		snitch = viper.GetString("snitch")
	}

	if viper.GetString("alert") != "" {
		if !checkAlertType(strings.ToLower(viper.GetString("alert"))) {
			fmt.Println("ERROR: Invalid Alert Type", strings.ToLower(viper.GetString("alert")), ". Please choose either \"basic\" or \"smart\"")
			os.Exit(1)
		}
	} else {
		fmt.Println("init: alert check")
	}

	if viper.GetString("interval") != "" {
		if !checkInterval(strings.ToLower(viper.GetString("interval"))) {
			fmt.Println("ERROR: Invalid Interval", strings.ToLower(viper.GetString("interval")), ". Please choose either \"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
			os.Exit(1)
		} else {
			fmt.Println("init: interval check")
		}

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

	if viper.GetString("update") != "" {
		fmt.Println("Updating snitch")
		updateSnitch(viper.GetString("update"))
		os.Exit(0)
	}

	if viper.GetBool("create") {

		var mytags []string

		mytags = append(mytags, strings.Split(viper.GetString("tags"), ",")...)

		//newsnitch := newSnitch{Name: viper.GetString("name"), Interval: viper.GetString("interval"), AlertType: viper.GetString("alert"), Notes: viper.GetString("notes"), Tags: mytags}
		newsnitch := newSnitch{Name: viper.GetString("name"), Interval: strings.ToLower(viper.GetString("interval")), AlertType: strings.ToLower(viper.GetString("alert")), Notes: viper.GetString("notes"), Tags: mytags}

		//if !checkAlertType(newsnitch.AlertType) {
		//	fmt.Println("ERROR: Invalid Alert Type", newsnitch.AlertType, ". Please choose either \"basic\" or \"smart\"")
		//	os.Exit(1)
		//}

		//if !checkInterval(newsnitch.Interval) {
		//	fmt.Println("ERROR: Invalid Interval", strings.ToLower(viper.GetString("interval")), ". Please choose either \"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
		//	os.Exit(1)
		//}

		createSnitch(newsnitch)
		os.Exit(0)
	}

	if viper.GetString("delete") != "" {
		deleteSnitch(viper.GetString("delete"))
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

	if len(mysnitches) == 0 {
		fmt.Println("ERROR: No snitches found")
		os.Exit(1)
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
	if actionSnitch(snitch+"/pause", "POST", "") {
		fmt.Println("Successfully paused", snitch)
	} else {
		fmt.Println("ERROR: Cannot pause snitch", snitch)
	}
}

func unpauseSnitch(snitch string) {
	fmt.Println("Unpausing snitch:", snitch)
	sendSnitch(snitch)
}

func actionSnitch(action string, httpaction string, customheader string) bool {

	if len(apikey) == 0 {
		fmt.Println("ERROR: No API Key provided")
		os.Exit(1)
	}

	fmt.Println("running action:", action, " ", httpaction)
	snitch = url.QueryEscape(snitch)

	url := ""

	if !viper.GetBool("debug") {
		url = "https://api.deadmanssnitch.com/v1/snitches"
	} else {
		// nc -l 127.0.0.1 8888
		url = "http://localhost:8888/v1/snitches"
	}

	if len(httpaction) == 0 {
		url = url + "/" + action
	} else {
		switch {
		case strings.ToLower(httpaction) == "delete":
			fmt.Println("Changing action to delete")
			url = url + "/" + action
		default:
			fmt.Println("Some default")
		}
	}

	fmt.Println("url:", url)

	bytesaction := []byte(action)

	req, err := http.NewRequest(httpaction, url, bytes.NewBuffer(bytesaction))
	req.SetBasicAuth(apikey, "")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return false
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

	htmlData, _ := ioutil.ReadAll(resp.Body)

	if viper.GetBool("verbose") {
		fmt.Println("responsebody=", string(htmlData))
		fmt.Printf("code=%d  text=%s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
		switch {
		case resp.StatusCode >= 100 && resp.StatusCode <= 199:
			fmt.Println("Informational:", resp.StatusCode)
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			fmt.Println("Success:", resp.StatusCode)
		case resp.StatusCode >= 300 && resp.StatusCode <= 399:
			fmt.Println("Redirection:", resp.StatusCode)
		case resp.StatusCode >= 400 && resp.StatusCode <= 499:
			fmt.Println("Client Errors:", resp.StatusCode)
		case resp.StatusCode >= 500 && resp.StatusCode <= 599:
			fmt.Println("Server Error:", resp.StatusCode)
		default:
			fmt.Println("StatusCode:", resp.StatusCode)
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		fmt.Println("Success")
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		var errorresponse dmsResp
		json.Unmarshal(htmlData, &errorresponse)
		fmt.Printf("ERROR: %s/%s  MESSAGE:\"%s\"\n", http.StatusText(resp.StatusCode), errorresponse.Type, errorresponse.Error)
		return false
	}

	defer resp.Body.Close()

	return true
}

func createSnitch(newsnitch newSnitch) {
	fmt.Println("Creating snitch")

	jsonsnitch, _ := json.Marshal(newsnitch)
	fmt.Println(string(jsonsnitch))

	fmt.Println("--------------\n", newsnitch, "\n--------------")

	if len(newsnitch.Name) == 0 {
		fmt.Println("snitch name blank")
		os.Exit(1)
	}

	if len(newsnitch.Interval) == 0 {
		fmt.Println("interval blank")
		os.Exit(1)
	}

	newsnitch.Interval = strings.ToLower(viper.GetString("interval"))
	newsnitch.AlertType = strings.ToLower(viper.GetString("alert"))

	//======
	// validation now occuring in func init
	//if checkInterval(viper.GetString("interval")) {
	//	newsnitch.Interval = strings.ToLower(viper.GetString("interval"))
	//} else {
	//	fmt.Println("ERROR: Invalid Interval", strings.ToLower(viper.GetString("interval")), ". Please choose either \"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
	//	os.Exit(1)
	//}

	//if viper.GetString("alert") != foundSnitch.AlertType {
	//	fmt.Println(foundSnitch.AlertType, "->", viper.GetString("alert"))
	//	if checkAlertType(viper.GetString("alert")) {
	//		newsnitch.AlertType = strings.ToLower(viper.GetString("alert"))
	//	} else {
	//		fmt.Println("ERROR: Invalid Alert Type", strings.ToLower(viper.GetString("alert")), ". Please choose either \"basic\" or \"smart\"")
	//		os.Exit(1)
	//	}
	//}
	//======

	// check if existing snitch exists
	if !existSnitch(newsnitch) {
		fmt.Printf("Snitch %s already exists\n")
	} else {
		fmt.Println("creating snitch")
		//============
		if actionSnitch(string(jsonsnitch), "POST", "application/json") {
			fmt.Println("Successfully created snitch")
		} else {
			fmt.Println("ERROR: Cannot create snitch", newsnitch.Name)
		}
		//===========

	}

}

func deleteSnitch(snitchid string) {
	var delSnitch oneSnitch
	delSnitch.Name = strings.ToLower(snitchid)
	//if existSnitch(delsnitch) {
	if true {
		fmt.Println("Deleting snitch:", snitchid)
		if actionSnitch(snitchid, "DELETE", "") {
			fmt.Println("Successfully deleted snitch", snitchid)
		} else {
			fmt.Println("ERROR: Cannot delete snitch", snitchid)
		}
	} else {
		fmt.Printf("ERROR: Snitch %s not found\n", snitch)
	}
}

func existSnitch(snitch newSnitch) bool {
	fmt.Println("checking existence of snitch:", snitch.Name)
	return true
}

func updateSnitch(snitchtoken string) {
	fmt.Println("updateSnitch function running:", snitchtoken)

	if len(apikey) == 0 {
		fmt.Println("ERROR: No API Key provided")
		os.Exit(1)
	}

	snitchtoken = url.QueryEscape(snitchtoken)
	url := fmt.Sprintf("https://api.deadmanssnitch.com/v1/snitches/%s", snitchtoken)

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

	var foundSnitch oneSnitch
	if err := json.NewDecoder(resp.Body).Decode(&foundSnitch); err != nil {
		log.Println(err)
	}

	fmt.Println("SINGLE SNITCH FOUND:", foundSnitch)

	var updatesnitch udSnitch

	if viper.GetString("name") != foundSnitch.Name {
		fmt.Println(foundSnitch.Name, "->", viper.GetString("name"))
		updatesnitch.Name = viper.GetString("name")
	}

	//if viper.GetString("notes") != foundSnitch.Notes {
	//	fmt.Println(foundSnitch.Notes, "->", viper.GetString("notes"))
	//	newSnitch.Notes = viper.GetString("notes")
	//}

	if viper.GetString("notes") != foundSnitch.Notes {
		fmt.Println(foundSnitch.Notes, "->", viper.GetString("notes"))
		updatesnitch.Notes = viper.GetString("notes")
	}

	if viper.GetString("interval") != foundSnitch.Interval {
		fmt.Println(foundSnitch.Interval, "->", viper.GetString("interval"))

		if checkInterval(viper.GetString("interval")) {
			updatesnitch.Interval = viper.GetString("interval")
		} else {
			fmt.Println("ERROR: Invalid Interval", strings.ToLower(viper.GetString("interval")), ". Please choose either \"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
			os.Exit(1)
		}
	}

	if viper.GetString("alert") != foundSnitch.AlertType {
		fmt.Println(foundSnitch.AlertType, "->", viper.GetString("alert"))
		if checkAlertType(viper.GetString("alert")) {
			updatesnitch.AlertType = strings.ToLower(viper.GetString("alert"))
		} else {
			fmt.Println("ERROR: Invalid Alert Type", strings.ToLower(viper.GetString("alert")), ". Please choose either \"basic\" or \"smart\"")
			os.Exit(1)
		}
	}

	jsonudsnitch, err := json.Marshal(updatesnitch)

	if err != nil {
		fmt.Println("ERROR: Cannot convert to json")
		os.Exit(1)
	}

	fmt.Println("OLD:", foundSnitch)
	fmt.Println("NEW:", updatesnitch)
	fmt.Println("NEW JSON:", jsonudsnitch)
	os.Exit(0)

	// generate update json
	// actionSnitch(generated json)
}

func checkAlertType(alerttype string) bool {
	switch strings.ToLower(alerttype) {
	case "basic", "smart":
		fmt.Println("valid alert type:", alerttype)
		return true
	default:
		fmt.Println("invalid alert type:", alerttype)
		return false
	}
}

func checkInterval(interval string) bool {
	switch strings.ToLower(interval) {
	case "15_minute", "30_minute", "hourly", "daily", "weekly", "monthly":
		fmt.Println("valid interval:", interval)
		return true
	default:
		fmt.Println("invalid interval:", interval)
		return false
	}
}

func displayHelp() {
	helpmessage := `
snitchit

  --alert [type]                     Alert type: "basic" or "smart"
  --apikey [api key]                 Deadmanssnitch.com API key
  --config [config file]             Configuration file, default = config.yaml
  --create                           Create snitch, requires --name and --interval, optional --tags & --notes
  --debug                            Enable debug mode
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
  --verbose                          Increase verbosity
  --version                          Version
`
	fmt.Printf("%s", helpmessage)
}
