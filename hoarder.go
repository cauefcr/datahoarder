package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"io/ioutil"
	"net/http"
	"net/url"

	goex "github.com/nntaoli-project/GoEx"
	"github.com/nntaoli-project/GoEx/builder"
)

var (
	strToKline = map[string]int{
		"1m":  goex.KLINE_PERIOD_1MIN,
		"3m":  goex.KLINE_PERIOD_3MIN,
		"5m":  goex.KLINE_PERIOD_5MIN,
		"15m": goex.KLINE_PERIOD_15MIN,
		"30m": goex.KLINE_PERIOD_30MIN,
		"60m": goex.KLINE_PERIOD_60MIN,
		"1h":  goex.KLINE_PERIOD_1H,
		"2h":  goex.KLINE_PERIOD_2H,
		"4h":  goex.KLINE_PERIOD_4H,
		"6h":  goex.KLINE_PERIOD_6H,
		"8h":  goex.KLINE_PERIOD_8H,
		"12h": goex.KLINE_PERIOD_12H,
		"1d":  goex.KLINE_PERIOD_1DAY,
		"3d":  goex.KLINE_PERIOD_3DAY,
		"7d":  goex.KLINE_PERIOD_1WEEK,
		"1M":  goex.KLINE_PERIOD_1MONTH,
		"12M": goex.KLINE_PERIOD_1YEAR,
	}
)

var (
	exchange       string
	scrapeKind     string
	depth          int
	symbol         string
	interval       string
	limitperminute int
	outputFolder   string
	proxy          string
	binanceticker  string
	wait           int
	fasthistorical bool
)

