# 0.5.3
    * Go back to 100 MB document file limit. Current EC2 is too small for that.

# 0.5.2
    * Increased document file size from 100 MB to 275 MB

# 0.5.1
    * Changed column chart to bar chart in Financial Explorer

# 0.5.0
    * First working version with Finance Explorer - monthly aggregated chart
      for filtered transactions.

# v0.4.2
    * Fix hardcoded 2022 year in financial summary

# v0.4.1
    * Add flags for logger

# v0.4.0
    * Add e-books and paper books info summary on the /home page

# v0.3.0
    * Add user session timer on /home page
    * Add functionality of user session prolongation by click onto user session
      timer on /home page

# v0.2.2
    * Extend fields (in the form and database) for books - category and language
    * Add sending Telegram message after new book entry was inserted

# v0.2.1
    * Make dynamic number of columns in table of documents based on screen size
    * Make dynamic number of columns in table of books based on screen size

# v0.2.0
    * Add (e)books page
    * Books (just info about a book) or e-book (info + e-book file) can be
      added
    * Books can be browsed and filtered
    * Books, in case of e-books, can be downloaded

# v0.1.4
    * Update monitoring page view counts by putting unregistered endpoint paths
      into single category, to make published statistics more readable (there
      are a lots of bots and crawlers out there!)
    * JWT signing key is now randomly generated

# v0.1.3
    * Add basic monitoring regarding page view counts. If Telegram 2FA is
      enabled monitoring messages would be send onto Telegram channel.

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
