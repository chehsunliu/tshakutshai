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

func (c *Client) FetchDayQuotes(date time.Time) (map[string]Quote, error) {
	_, err := c.fetchDayQuotes(date)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
