# cinnabot  [![Build Status](https://travis-ci.org/usdevs/cinnabot.svg)](https://travis-ci.org/usdevs/cinnabot/)
Telegram Bot for Cinnamon College

## Documentation quick links
- Setting up the environment
- Testing cinnabot locally

## Setting up the environment

0. Install essential packages

We will first need to install git and golang.

### Ubuntu/Debian derivatives using APT
```bash
sudo apt install git golang
```

### MacOS using Homebrew
```bash
brew install git golang
```

### Windows: WIP

1. Set the environment variables

All the Go related files will go under a folder called `GOPATH`. We first have to tell your computer where that is, by setting up this path in your environment variables. 

Follow the official guide [here](https://github.com/golang/go/wiki/SettingGOPATH).

For Linux and MacOS users, follow the section under *UNIX Systems*. If you are not sure which shell you are using, enter `$0` into your command line.

You can also follow the steps below:
### Bash (if you are not sure, you are most likely using BASH)
```bash
echo "export GOPATH=$HOME/go" >> ~/.bashrc
source ~/.bashrc
```

### Zsh
```bash
echo "export GOPATH=$HOME/go" >> ~/.zshrc
source ~/.zshrc
```

At this stage, you can verify that Golang and `GOPATH` are set up properly by:
```bash
$ go version
go version go1.10.4 linux/amd64 # version, OS/architecture
$ echo $GOPATH
/home/your_username/go # default GOPATH
```

2. Create a directory under GO
