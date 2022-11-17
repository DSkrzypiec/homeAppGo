# v0.1.2
    * If 2FA is enabled, Telegram messages will be sent after insertions into
      database (counters state, financial transactions and documents)

# v0.1.1
    * Extend configuration (port)
    * Embed application version and git commit SHA into footer on /home
    * Add ./build.sh to build application (`go generate` is not needed before `go build`)
    * Initialized README

# v0.1.0
    * First deployed production version
    * Home database summary on /home page
    * Financial summary
    * Financial uploader for PKO Bank XML transactions history files
    * Documents listing
    * Uploading new documents
    * Possible 2FA via Telegram channel
