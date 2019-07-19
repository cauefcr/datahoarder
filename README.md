# datahoarder
Exchange pair scraper with support for proxies and 30 exchanges

Usage:
```
Usage of ./hoarder:
  -depth int
    	select how many orders to get from book and how many prices to get from finished trades (default 1000)
  -exchange string
    	select the exchange to scrape (default "binance.com")
  -fast
    	select if the historical tick data gatherer can ignore the wait (default true)
  -interval string
    	what's the interval to get tickers at (default "5m")
  -limit int
    	what's the maximum api calls per minute, so it's throttled properly (default 200)
  -out string
    	output folder where the things will be saved (default "data")
  -proxy string
    	What proxy to use, leave black to select no proxy
  -symbol string
    	select which pair to get the orderbooks from (default "BTC_USD")
  -symboltrades string
    	set symbol for getting trades (default "BTCTUSD")
  -type string
    	select scrape type, can be: all. but in the future, price, candles, orderbook, finished trades (default "all")
  -wait int
    	Select how much time to wait in between getting data (default 300)
```

Changelog:

version 0.1: all features working on binance, on other exchanges we can't get the past trades.