# I See U

A tool that helps you monitor a collection of websites using various metrics.

<p align="center">
  <img src="https://static.wikia.nocookie.net/dumbledoresarmyroleplay/images/0/09/Vigilance.gif/revision/latest?cb=20180516193632" />
</p>

> "Concealed within his fortress, the lord of Mordor sees all. His gaze pierces cloud, shadow, earth, and flesh. You know of what I speak, Gandalf: a great Eye, lidless, wreathed in flame."

## Preview

![Website Monitor Demo](demo.gif)

## Quickstart

    iseeu google.com 2 github.com 3

This command starts monitoring of the websites:

* `google.com` every `2sec`
* `github.com` every `3sec`

## Project Description

### Stats

* checks the different websites with their corresponding check intervals.
* Every 2s, display the stats for the past 10 seconds for each website
* Every 10s, displays the stats for the past minute for each website

### Alerts

* When a website availability is below 80% for the past 2 minutes, add a message saying that "Website {website} is down. availability={availability}, time={time}"
* When availability recovers for each website.
* We can scroll through alerts using keyboard arrows.

## Install

### Precompiled binaries

Precompiled binaries for released versions are available in for each platform in the [packages section](https://github.com/NouamaneTazi/website-monitor/releases/)

### Building from source

To build Website Monitor from source code, first ensure that you have a working
Go environment with [version 1.15 or greater installed](https://golang.org/doc/install).

You can directly use the `go` tool to download and install the `website-monitor` tool into your `GOPATH`:

    go get -v github.com/NouamaneTazi/website-monitor

## Usage

```bash
$ website-monitor
Usage: iseeu [OPTIONS] URL1 POLLING_INTERVAL1 URL2 POLLING_INTERVAL2

Example: iseeu -crit 0.3 -sui 1s google.com 2 http://google.fr 1

OPTIONS:
  -alertint WebsiteAlertInterval
        Shows alert if website is down for WebsiteAlertInterval minutes (default 10s)
  -crit float
        Availability of websites below which we show an alert (default 0.8)
  -lstats LongStatsHistoryInterval
        Long refreshes show stats for past LongStatsHistoryInterval minutes (default 1m0s)
  -lui duration
        Long refreshing UI interval (in seconds) (default 10s)
  -sstats ShortStatsHistoryInterval
        Short refreshes show stats for past ShortStatsHistoryInterval minutes (default 10s)
  -sui duration
        Short refreshing UI interval (in seconds) (default 2s)
```

## Testing

To run tests run the command

```bash
go test ./...
```
