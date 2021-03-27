package twse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Quote struct {
	Code string
	Name string
	Date time.Time

	Volume       uint64
	Transactions uint64
	Value        uint64

	Open  float64
	High  float64
	Low   float64
	Close float64
}

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

func (c *Client) fetch(p string, rawQuery url.Values) (map[string]json.RawMessage, error) {
	u := site
	u.Path = p
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

func (c *Client) fetchDayQuotes(date time.Time) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("type", "ALL")
	return c.fetch("/exchangeReport/MI_INDEX", rawQuery)
}

func (c *Client) fetchDailyQuotes(code string, year int, month time.Month) (map[string]json.RawMessage, error) {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("stockNo", code)
	return c.fetch("/exchangeReport/STOCK_DAY", rawQuery)
}

func (c *Client) fetchMonthlyQuotes(code string, year int) (map[string]json.RawMessage, error) {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("date", date.Format("20060102"))
	rawQuery.Set("stockNo", code)
	return c.fetch("/exchangeReport/FMSRFK", rawQuery)
}

func (c *Client) fetchYearlyQuotes(code string) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("response", "json")
	rawQuery.Set("stockNo", code)
	return c.fetch("/exchangeReport/FMNPTK", rawQuery)
}

func convertRawQuote(rawDayQuote map[string]interface{}) (*Quote, error) {
	volume, err := convertToUint64(rawDayQuote, "成交股數")
	if err != nil {
		return nil, err
	}

	transactions, err := convertToUint64(rawDayQuote, "成交筆數")
	if err != nil {
		return nil, err
	}

	value, err := convertToUint64(rawDayQuote, "成交金額")
	if err != nil {
		return nil, err
	}

	open, err := convertToFloat64(rawDayQuote, "開盤價")
	if err != nil {
		return nil, err
	}

	high, err := convertToFloat64(rawDayQuote, "最高價")
	if err != nil {
		return nil, err
	}

	low, err := convertToFloat64(rawDayQuote, "最低價")
	if err != nil {
		return nil, err
	}

	klose, err := convertToFloat64(rawDayQuote, "收盤價")
	if err != nil {
		return nil, err
	}

	return &Quote{
		Volume:       volume,
		Transactions: transactions,
		Value:        value,

		Open:  open,
		High:  high,
		Low:   low,
		Close: klose,
	}, nil
}

func convertRawDayQuote(rawDayQuote map[string]interface{}, date time.Time) (*Quote, error) {
	code, err := convertToString(rawDayQuote, "證券代號")
	if err != nil {
		return nil, err
	}

	name, err := convertToString(rawDayQuote, "證券名稱")
	if err != nil {
		return nil, err
	}

	q, err := convertRawQuote(rawDayQuote)
	if err != nil {
		return nil, err
	}

	q.Code = code
	q.Name = name
	q.Date = date
	return q, nil
}

func convertRawDailyQuote(rawDailyQuote map[string]interface{}, code string, year int, month time.Month) (*Quote, error) {
	q, err := convertRawQuote(rawDailyQuote)
	if err != nil {
		return nil, err
	}

	rawDate, err := convertToString(rawDailyQuote, "日期")
	if err != nil {
		return nil, err
	}

	splitRawDate := strings.Split(rawDate, "/")
	if len(splitRawDate) != 3 {
		return nil, fmt.Errorf("'%s' of %v is ill-formatted", rawDate, rawDailyQuote)
	}

	day, err := strconv.ParseInt(splitRawDate[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ill-formatted date '%s' in %v", rawDate, rawDailyQuote)
	}

	q.Code = code
	q.Date = time.Date(year, month, int(day), 0, 0, 0, 0, time.UTC)
	return q, nil
}

func (c *Client) FetchDayQuotes(date time.Time) (map[string]Quote, error) {
	rawData, err := c.fetchDayQuotes(date)
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

	rawDayQuotes, err := zipFieldsAndItems(fields, items)
	if err != nil {
		return nil, err
	}

	qs := map[string]Quote{}
	for _, rawDayQuote := range rawDayQuotes {
		q, err := convertRawDayQuote(rawDayQuote, date)
		if err != nil {
			return nil, err
		}

		qs[q.Code] = *q
	}

	return qs, nil
}

func (c *Client) FetchDailyQuotes(code string, year int, month time.Month) ([]Quote, error) {
	rawData, err := c.fetchDailyQuotes(code, year, month)
	if err != nil {
		return nil, err
	}

	fields, err := retrieveFields(rawData, "fields")
	if err != nil {
		return nil, err
	}

	items, err := retrieveItems(rawData, "data")
	if err != nil {
		return nil, err
	}

	rawDailyQuotes, err := zipFieldsAndItems(fields, items)
	if err != nil {
		return nil, err
	}

	qs := make([]Quote, 0)
	for _, rawDailyQuote := range rawDailyQuotes {
		q, err := convertRawDailyQuote(rawDailyQuote, code, year, month)
		if err != nil {
			return nil, err
		}

		qs = append(qs, *q)
	}

	return qs, nil
}
