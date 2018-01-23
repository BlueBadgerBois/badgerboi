# Quote Server

Server that returns a random quote for a requested stock. This server is built to mimic the one used in the lab so that we are able to work on this outside of the lab.

```
go run quoteServer.go
```

## Input/Output Formats

Input and output are in CSV form. The input is the following:

```
StockSYM, UserID
```

The returned values are formatted like this:

```
Quote, SYM, UserID, Timestamp, Cryptokey

Quote:      range 0-1000
Sym:        Same as input but capitalized
UserID:     Same as input
Timestamp:  In milliseconds (not sure when the origin is)
Cryptokey:  44 byte random string
```
