// Copyright 2016 Tristan Colgate-McFarlane
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr = flag.String("listen-address", ":9666", "The address to listen on for HTTP requests.")

	// These are metrics for the collector itself
	queryDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "yquotes_query_duration_seconds",
			Help: "Duration of queries to the yahoo API",
		},
	)
	queryCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yquotes_queries_total",
			Help: "Count of completed queries",
		},
		[]string{"symbol"},
	)
	errorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yquotes_failed_queries_total",
			Help: "Count of failed queries",
		},
		[]string{"symbol"},
	)
)

func transformResponseQuote(responseQuote ResponseQuote) Quote {

	if responseQuote.MarketState == "REGULAR" {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.RegularMarketPrice,
			PricePrevClose:          responseQuote.RegularMarketPreviousClose,
			PriceOpen:               responseQuote.RegularMarketOpen,
			PriceDayHigh:            responseQuote.RegularMarketDayHigh,
			PriceDayLow:             responseQuote.RegularMarketDayLow,
			Change:                  (responseQuote.RegularMarketChange),
			ChangePercent:           responseQuote.RegularMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: true,
			IsVariablePrecision:     false,
			CurrencyConverted:       "USD",
			Name:                    responseQuote.ShortName,
			Symbol:                  responseQuote.Symbol,
		}
	}

	if responseQuote.MarketState == "POST" && responseQuote.PostMarketPrice == 0.0 {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.RegularMarketPrice,
			PricePrevClose:          responseQuote.RegularMarketPreviousClose,
			PriceOpen:               responseQuote.RegularMarketOpen,
			PriceDayHigh:            responseQuote.RegularMarketDayHigh,
			PriceDayLow:             responseQuote.RegularMarketDayLow,
			Change:                  (responseQuote.RegularMarketChange),
			ChangePercent:           responseQuote.RegularMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: false,
			IsVariablePrecision:     false,
			CurrencyConverted:       "USD",
			Name:                    responseQuote.ShortName,
			Symbol:                  responseQuote.Symbol,
		}
	}

	if responseQuote.MarketState == "PRE" && responseQuote.PreMarketPrice == 0.0 {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.RegularMarketPrice,
			PricePrevClose:          responseQuote.RegularMarketPreviousClose,
			PriceOpen:               responseQuote.RegularMarketOpen,
			PriceDayHigh:            responseQuote.RegularMarketDayHigh,
			PriceDayLow:             responseQuote.RegularMarketDayLow,
			Change:                  (responseQuote.RegularMarketChange),
			ChangePercent:           responseQuote.RegularMarketChangePercent,
			IsActive:                false,
			IsRegularTradingSession: false,
			IsVariablePrecision:     false,
			CurrencyConverted:       "USD",
			Name:                    responseQuote.ShortName,
			Symbol:                  responseQuote.Symbol,
		}
	}

	if responseQuote.MarketState == "POST" {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.PostMarketPrice,
			PricePrevClose:          responseQuote.RegularMarketPreviousClose,
			PriceOpen:               responseQuote.RegularMarketOpen,
			PriceDayHigh:            responseQuote.RegularMarketDayHigh,
			PriceDayLow:             responseQuote.RegularMarketDayLow,
			Change:                  (responseQuote.PostMarketChange + responseQuote.RegularMarketChange),
			ChangePercent:           responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: false,
			IsVariablePrecision:     false,
			CurrencyConverted:       "USD",
			Name:                    responseQuote.ShortName,
			Symbol:                  responseQuote.Symbol,
		}
	}

	if responseQuote.MarketState == "PRE" {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.PreMarketPrice,
			PricePrevClose:          responseQuote.RegularMarketPreviousClose,
			PriceOpen:               responseQuote.RegularMarketOpen,
			PriceDayHigh:            responseQuote.RegularMarketDayHigh,
			PriceDayLow:             responseQuote.RegularMarketDayLow,
			Change:                  (responseQuote.PreMarketChange),
			ChangePercent:           responseQuote.PreMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: false,
			IsVariablePrecision:     false,
			CurrencyConverted:       "USD",
			Name:                    responseQuote.ShortName,
			Symbol:                  responseQuote.Symbol,
		}
	}

	if responseQuote.PostMarketPrice != 0.0 {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.PostMarketPrice,
			PricePrevClose:          responseQuote.RegularMarketPreviousClose,
			PriceOpen:               responseQuote.RegularMarketOpen,
			PriceDayHigh:            responseQuote.RegularMarketDayHigh,
			PriceDayLow:             responseQuote.RegularMarketDayLow,
			Change:                  (responseQuote.PostMarketChange + responseQuote.RegularMarketChange),
			ChangePercent:           responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent,
			IsActive:                false,
			IsRegularTradingSession: false,
			IsVariablePrecision:     false,
			CurrencyConverted:       "USD",
			Name:                    responseQuote.ShortName,
			Symbol:                  responseQuote.Symbol,
		}
	}

	return Quote{
		ResponseQuote: responseQuote,

		Price:                   responseQuote.RegularMarketPrice,
		PricePrevClose:          responseQuote.RegularMarketPreviousClose,
		PriceOpen:               responseQuote.RegularMarketOpen,
		PriceDayHigh:            responseQuote.RegularMarketDayHigh,
		PriceDayLow:             responseQuote.RegularMarketDayLow,
		Change:                  (responseQuote.RegularMarketChange),
		ChangePercent:           responseQuote.RegularMarketChangePercent,
		IsActive:                false,
		IsRegularTradingSession: false,
		IsVariablePrecision:     false,
		CurrencyConverted:       "USD",
		Name:                    responseQuote.ShortName,
		Symbol:                  responseQuote.Symbol,
	}

}

