## Running the program

The basic interface looks like this and there is no interaction possible at this time.

![Go Cryptotracker](https://i.gyazo.com/ff5f2a54d83063038613c70edef340e8.gif)

Please see the config.json.sample file for possible settings.

In order to run this download the latest zip file from the [releases page](https://github.com/markustenghamn/golang-cryptotracker/releases) and extract the contents in a folder. Rename the config.json.sample file to config.json and edit it as needed. Type `./golang-cryptotracker` from a terminal to run it making sure the config file is in the same folder.

This script uses the [CryptoCompare API](https://www.cryptocompare.com/api/) which is free to use, please limit the number of requests you do, an interval of 60 seconds should be more than enough.

## Build from source

Navigate to the project folder and type ´go run main.go´ in your terminal.

See config.json.sample on how to configure, rename config.json.sample to config.json before running.
