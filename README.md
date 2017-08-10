# Bruno Bot

Bruno Bot generates Discord notifications when pro Dota 2 games are being played.  
It can be configured to notify on any professional game or only teams on a whitelist.

Built using: `Go 1.7.5`

## Installing

To install this package, assuming [$GOPATH is setup](https://golang.org/doc/install), golang is installed and working, run:

```bash
$ go install github.com/jasonodonnell/brunobot
```

## Setup

Bruno Bot will only notify on games less than a minute old (using a sliding window)
polling more often is a waste of resources).

Prior to running, a Discord Webhook should be created for the room you want to send 
notifications to.  See the `Webhooks` menu in `Server Settings` for your Discord.

Next, configure `config.json` with the parameters specified.

| Parameter   | Description                                                            |
|-------------|------------------------------------------------------------------------|
| `Path`      | This is where Brunobot stores which games have been alerted.           |
| `Teams`     | This is the array of teams for filtering notifications.                |
| `Webhook`   | URL for the Webhook generated by Discord for your server.              |
| `Whitelist` | If true, enables team filter.  If false, all matches will be reported. |


With the `config.json` file configured, `brunobot` can now be run:

```bash
$ brunobot --config=/path/to/config.json
```

## Cron

To avoid duplicated notifications, Bruno Bot will keep track of the games it has 
already notified on.  It's suggested that Brunobot be run via `cron` every minute 
and the tracker file be rotated weekly

```bash
* * * * * /usr/local/bin/brunobot --config=/path/to/config.json
0 0 * * 7 > ~/.urls
```
