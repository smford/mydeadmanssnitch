package main

// snitchit.go

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/go-cmp/cmp"
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

const appversion = "0.0.9"

var (
	apikey        string
	defaultsnitch string
	interval      string
	message       string
	name          string
	showsnitches  bool
	silent        bool
	snitch        string
	verbose       bool
)

func init() {
	flag.Bool("as1", true, "")
	flag.Bool("as2", false, "")
	flag.String("alert", "basic", "Alert type: \"basic\" or \"smart\"")
	flag.String("apikey", "", "Deadmanssnitch.com API Key")
	configFile := flag.String("config", "config.yaml", "Configuration file, default = config.yaml")
	flag.Bool("create", false, "Create snitch, requires --name and --interval, optional --tags & --notes")
	flag.String("delete", "", "Delete a snitch")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help")
	flag.String("interval", "", "\"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
	tempmessage := flag.String("message", "", "Mesage to send, default = \"2006-01-02T15:04:05Z07:00\" format")
	flag.String("name", "", "Name of snitch")
	flag.String("notes", "", "Notes")
	configPath := flag.String("path", ".", "Path to configuration file, default = current directory")
	flag.String("pause", "", "Pause a snitch")
	flag.String("plan", "free", "Plan type: \"free\", \"small\", \"medium\" or \"large\", default = free")
	showsnitches = *flag.Bool("show", false, "Show snitches")
	flag.Bool("silent", false, "Be silent")
	flag.String("snitch", "", "Snitch to use")
	flag.String("tags", "", "Tags separated by commas, \"tag1,tag2,tag3\"")
	flag.String("unpause", "", "Unpause a snitch")
	flag.String("update", "", "Update a snitch, can be used with --name, --interval, --tags & --notes")
	flag.Bool("verbose", false, "Be verbose")
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
		}
	}

	if !checkPlan(viper.GetString("plan"), viper.GetString("alert"), viper.GetString("interval")) {
		fmt.Println("ERROR: Basic Alerts are available for any snitch. Smart Alerts are available for hourly, daily, weekly, and monthly interval snitches on the Surveillance Van plan, and for monthly interval snitches on all other plans.")
		os.Exit(1)
	}

	apikey = viper.GetString("apikey")
	silent = viper.GetBool("silent")
	verbose = viper.GetBool("verbose")

	if len(apikey) == 0 {
		fmt.Println("ERROR: No API Key provided")
		os.Exit(1)
	}

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

		newsnitch := newSnitch{Name: viper.GetString("name"), Interval: strings.ToLower(viper.GetString("interval")), AlertType: strings.ToLower(viper.GetString("alert")), Notes: viper.GetString("notes"), Tags: mytags}

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
	if verbose {
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

	if verbose {
		fmt.Println("Response Code:", resp.StatusCode, "Response Text:", http.StatusText(resp.StatusCode), "Message:", snitchresponse)
	}

	if !silent {
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			fmt.Println("Success")
		}
	}

}

func displaySnitch(snitch string) {

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
		log.Fatal(err)
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
	// if actionSnitch("/pause", "POST", "", snitch) {
	if actionSnitch2("pause", snitch, "") {
		fmt.Println("Successfully paused", snitch)
	} else {
		fmt.Println("ERROR: Cannot pause snitch", snitch)
	}
}

func unpauseSnitch(snitch string) {
	fmt.Println("Unpausing snitch:", snitch)
	sendSnitch(snitch)
}

func actionSnitch(action string, httpaction string, customheader string, snitchid string) bool {

	snitch = url.QueryEscape(snitch)
	url := "https://api.deadmanssnitch.com/v1/snitches"
	fmt.Println("httpaction=", httpaction)
	switch strings.ToLower(httpaction) {
	case "delete", "patch":
		url = url + "/" + snitchid
	case "post":
		url = url + "/"
	default:
		// maybe throw an error if not http action provided rather than have a fall back
		url = url + "/" + action
	}

	bytesaction := []byte(action)
	req, err := http.NewRequest(httpaction, url, bytes.NewBuffer(bytesaction))
	req.SetBasicAuth(apikey, "")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return false
	}

	if len(customheader) != 0 {
		req.Header.Add("Content-Type", customheader)
	}

	client := &http.Client{}
	client.Timeout = time.Second * 15

	resp, err := client.Do(req)
	htmlData, _ := ioutil.ReadAll(resp.Body)

	if verbose {
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			fmt.Println("Success")
		}
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

	if verbose {
		fmt.Println("JSON Payload:", string(jsonsnitch))
	}

	if len(newsnitch.Name) == 0 {
		fmt.Println("ERROR: name cannot be blank")
		os.Exit(1)
	}

	if len(newsnitch.Interval) == 0 {
		fmt.Println("ERROR: interval cannot be blank")
		os.Exit(1)
	}

	newsnitch.Interval = strings.ToLower(viper.GetString("interval"))
	newsnitch.AlertType = strings.ToLower(viper.GetString("alert"))

	// check if existing snitch exists
	if !existSnitch(newsnitch) {
		fmt.Printf("Snitch %s already exists\n")
	} else {
		fmt.Println("****outer jsonsnitch:", jsonsnitch)
		//if actionSnitch(string(jsonsnitch), "POST", "application/json", "") {
		if actionSnitch2("create", "", string(jsonsnitch)) { // error here
			fmt.Println("****inner jsonsnitch:", string(jsonsnitch))
			fmt.Println("Successfully created snitch")
		} else {
			fmt.Println("ERROR: Cannot create snitch", newsnitch.Name)
		}
	}

}

