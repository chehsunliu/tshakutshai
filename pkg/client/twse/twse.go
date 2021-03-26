package twse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var site = url.URL{Scheme: "https", Host: "www.twse.com.tw"}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	http httpClient
}

func NewClient() *Client {
	return &Client{http: &http.Client{}}
}

func (c *Client) get(u url.URL) (map[string]json.RawMessage, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http.Request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query '%s': %w", u.String(), err)
	}
	defer resp.Body.Close()

	data := map[string]json.RawMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return data, nil
}

func (c *Client) getRawQuotesOfDay(date time.Time) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("type", "ALL")

	u := site
	u.Path = "exchangeReport/MI_INDEX"
	u.RawQuery = rawQuery.Encode()

	return c.get(u)
}

func (c *Client) getRawDailyQuotes(code string, year int, month time.Month) (map[string]json.RawMessage, error) {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("stockNo", code)

	u := site
	u.Path = "exchangeReport/STOCK_DAY"
	u.RawQuery = rawQuery.Encode()

	return c.get(u)
}

func (c *Client) getRawMonthlyQuotes(code string, year int) (map[string]json.RawMessage, error) {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("stockNo", code)

	u := site
	u.Path = "exchangeReport/FMSRFK"
	u.RawQuery = rawQuery.Encode()

	return c.get(u)
}

func (c *Client) getRawYearlyQuotes(code string) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("stockNo", code)

	u := site
	u.Path = "exchangeReport/FMNPTK"
	u.RawQuery = rawQuery.Encode()

	return c.get(u)
}
