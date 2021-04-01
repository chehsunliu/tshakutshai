package tpex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	tkthttp "github.com/chehsunliu/tshakutshai/pkg/http"
)

type Quote struct {
	Code         string
	Name         string
	Date         time.Time
	Volume       uint64
	Transactions uint64
	Value        uint64
	High         float64
	Low          float64
	Open         float64
	Close        float64
}

type Client struct {
	HttpClient tkthttp.Client
}

func NewClient(minInterval time.Duration) *Client {
	return &Client{HttpClient: tkthttp.NewThrottledClient(&http.Client{}, minInterval)}
}

func (c *Client) fetch(p string, rawQuery url.Values) (map[string]json.RawMessage, error) {
	if c.HttpClient == nil {
		panic("Client.HttpClient should not be nil")
	}

	u := url.URL{Scheme: "https", Host: "www.tpex.org.tw", Path: p, RawQuery: rawQuery.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		panic(err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawData := map[string]json.RawMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	return rawData, nil
}

func (c *Client) fetchDayQuotes(date time.Time) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("d", fmt.Sprintf("%d/%s", date.Year()-1911, date.Format("01/02")))
	rawQuery.Set("l", "zh-tw")
	return c.fetch("/web/stock/aftertrading/daily_close_quotes/stk_quote_result.php", rawQuery)
}

func (c *Client) fetchDailyQuotes(code string, year int, month time.Month) (map[string]json.RawMessage, error) {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	rawQuery := url.Values{}
	rawQuery.Set("d", fmt.Sprintf("%d/%s", date.Year()-1911, date.Format("01/02")))
	rawQuery.Set("l", "zh-tw")
	rawQuery.Set("stkno", code)
	return c.fetch("/web/stock/aftertrading/daily_trading_info/st43_result.php", rawQuery)
}

func (c *Client) FetchDayQuotes(date time.Time) (map[string]Quote, error) {
	rawData, err := c.fetchDayQuotes(date)
	if err != nil {
		return nil, err
	}

	qs := map[string]Quote{}

	items := deserializeSliceOfSlicesOfStrings(rawData, "aaData")
	for _, item := range items {
		q := Quote{
			Code:         item[0],
			Name:         item[1],
			Date:         date,
			Volume:       stringToUint64(item[8]),
			Transactions: stringToUint64(item[10]),
			Value:        stringToUint64(item[9]),
			High:         stringToFloat64(item[5]),
			Low:          stringToFloat64(item[6]),
			Open:         stringToFloat64(item[4]),
			Close:        stringToFloat64(item[2]),
		}
		qs[q.Code] = q
	}

	return qs, nil
}

func (c *Client) FetchDailyQuotes(code string, year int, month time.Month) ([]Quote, error) {
	rawData, err := c.fetchDailyQuotes(code, year, month)
	if err != nil {
		return nil, err
	}

	qs := make([]Quote, 0)

	items := deserializeSliceOfSlicesOfStrings(rawData, "aaData")
	for _, item := range items {
		q := Quote{
			Code:         deserializeString(rawData, "stkNo"),
			Name:         deserializeString(rawData, "stkName"),
			Date:         stringToDate(item[0]),
			Volume:       stringToUint64(item[1]) * 1000,
			Transactions: stringToUint64(item[8]),
			Value:        stringToUint64(item[2]) * 1000,
			Open:         stringToFloat64(item[3]),
			Close:        stringToFloat64(item[6]),
			High:         stringToFloat64(item[4]),
			Low:          stringToFloat64(item[5]),
		}
		qs = append(qs, q)
	}

	return qs, nil
}
