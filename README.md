# Cinnabot :robot: 
[![Build Status](https://travis-ci.com/usdevs/cinnabot.svg)](https://travis-ci.com/usdevs/cinnabot/)
[![Telegram](https://img.shields.io/badge/telegram-ready-brightgreen.svg)](https://t.me/cinnabot)

Telegram Bot for Cinnamon College. [Telegram](https://t.me/cinnabot)

## Features:
- Show NUS Internal Shuttle Bus Timings :oncoming_bus: `/nusbus`
- Show Singapore Public Bus Timings :oncoming_bus: `/publicbus`
- Check facilities booking/events in Cinnamon College :school: `/spaces`
- 2h Weather Forecast based on your location. :sunny: :umbrella: `/weather`
- Lost? View maps of various parts of NUS :school: `/map`
- Check laundry machine availability in Cinnamon College :shirt: `/laundry`

Got a feature to suggest? :bulb:
Bug to report? :bug:
You are welcome to file an issue [here](https://github.com/usdevs/cinnabot/issues).

## Documentation quick links
- [Setting up the environment](#setting-up-the-environment)
- [Testing cinnabot locally](#testing-cinnabot-locally)

## Setting up the environment :earth_asia:

### 0. Install essential packages

We will first need to install git and golang. First install git using your favourite
package manager:

#### Ubuntu/Debian derivatives using APT
```bash
sudo apt install git
```

#### MacOS using Homebrew :beer:
```bash
brew install git
```

Then, install Golang **version 1.15**. You can either [download it from the official website](https://golang.org/dl/) or download it using a package manager (remember to check the go version provided by the package manager).

For Linux systems, you can follow the steps if you downloaded go from the website:

```bash
cd Downloads # or wherever you downloaded the file
sudo tar -C /usr/local -xzf go1.15.2.linux-amd64.tar.gz
# change the lines below to the shell you are using
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
```

To verify that Golang is installed properly:
```bash
$ which go # only for Mac/Linux
/usr/local/go/bin/go # it's okay for this to differ, it just shouldn't be blank
$ go version
go version go1.15.2 linux/amd64 # version, operating system/architecture
```

### 1. Clone the cinnabot repository

Go the the directory where you want to store the cinnabot code, and run this command to download the cinnabot repository into a new directory called 'cinnabot': 
```bash
git clone https://github.com/usdevs/cinnabot.git 
# or git@github.com:usdevs/cinnabot.git if you are using a ssh key
```
Voila! Now we have cinnabot on our machine. Ready to go!! :tada:


## Testing Cinnabot locally
Overview:
1. [Register for an API Token with `BotFather`](#1-register-for-an-api-token-with-botfather)
2. [Running a test bot on Telegram](#2-running-a-test-bot-on-telegram)

**All instructions below assume you are at the cinnabot root path, unless stated otherwise.**

### 1. Register for an API Token with Botfather
Ask for the blessings of the Botfather [here](https://t.me/botfather), as you register for one of the bots.
You will be provided, with honor, an API token where you should put into `main/config.json`.

In other words, click on the link and choose a Telegram handle for your bot, which ends with `...bot` and is not taken yet. Once you are done, you will be provided with the API Token.

Then, we create our `config.json` file using the example file `config.json.example`, and replacing the dummy API Token with the one we just registered.
```bash
cd main
cp config.json.example config.json
```

Fire up your favourite text editor and replace the dummy string in `config.json` with your API Token as a string.

### 2. Running a test bot on Telegram
```bash
cd main
go run main.go
```

And start testing! Fire up your favourite Telegram client, and find the bot by the name you registered it with. You can now test all cinnabot functionalities on your testbot.

When you are done, press <kbd>Ctrl</kbd>+<kbd>C</kbd> on your terminal to end testing.


