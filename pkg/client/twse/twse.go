package twse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/chehsunliu/tshakutshai/pkg/quote"
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

func (c *Client) get(p string, rawQuery url.Values) (map[string]json.RawMessage, error) {
	u := site
	u.Path = path.Join("en", p)
	u.RawQuery = rawQuery.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http.Request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query '%s': %w", u.String(), err)
	}
	defer resp.Body.Close()

	rawData := map[string]json.RawMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return rawData, nil
}

func (c *Client) getRawQuotesOfDay(date time.Time) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("type", "ALL")
	return c.get("/exchangeReport/MI_INDEX", rawQuery)
}

func (c *Client) getRawDailyQuotes(code string, year int, month time.Month) (map[string]json.RawMessage, error) {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("stockNo", code)
	return c.get("/exchangeReport/STOCK_DAY", rawQuery)
}

func (c *Client) getRawMonthlyQuotes(code string, year int) (map[string]json.RawMessage, error) {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("stockNo", code)
	return c.get("/exchangeReport/FMSRFK", rawQuery)
}

func (c *Client) getRawYearlyQuotes(code string) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("stockNo", code)
	return c.get("/exchangeReport/FMNPTK", rawQuery)
}

func (c *Client) GetQuotesOfDay(date time.Time) ([]quote.Quote, error) {
	rawData, err := c.getRawQuotesOfDay(date)
	if err != nil {
		return nil, err
	}

	fields, err := retrieveFields(rawData, "fields9")
	if err != nil {
		return nil, err
	}

	items, err := retrieveItems(rawData, "data9")
	if err != nil {
		return nil, err
	}

	rawItems, err := zipFieldsAndItems(fields, items)
	if err != nil {
		return nil, err
	}

	fmt.Println(rawItems[0])

	return nil, nil
}