func init() {
	flag.StringVar(&exchange, "exchange", "binance.com", "select the exchange to scrape")
	flag.StringVar(&scrapeKind, "type", "all", "select scrape type, can be: all. but in the future, price, candles, orderbook, finished trades")
	flag.IntVar(&depth, "depth", 1000, "select how many orders to get from book and how many prices to get from finished trades")
	flag.StringVar(&symbol, "symbol", "BTC_USD", "select which pair to get the orderbooks from")
	flag.StringVar(&interval, "interval", "5m", "what's the interval to get tickers at")
	flag.StringVar(&outputFolder, "out", "data", "output folder where the things will be saved")
	flag.IntVar(&limitperminute, "limit", 200, "what's the maximum api calls per minute, so it's throttled properly")
	flag.StringVar(&proxy, "proxy", "", "What proxy to use, leave black to select no proxy")
	flag.StringVar(&binanceticker, "symboltrades", "BTCTUSD", "set symbol for getting trades")
	flag.IntVar(&wait, "wait", 5*60, "Select how much time to wait in between getting data")
	flag.BoolVar(&fasthistorical, "fast", true, "select if the historical tick data gatherer can ignore the wait")
	flag.Parse()
	outputFolder = outputFolder + "/"
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ratelimit() {
	time.Sleep(time.Duration(limitperminute * 1000 / 60))
}

func proxyGet(proxyStr string, urlStr string) ([]byte, error) {

	// proxyStr := "http://localhost:7000"
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return nil, err
	}

	//creating the URL to be loaded through the proxy
	// urlStr := "http://httpbin.org/get"
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{}
	//adding the proxy settings to the Transport object
	if proxyStr != "" {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	//adding the Transport object to the http Client
	client := &http.Client{
		Transport: transport,
	}

	//generating the HTTP GET request
	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	//calling the URL
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	//getting the response
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return data, nil

}

// func historicRecordCandles(api goex.API, pair goex.CurrencyPair) {
// 	last := 0
// 	//check if file with such date exists
// 	//if it exists, open, read, and only add what it's needed
// 	//if it doesn't create it and start appending files
// 	newf := func(date int64) *csv.Writer {
// 		newfile := false
// 		tm := time.Unix(date, 0)
// 		fticker := outputFolder + pair.String() + fmt.Sprint(tm.Year()) + "-" + tm.Month().String() + "-" + fmt.Sprint(tm.Day()) + "-candles.csv"
// 		fileTicker := &os.File{}

// 		if _, err := os.Stat(fticker); os.IsNotExist(err) {
// 			fileTicker, err = os.Create(fticker)
// 			fileTicker.Close()
// 			must(err)
// 			newfile = true
// 		}

// 		fileTicker, err := os.OpenFile(fticker, os.O_APPEND|os.O_WRONLY, 0600)
// 		must(err)

// 		wTicker := csv.NewWriter(fileTicker)
// 		if newfile {
// 			newfile = false
// 			wTicker.Write([]string{"Date", "Symbol", "Buy", "Sell", "High", "Low", "Volume"})
// 		}
// 		return wTicker
// 	}
// 	fileTicker := &csv.Writer{}
// 	for int64(last) <= time.Now().Unix()-3600 {
// 		lines, err := api.GetKlineRecords(pair, strToKline[interval], depth, last)
// 		if err != nil {
// 			panic(err)
// 		}
// 		fileTicker = newf(lines[0].Timestamp)
// 		// log.Printf("%+v\n",lines)
// 		if scrapeKind == "all" {
// 			ratelimit()
// 			ratelimit()
// 			ratelimit()
// 		} else {
// 			ratelimit()
// 		}
// 		for _, v := range lines {
// 			fileTicker.Write([]string{fmt.Sprint(v.Timestamp), symbol, fmt.Sprint(v.Open), fmt.Sprint(v.Close), fmt.Sprint(v.High), fmt.Sprint(v.Low), fmt.Sprint(v.Vol)})
// 			// b, err := json.Marshal(v)
// 			// if err != nil {
// 			// 	panic(err)
// 			// }
// 			// fmt.Println(string(b))
// 		}
// 		fileTicker.Flush()
// 		if !fasthistorical {
// 			time.Sleep(time.Duration(wait) * time.Second)
// 		}
// 		// fmt.Print("here")
// 	}
// }

//     "id": 28457,
//     "price": "4.00000100",
//     "qty": "12.00000000",
//     "time": 1499865549590,
//     "isBuyerMaker": true,
//     "isBestMatch": true
//   }
type trade struct {
	Time         uint64  `json: time`
	Id           int     `json: id`
	Price        float64 `json: price`
	Quantity     float64 `json: qty`
	IsBuyerMaker bool    `json: isBuyerMaker`
	IsBestMatch  bool    `json: isBestMatch`
}

type errmsg struct {
	Code int    `json: code`
	Msg  string `json: msg`
}

func main() {
	pair := goex.NewCurrencyPair2(symbol)
	apiBuilder := builder.NewAPIBuilder().HttpTimeout(5 * time.Second)
	//apiBuilder := builder.NewAPIBuilder().HttpTimeout(5 * time.Second).HttpProxy("socks5://127.0.0.1:1080")

	//build spot api
	//api := apiBuilder.APIKey("").APISecretkey("").ClientID("123").Build(goex.BITSTAMP)
	api := apiBuilder.APIKey("").APISecretkey("").Build(exchange)

	log.Println(api.GetExchangeName())
	// ratelimit()
	//check if files exist, if not create them

	go func() {
		year, month, day := time.Now().Date()
		day--
		lastTicker := goex.Ticker{}
		lastTrade := trade{}
		wTicker := &csv.Writer{}
		wDepth := &csv.Writer{}
		wTrades := &csv.Writer{}
		fileTicker := &os.File{}
		fileDepth := &os.File{}
		fileTrades := &os.File{}

		for {
			if time.Now().Day() != day {
				year, month, day = time.Now().Date()

				newfile := false

				lastTicker = goex.Ticker{}
				fticker := outputFolder + pair.String() + exchange + fmt.Sprint(year) + "-" + month.String() + "-" + fmt.Sprint(day) + "-candles.csv"
				fileTicker = &os.File{}

				if _, err := os.Stat(fticker); os.IsNotExist(err) {
					fileTicker, err = os.Create(fticker)
					fileTicker.Close()
					must(err)
					newfile = true
				}

				fileTicker, err := os.OpenFile(fticker, os.O_APPEND|os.O_WRONLY, 0600)
				must(err)

				wTicker = csv.NewWriter(fileTicker)
				if newfile {
					newfile = false
					wTicker.Write([]string{"Date", "Symbol", "Buy", "Sell", "High", "Low", "Volume"})
				}

				// lastDepth := goex.Depth{UTime: time.Now()}
				fDepth := outputFolder + pair.String() + exchange + fmt.Sprint(year) + "-" + month.String() + "-" + fmt.Sprint(day) + "-ob.csv"
				fileDepth = &os.File{}
				if _, err := os.Stat(fDepth); os.IsNotExist(err) {
					fileDepth, err = os.Create(fDepth)
					fileDepth.Close()
					must(err)
					newfile = true
				}
				// else {
				fileDepth, err = os.OpenFile(fDepth, os.O_APPEND|os.O_WRONLY, 0600)
				must(err)
				// }
				wDepth = csv.NewWriter(fileDepth)
				if newfile {
					newfile = false
					out := []string{"Date", "Symbol"}
					for i := 0; i < depth; i++ {
						out = append(out, "askPrice"+fmt.Sprint(i))
						out = append(out, "askValue"+fmt.Sprint(i))
					}
					for i := 0; i < depth; i++ {
						out = append(out, "bidPrice"+fmt.Sprint(i))
						out = append(out, "bidValue"+fmt.Sprint(i))
					}
					wDepth.Write(out)
				}

				lastTrade = trade{}
				fTrades := outputFolder + pair.String() + exchange + fmt.Sprint(year) + "-" + month.String() + "-" + fmt.Sprint(day) + "-trades.csv"
				fileTrades = &os.File{}
				if _, err := os.Stat(fTrades); os.IsNotExist(err) {
					fileTrades, err = os.Create(fTrades)
					must(err)
					fileTrades.Close()
					newfile = true
				}
				// else {
				fileTrades, err = os.OpenFile(fTrades, os.O_APPEND|os.O_WRONLY, 0600)
				must(err)
				// }
				wTrades = csv.NewWriter(fileTrades)
				if newfile {
					newfile = false
					wTrades.Write([]string{"Date", "Symbol", "Id", "Price", "Quantity", "isBuyerMaker", "isBestMatch"})
				}
			}

			ticker, err := api.GetTicker(pair)
			must(err)

			fmt.Print(lastTicker.Date, ticker.Date)
			if *ticker == lastTicker {
				_ = ticker
			} else if ticker.Date >= lastTicker.Date {
				wTicker.Write([]string{fmt.Sprint(ticker.Date), symbol, fmt.Sprint(ticker.Buy), fmt.Sprint(ticker.Sell), fmt.Sprint(ticker.High), fmt.Sprint(ticker.Low), fmt.Sprint(ticker.Vol)})
				lastTicker = *ticker
				fmt.Print("here")
			}

			orderbook, err := api.GetDepth(depth, pair)
			must(err)
			out := []string{fmt.Sprint(orderbook.UTime.Unix()), symbol}
			for _, v := range orderbook.AskList {
				out = append(out, fmt.Sprint(v.Price))
				out = append(out, fmt.Sprint(v.Amount))
			}
			for _, v := range orderbook.BidList {
				out = append(out, fmt.Sprint(v.Price))
				out = append(out, fmt.Sprint(v.Amount))
			}
			wDepth.Write(out)
			// lastDepth = *orderbook

			data := []byte{}
			switch api.GetExchangeName() {
			case goex.BINANCE:
				data, err = proxyGet(proxy, fmt.Sprintf("http://%s/api/v1/trades?limit=%d&symbol=%s",
					goex.BINANCE, depth, binanceticker))
				must(err)

				tmp := []trade{}
				json.Unmarshal(data, &tmp)

				errcheck := errmsg{}
				json.Unmarshal(data, &errcheck)
				if errcheck.Code == -1121 {
					log.Fatal("Invalid ticker")
				}

				//todo: check if there are no duplicates
				for i, v := range tmp {
					if v.Id == lastTrade.Id {
						tmp = tmp[i:]
					}
				}
				for _, v := range tmp {
					// fmt.Print("here")
					wTrades.Write([]string{fmt.Sprint(v.Time), symbol, fmt.Sprint(v.Id), fmt.Sprint(v.Price), fmt.Sprint(v.Quantity), fmt.Sprint(v.IsBuyerMaker), fmt.Sprint(v.IsBestMatch)})
					lastTrade = v
				}
			default:
				// panic("not implemented!")
				log.Print("not implemented")
			}

			ratelimit()
			ratelimit()
			ratelimit()
			time.Sleep(time.Duration(wait) * time.Second)
			wDepth.Flush()
			wTicker.Flush()
			wTrades.Flush()
			fileTrades.Close()
			fileTicker.Close()
			fileDepth.Close()
		}
	}()

	// go historicRecordCandles(api, pair)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	//log.Println(api.GetAccount())
	//log.Println(api.GetUnfinishOrders(goex.BTC_USD))
}