type ResponseQuote struct {
	ShortName                  string  `json:"shortName"`
	Symbol                     string  `json:"symbol"`
	MarketState                string  `json:"marketState"`
	Currency                   string  `json:"currency"`
	ExchangeName               string  `json:"fullExchangeName"`
	ExchangeDelay              float64 `json:"exchangeDataDelayedBy"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayRange      string  `json:"regularMarketDayRange"`
	RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
	RegularMarketVolume        float64 `json:"regularMarketVolume"`
	PostMarketChange           float64 `json:"postMarketChange"`
	PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
	PostMarketPrice            float64 `json:"postMarketPrice"`
	PreMarketChange            float64 `json:"preMarketChange"`
	PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
	PreMarketPrice             float64 `json:"preMarketPrice"`
	FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
	QuoteType                  string  `json:"quoteType"`
	MarketCap                  float64 `json:"marketCap"`
}

type Quote struct {
	ResponseQuote
	Price                   float64
	PricePrevClose          float64
	PriceOpen               float64
	PriceDayHigh            float64
	PriceDayLow             float64
	Change                  float64
	ChangePercent           float64
	IsActive                bool
	IsRegularTradingSession bool
	IsVariablePrecision     bool
	CurrencyConverted       string
	Name                    string
	Symbol                  string
}

type Response struct {
	QuoteResponse struct {
		Quotes []ResponseQuote `json:"result"`
		Error  interface{}     `json:"error"`
	} `json:"quoteResponse"`
}

func GetQuotes(symbols []string) []Quote {
	client := resty.New()
	symbolsString := strings.Join(symbols, ",")
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=%s", symbolsString)
	res, _ := client.R().
		SetResult(Response{}).
		Get(url)

	return transformResponseQuotes((res.Result().(*Response)).QuoteResponse.Quotes)
}

func transformResponseQuotes(responseQuotes []ResponseQuote) []Quote {

	quotes := make([]Quote, 0)
	for _, responseQuote := range responseQuotes {
		quotes = append(quotes, transformResponseQuote(responseQuote))
	}
	return quotes

}

func init() {
	// register the collector metrics in the default
	// registry.
	prometheus.MustRegister(queryDuration)
	prometheus.MustRegister(queryCount)
	prometheus.MustRegister(errorCount)
}

func main() {
	flag.Parse()
	http.HandleFunc("/price", getPrice)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}

type collector []string

func (c collector) Describe(ch chan<- *prometheus.Desc) {
	// Must send one description, or the registry panics
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (c collector) Collect(ch chan<- prometheus.Metric) {
	stocks := make([]string, 0)

	for _, s := range c {
		if s == "" {
			// should never happen
			continue
		}

		queryCount.WithLabelValues(s).Inc()
		if glog.V(2) {
			glog.Infof("looking up %s\n", s)
		}
		stocks = append(stocks, s)
	}

	start := time.Now()
	quoteList := GetQuotes(stocks)
	queryDuration.Observe(float64(time.Since(start).Seconds()))
	for _, singleStock := range quoteList {

		ls := []string{"symbol", "name", "active"}
		lvs := []string{singleStock.Symbol, singleStock.Name, strconv.FormatBool(singleStock.IsActive)}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("yquotes_last_price_dollars", "Last price paid.", ls, nil),
			prometheus.GaugeValue,
			singleStock.Price,
			lvs...,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("yquotes_opening_price_dollars", "Opening price.", ls, nil),
			prometheus.GaugeValue,
			singleStock.PriceOpen,
			lvs...,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("yquotes_previous_close_price_dollars", "Previous close price.", ls, nil),
			prometheus.GaugeValue,
			singleStock.PricePrevClose,
			lvs...,
		)
	}
}

func getPrice(w http.ResponseWriter, r *http.Request) {
	syms, ok := r.URL.Query()["sym"]
	if !ok {
		glog.Infof("no syms given")
		return
	}

	registry := prometheus.NewRegistry()

	collector := collector(syms)
	registry.MustRegister(collector)

	// Delegate http serving to Promethues client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