func deleteSnitch(snitchid string) {
	var delSnitch oneSnitch
	delSnitch.Name = strings.ToLower(snitchid)
	//if existSnitch(delsnitch) {
	if true {
		fmt.Println("Deleting snitch:", snitchid)
		// if actionSnitch("", "DELETE", "", snitchid) {
		if actionSnitch2("delete", snitchid, "") {
			fmt.Println("Successfully deleted snitch", snitchid)
		} else {
			fmt.Println("ERROR: Cannot delete snitch", snitchid)
		}
	} else {
		fmt.Printf("ERROR: Snitch %s not found\n", snitch)
	}
}

func existSnitch(snitch newSnitch) bool {
	fmt.Println("REFACTOR: checking existence of snitch:", snitch.Name)
	return true
}

func updateSnitch(snitchtoken string) {
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

	if foundSnitch.Token != snitchtoken {
		fmt.Println("ERROR: No snitch found matching:", snitchtoken)
		os.Exit(1)
	}

	var updatesnitch udSnitch

	if viper.GetString("name") != foundSnitch.Name {
		if verbose {
			fmt.Println("Name:", foundSnitch.Name, "->", viper.GetString("name"))
		}
		updatesnitch.Name = viper.GetString("name")
	}

	var newtags []string
	newtags = append(newtags, strings.Split(viper.GetString("tags"), ",")...)

	fmt.Println("TAGS FOUND:", foundSnitch.Tags)

	if !cmp.Equal(foundSnitch.Tags, newtags) {
		if verbose {
			fmt.Println("Tags:", foundSnitch.Tags, "->", newtags)
		}
		updatesnitch.Tags = newtags
	}

	if viper.GetString("notes") != foundSnitch.Notes {
		if verbose {
			fmt.Println("Notes:", foundSnitch.Notes, "->", viper.GetString("notes"))
		}
		updatesnitch.Notes = viper.GetString("notes")
	}

	if viper.GetString("interval") == "" {
		updatesnitch.Interval = foundSnitch.Interval
	} else {
		if viper.GetString("interval") != foundSnitch.Interval {
			if verbose {
				fmt.Println("Interval:", foundSnitch.Interval, "->", viper.GetString("interval"))
			}

			if checkInterval(viper.GetString("interval")) {
				updatesnitch.Interval = viper.GetString("interval")
			} else {
				fmt.Println("ERROR: Invalid Interval", strings.ToLower(viper.GetString("interval")), ". Please choose either \"15_minute\", \"30_minute\", \"hourly\", \"daily\", \"weekly\", or \"monthly\"")
				os.Exit(1)
			}
		} else {
			// no change to interval
			updatesnitch.Interval = foundSnitch.Interval
		}
	}

	if viper.GetString("alert") != foundSnitch.AlertType {
		if verbose {
			fmt.Println("Alert:", foundSnitch.AlertType, "->", viper.GetString("alert"))
		}
		if checkAlertType(viper.GetString("alert")) {
			updatesnitch.AlertType = strings.ToLower(viper.GetString("alert"))
		} else {
			fmt.Println("ERROR: Invalid Alert Type", strings.ToLower(viper.GetString("alert")), ". Please choose either \"basic\" or \"smart\"")
			os.Exit(1)
		}
	} else {
		updatesnitch.AlertType = foundSnitch.AlertType
	}

	// generate update json payload
	jsonudsnitch, err := json.Marshal(updatesnitch)

	if err != nil {
		fmt.Println("ERROR: Cannot convert to json")
		os.Exit(1)
	}

	if verbose {
		fmt.Println("Current Snitch:", foundSnitch)
		fmt.Println("    New Snitch:", updatesnitch)
	}

	//============
	//fmt.Println("========================================================")
	//if viper.GetBool("as1") {
	//	actionSnitch(string(jsonudsnitch), "PATCH", "application/json", snitchtoken)
	//}

	//if viper.GetBool("as2") {
	//	fmt.Println("========================================================")
	//	actionSnitch2("update", snitchtoken, string(jsonudsnitch))
	//}
	//os.Exit(1)
	//============

	//if actionSnitch(string(jsonudsnitch), "PATCH", "application/json", snitchtoken) {
	if actionSnitch2("update", snitchtoken, string(jsonudsnitch)) {
		fmt.Println("Successfully updated snitch")
		os.Exit(0)
	} else {
		fmt.Println("ERROR: Cannot update snitch", snitchtoken)
		os.Exit(1)
	}
}

