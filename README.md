# Snitchit

A simple command line tool that can be used to:
- raise snitches (check in) to [Deadmanssnitch.com](https://deadmanssnitch.com)
- create a snitch
- read a snitch
- update a snitch
- delete a snitch
- list snitches
- check status of snitches
- pause and unpause snitches

Typically used in cronjob to send snitch messages, but useful for self registration of snitches in a cloud environment. 


## Command line options
```
  --alert [type]                     Alert type: "basic" or "smart"
  --apikey [api key]                 Deadmanssnitch.com API Key
  --create                           Create snitch, requires --name and --interval, optional --tags & --notes
  --config [config file]             Configuration file, default = config.yaml
  --debug                            Enable debug mode
  --displayconfig                    Display configuration
  --help                             Display help
  --message [message to send]        Message to send, default = "2006-01-02T15:04:05Z07:00" format
  --path [path to config file]       Path to configuration file, default = current directory
  --pause [snitch]                   Pauses a snitch
  --show                             Display all snitches
  --show --snitch [snitch]           Show details for a specific snitch
  --silent                           Be silent
  --snitch [snitch]                  Snitch to use, default = defaultsnitch from config.yaml
  --tags [tags]                      Tags separated by commas, "tag1,tag2,tag3"
  --unpause [snitch]                 Unpause a snitch
  --update [snitch]                  Update a snitch, can be used with --name, --interval, --tags & --notes
  --verbose                          Increase verbosity
  --version                          Version
```

## Configuration file
```
apikey: my-api-key
defaultsnitch: 10ffbf9437f6
silent: false
snitches:
- 10ffbf9437f6
- snitch2
- snitch3
```
