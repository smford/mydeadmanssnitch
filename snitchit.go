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
	"time"
)

const appversion = 0.01

var (
	configFile string
	message    string
	silent     bool
	snitch     string
)

func init() {
	flag.Bool("help", false, "Display help")
	tempmessage := flag.String("message", "", "Mesage to display, default = \"Thursday, 21-Feb-19 19:15:09 GMT\" format")
	flag.Bool("version", false, "Display version")
	configFile := flag.String("config", "", "name of configuration file")

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

	viper.AddConfigPath(".")

	if *tempmessage == "" {
		currenttime := time.Now().Format(time.RFC850)
		message = currenttime
	} else {
		message = *tempmessage
	}

	if *configFile == "" {
		viper.SetConfigName("config")
	} else {
		viper.SetConfigName(*configFile)
	}

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("error loading config")
		os.Exit(1)
	}
	snitch = viper.GetString("snitch")
	silent = viper.GetBool("silent")
	fmt.Println("snitch=", snitch)
	fmt.Println("silent=", silent)
}

func main() {
	displayConfig()
	client := &http.Client{}
	client.Timeout = time.Second * 15

	//currenttime := time.Now().Format(time.RFC850)

	//fmt.Println(currenttime)

	if !viper.GetBool("silent") {
		fmt.Println("Message:", message)
	}

	uri := fmt.Sprintf("https://nosnch.in/%s", snitch)
	data := url.Values{
		//"m": []string{currenttime},
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

func displayHelp() {
	helpmessage := `
some temp help message
`
	fmt.Printf("%s", helpmessage)
}