func checkAlertType(alerttype string) bool {
	switch strings.ToLower(alerttype) {
	case "basic", "smart":
		return true
	default:
		return false
	}
}

func checkInterval(interval string) bool {
	switch strings.ToLower(interval) {
	case "15_minute", "30_minute", "hourly", "daily", "weekly", "monthly":
		return true
	default:
		return false
	}
}

func checkPlan(plan string, alert string, interval string) bool {
	if strings.ToLower(alert) == "basic" {
		// all plans allow basic snitches
		return true
	} else {
		if strings.ToLower(plan) == "free" {
			// free plans do not allow smart snitches
			return false
		}
		if (strings.ToLower(interval) == "15_minute") || (strings.ToLower(interval) == "30_minute") {
			return false
		}
		if strings.ToLower(plan) == "large" {
			// large plans can have smart snitches for hourly, daily, weekly or monthly
			return true
		}
		if strings.ToLower(interval) == "monthly" {
			// small, medium and large allow smart snitches for monthly
			return true
		} else {
			return false
		}
	}
}

func displayHelp() {
	helpmessage := `
snitchit

  --alert [type]                     Alert type: "basic" or "smart"
  --apikey [api key]                 Deadmanssnitch.com API key
  --config [config file]             Configuration file, default = config.yaml
  --create                           Create snitch, requires --name and --interval, optional --tags & --notes
  --displayconfig                    Display configuration
  --help                             Display help
  --interval [interval window]       "15_minute", "30_minute", "hourly", "daily", "weekly", or "monthly"
  --message [messgage to send]       Message to send, default = "2006-01-02T15:04:05Z07:00" format
  --name [name]                      Name of snitch
  --notes [notes]                    Notes for snitch
  --path [path to config file]       Path to configuration file, default = current directory
  --pause [snitch]                   Pauses a snitch
  --plan [plan type]                 Plan type: "free", "small", "medium" or "large", default = free
  --show                             Display all snitches
  --show --snitch [snitch]           Show details for a specific snitch
  --silent                           Be silent
  --snitch [snitch]                  Snitch to use, default = defaultsnitch from config.yaml
  --tags [tags]                      Tags separated by commas, "tag1,tag2,tag3"
  --unpause [snitch]                 Unpause a snitch
  --update [snitch]                  Update a snitch, can be used with --name, --interval, --tags & --notes
  --verbose                          Be verbose
  --version                          Version
`
	fmt.Printf("%s", helpmessage)
}

//======================================

func actionSnitch2(todo string, token string, jsonpayload string) bool {
	token = url.QueryEscape(token)
	url := "https://api.deadmanssnitch.com/v1/snitches"

	var httpaction string
	var header string

	fmt.Println("jsonstring:", jsonpayload)
	//fmt.Println("  jsonbyte:", []byte(jsonpayload))

	//----------
	switch strings.ToLower(todo) {
	case "create":
		httpaction = "POST"
		//header = "Content-Type: application/json"
		header = "application/json"
	case "read":
		httpaction = "POST"
	case "update":
		httpaction = "PATCH"
		//header = "Content-Type: application/json"
		header = "application/json"
		url = url + "/" + token
	case "delete":
		httpaction = "DELETE"
		url = url + "/" + token
	case "pause":
		httpaction = "POST"
		url = url + "/" + token + "/pause"
	default:
		fmt.Println("ERROR: todo is invalid")
		os.Exit(1)
	}
	//----------

	fmt.Println("Header:", header)
	fmt.Println("   URL:", url)
	fmt.Println("  TODO:", todo)
	fmt.Println("  JSON:", jsonpayload)
	fmt.Println("Action:", httpaction)

	bytesaction := []byte(jsonpayload)
	req, err := http.NewRequest(httpaction, url, bytes.NewBuffer(bytesaction))

	fmt.Println(" BYTES:", bytesaction)
	fmt.Printf("    CMD: req, err := http.NewRequest(%s, %s, %s)\n", httpaction, url, bytes.NewBuffer(bytesaction))

	req.SetBasicAuth(apikey, "")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return false
	}

	if len(header) != 0 {
		fmt.Println("header is not 0:", header)
		req.Header.Add("Content-Type", header)
	}

	client := &http.Client{}
	client.Timeout = time.Second * 15

	resp, err := client.Do(req)
	htmlData, _ := ioutil.ReadAll(resp.Body)

	if verbose {
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			fmt.Println("Success")
		}
	}

	//if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
	var errorresponse dmsResp
	json.Unmarshal(htmlData, &errorresponse)
	fmt.Printf("ERROR: %s/%s  MESSAGE:\"%s\"\n", http.StatusText(resp.StatusCode), errorresponse.Type, errorresponse.Error)
	//return false
	//}

	defer resp.Body.Close()

	return true
}
