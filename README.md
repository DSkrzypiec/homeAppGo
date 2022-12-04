# HomeApp

## Overview

HomeApp is an application for handling I/O for Home Database which are designed to contains personal and private
information. Motivations for creating this project can be read on my personal
[blog post](https://dskrzypiec.dev/home-db/).

If you want to check HomeApp without compiling the project locally, you can visit

* [https://homeappdev.dskrzypiec.dev](https://homeappdev.dskrzypiec.dev)
* and use credentials `testuser/password` to log in

## Build

On Linux or MacOS simply do

```
bash build.sh
```

On other platforms you can build directly using Go compiler:

```
go generate
go build
```

Running the HomeApp by `./homeApp`. As default will not include 2FA via Telegram and will use test SQLite database
(`test.db`).

### Credentials

For test database (`test.db`) there is single user:

```
login: testuser
password: password
```

### Flags

HomeApp supports the following options

* `-dbPath path` - path to Home Database
* `-port 8080` - port on which HomeApp will be listening
* `-telegram2fa` - if enabled, then HomeApp will use two-factor authentication (2FA) using Telegram channel. More details below.
* `-publishViewsAfter 300` - numbers of minutes after which endpoints view
      statistics will be published. If Telegram is configured, then it'll be sent
      over the Telegram channel. Otherwise just logged
* `-logDebug` - flag for enabling debug log level. If disabled, then "info" log level would be used
* `-logConsole` - flag for using `ConsoleWriter` within `zerolog`. Convenient for local development but is less efficient then standard writer.


### 2FA via Telegram

In case when 2FA via Telegram is enabled you have to provide the following environment variables:

```
export HOMEAPP_TELEGRAM_CHANNEL_ID=-100xxxxxxxx65
export HOMEAPP_TELEGRAM_BOT_TOKEN=11111111:AAAAAA-jjjjj-vTDSFSKDFndfkldsG
```

Values of those variables came from setting up [Telegram bot](https://core.telegram.org/bots/api) and dedicated Telegram
channel for the communication.


## High level design

**TODO**
