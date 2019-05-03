# Snitchit

A simple command line tool that can be used to raise snitches (check in) to [Deadmanssnitch.com](https://deadmanssnitch.com), list snitches, check status of snitches, and pause/unpause snitches.

Typically used in cronjob to send snitch messages, but useful for self registration of snitches in a cloud environment. 


## Command line options
```
  --config [config file]             Configuration file, default = config.yaml
  --displayconfig                    Display configuration
  --help                             Display help
  --message [messgage to send]       Message to send, default = "2006-01-02T15:04:05Z07:00" format
  --path [path to config file]       Path to configuration file, default = current directory
  --pause [snitch]                   Pauses a snitch
  --show                             Display all snitches
  --show --snitch [snitch]           Show details for a specific snitch
  --silent                           Be silent
  --snitch [snitch]                  Snitch to use, default = defaultsnitch from config.yaml
  --unpause [snitch]                 Unpause a snitch
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
