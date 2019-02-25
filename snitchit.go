package main

// snitchit.go

import (
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
	"time"
)

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

}

func displayHelp() {
	helpmessage := `
snitchit

  --config [config file]             Configuration file, default = config.yaml
  --help                             Display help
  --message [messgage to send]       Message to send, default = "2006-01-02T15:04:05Z07:00" format
  --path [path to config file]       Path to configuration file, default = current directory
  --show                             Display snitches
  --snitch [snitch]                  Snitch to use, default = defaultsnitch from config.yaml
  --version                          Version
`
	fmt.Printf("%s", helpmessage)
}
