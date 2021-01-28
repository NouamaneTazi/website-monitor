# Website Monitor in Go

A tool that helps you monitor a collection of websites using various metrics.
### Stats:
* checks the different websites with their corresponding check intervals.
* Every 2s, display the stats for the past 10 seconds for each website
* Every 10s, displays the stats for the past minute for each website

### Alerts:
* When a website availability is below 80% for the past 2 minutes, add a message saying that "Website {website} is down. availability={availability}, time={time}"
* When availability recovers for each website.
* We can scroll through alerts using keyboard arrows.

![Website Monitor Demo](https://recordit.co/uCb22IUQ4G.gif)
## Quickstart


    $ website-monitor google.com 2 jeux.fr 1 github.com 3

This command starts monitoring of the websites:
- `google.com` every `2sec`
- `jeux.fr` every `1sec`
- `github.com` every `3sec`


## Install

### Precompiled binaries
Precompiled binaries for released versions are available in for each platform in the [packages section](https://github.com/NouamaneTazi/website-monitor/releases/)

### Building from source
To build Website Monitor from source code, first ensure that you have a working
Go environment with [version 1.14 or greater installed](https://golang.org/doc/install).

You can directly use the `go` tool to download and install the `website-monitor` tool into your `GOPATH`:

    $ go get -v github.com/NouamaneTazi/website-monitor


## Advanced Usage
    $ website-monitor
    $ Usage: website-monitor [OPTIONS] URL1 POLLING_INTERVAL1 URL2 POLLING_INTERVAL2

    OPTIONS:
    -lstats duration
            Long history interval (in minutes) (default 1m0s)
    -lui duration
            Long refreshing UI interval (in seconds) (default 10s)
    -sstats duration
            Short history interval (in minutes) (default 10s)
    -sui duration
            Short refreshing UI interval (in seconds) (default 2s)


## Improvements
* Better http client configuration to make tracing more failproof.
* Prettier widgets for stats
* Ability to add and omit monitored URLs while the app is running
* Ability to export logs
