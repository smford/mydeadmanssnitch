package main

// snitchit.go

import (
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

const appversion = 0.01

var (
	defaultsnitch string
	message       string
	showsnitches  bool
	silent        bool
	snitch        string
)

func init() {
	flag.Bool("help", false, "Display help")
	tempmessage := flag.String("message", "", "Mesage to send, default = \"2006-01-02T15:04:05Z07:00\" format")
	flag.Bool("version", false, "Version")
	configFile := flag.String("config", "config.yaml", "Configuration file, default = config.yaml")
	configPath := flag.String("path", ".", "Path to configuration file, default = current directory")
	showsnitches = *flag.Bool("show", false, "Show snitches")
	flag.String("snitch", "", "Snitch to use")

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
	fmt.Printf("Loading: %s/%s\n", *configPath, *configFile)

	viper.SetConfigName(config)
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("ERROR loading configuration file: %s/%s\n", *configPath, *configFile)
		os.Exit(1)
	}

	fmt.Println("viper snitch:", viper.GetString("snitch"))

	if viper.GetString("snitch") == "" {
		snitch = viper.GetString("defaultsnitch")
	} else {
		snitch = viper.GetString("snitch")
	}

	silent = viper.GetBool("silent")
	fmt.Println("snitch=", snitch)
	fmt.Println("silent=", silent)
}

func main() {
	displayConfig()

	if !viper.GetBool("silent") {
		fmt.Println("Message:", message)
	}

	if viper.GetBool("show") {
		displaySnitch(snitch)
		os.Exit(0)
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
	fmt.Printf("sending snitch: https://nosnch.in/%s\n", sendsnitch)
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
	if !viper.GetBool("silent") {
		fmt.Printf("response=%s\n", snitchresponse)
	}
}

func displaySnitch(snitch string) {
	snitch = url.QueryEscape(snitch)
	fmt.Printf("displaying snitches: https://api.deadmanssnitch.com/v1/snitches/%s\n", snitch)
	url := fmt.Sprintf("https://api.deadmanssnitch.com/v1/snitches/%s", snitch)

	apikey := viper.GetString("apikey")

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

func displayHelp() {
	helpmessage := `
snitchit

  --config [config file]             Configuration file, default = config.yaml
  --help                             Display help
  --message [messgage to send]       Message to send, default = "2006-01-02T15:04:05Z07:00" format
  --path [path to config file]       Path to configuration file, default = current directory
  --show                             Display all snitches
  --show --snitch [snitch]           Show details for a specific snitch
  --snitch [snitch]                  Snitch to use, default = defaultsnitch from config.yaml
  --version                          Version
`
	fmt.Printf("%s", helpmessage)
}
