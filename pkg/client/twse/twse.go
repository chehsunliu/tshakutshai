package twse

import (
	"encoding/json"
	"fmt"
	"mime"
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

type MonthlyQuote struct {
	Code  string
	Year  int
	Month time.Month

	Volume       uint64
	Transactions uint64
	Value        uint64

	High float64
	Low  float64
}

type YearlyQuote struct {
	Code string
	Year int

	Volume       uint64
	Transactions uint64
	Value        uint64

	High       float64
	Low        float64
	DateOfHigh time.Time
	DateOfLow  time.Time
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

	contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err == nil && contentType == "text/html" {
		return nil, ErrQuotaExceeded
	}

	rawData := map[string]json.RawMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	if ok, err := isStatOK(rawData); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoData
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
	volume, err := convertToStringThenUint64(rawDayQuote, "成交股數")
	if err != nil {
		return nil, err
	}

	transactions, err := convertToStringThenUint64(rawDayQuote, "成交筆數")
	if err != nil {
		return nil, err
	}

	value, err := convertToStringThenUint64(rawDayQuote, "成交金額")
	if err != nil {
		return nil, err
	}

	open, err := convertToStringThenFloat64(rawDayQuote, "開盤價")
	if err != nil {
		return nil, err
	}

	high, err := convertToStringThenFloat64(rawDayQuote, "最高價")
	if err != nil {
		return nil, err
	}

	low, err := convertToStringThenFloat64(rawDayQuote, "最低價")
	if err != nil {
		return nil, err
	}

	klose, err := convertToStringThenFloat64(rawDayQuote, "收盤價")
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

func convertRawMonthlyQuote(rawMonthlyQuote map[string]interface{}, code string, year int) (*MonthlyQuote, error) {
	rawMonth, err := convertToFloat64(rawMonthlyQuote, "月份")
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("01", fmt.Sprintf("%02d", int(rawMonth)))
	if err != nil {
		return nil, fmt.Errorf("%f is not a legal month: %w", rawMonth, err)
	}

	high, err := convertToStringThenFloat64(rawMonthlyQuote, "最高價")
	if err != nil {
		return nil, err
	}

	low, err := convertToStringThenFloat64(rawMonthlyQuote, "最低價")
	if err != nil {
		return nil, err
	}

	transactions, err := convertToStringThenUint64(rawMonthlyQuote, "成交筆數")
	if err != nil {
		return nil, err
	}

	value, err := convertToStringThenUint64(rawMonthlyQuote, "成交金額(A)")
	if err != nil {
		return nil, err
	}

	volume, err := convertToStringThenUint64(rawMonthlyQuote, "成交股數(B)")
	if err != nil {
		return nil, err
	}

	return &MonthlyQuote{
		Code:  code,
		Year:  year,
		Month: t.Month(),

		Volume:       volume,
		Transactions: transactions,
		Value:        value,

		High: high,
		Low:  low,
	}, nil
}

func convertRawYearlyQuote(rawYearlyQuote map[string]interface{}, code string) (*YearlyQuote, error) {
	rawYear, err := convertToFloat64(rawYearlyQuote, "年度")
	if err != nil {
		return nil, err
	}
	year := 1911 + int(rawYear)

	transactions, err := convertToStringThenUint64(rawYearlyQuote, "成交筆數")
	if err != nil {
		return nil, err
	}

	value, err := convertToStringThenUint64(rawYearlyQuote, "成交金額")
	if err != nil {
		return nil, err
	}

	volume, err := convertToStringThenUint64(rawYearlyQuote, "成交股數")
	if err != nil {
		return nil, err
	}

	high, err := convertToStringThenFloat64(rawYearlyQuote, "最高價")
	if err != nil {
		return nil, err
	}

	low, err := convertToStringThenFloat64(rawYearlyQuote, "最低價")
	if err != nil {
		return nil, err
	}

	rawDateOfHigh, err := convertToString(rawYearlyQuote, "日期")
	if err != nil {
		return nil, err
	}

	dateOfHigh, err := time.Parse("2006/1/02", fmt.Sprintf("%d/%s", year, rawDateOfHigh))
	if err != nil {
		return nil, fmt.Errorf("%s is not a legal month/day: %w", rawDateOfHigh, err)
	}

	rawDateOfLow, err := convertToString(rawYearlyQuote, "日期2")
	if err != nil {
		return nil, err
	}

	dateOfLow, err := time.Parse("2006/1/02", fmt.Sprintf("%d/%s", year, rawDateOfLow))
	if err != nil {
		return nil, fmt.Errorf("%s is not a legal month/day: %w", rawDateOfLow, err)
	}

	return &YearlyQuote{
		Code: code,
		Year: year,

		Volume:       volume,
		Transactions: transactions,
		Value:        value,

		High:       high,
		Low:        low,
		DateOfHigh: dateOfHigh,
		DateOfLow:  dateOfLow,
	}, nil
}

func (c *Client) FetchDayQuotes(date time.Time) (map[string]Quote, error) {
	rawData, err := c.fetchDayQuotes(date)
	if err != nil {
		return nil, err
	}

	rawDayQuotes, err := zipFieldsAndItems(rawData, "fields9", "data9")
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

	rawDailyQuotes, err := zipFieldsAndItems(rawData, "fields", "data")
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

func (c *Client) FetchMonthlyQuotes(code string, year int) ([]MonthlyQuote, error) {
	rawData, err := c.fetchMonthlyQuotes(code, year)
	if err != nil {
		return nil, err
	}

	rawMonthlyQuotes, err := zipFieldsAndItems(rawData, "fields", "data")
	if err != nil {
		return nil, err
	}

	qs := make([]MonthlyQuote, 0)
	for _, rawMonthlyQuote := range rawMonthlyQuotes {
		q, err := convertRawMonthlyQuote(rawMonthlyQuote, code, year)
		if err != nil {
			return nil, err
		}

		qs = append(qs, *q)
	}

	return qs, nil
}

func (c *Client) FetchYearlyQuotes(code string) ([]YearlyQuote, error) {
	rawData, err := c.fetchYearlyQuotes(code)
	if err != nil {
		return nil, err
	}

	rawYearlyQuotes, err := zipFieldsAndItems(rawData, "fields", "data")
	if err != nil {
		return nil, err
	}

	qs := make([]YearlyQuote, 0)
	for _, rawYearlyQuote := range rawYearlyQuotes {
		q, err := convertRawYearlyQuote(rawYearlyQuote, code)
		if err != nil {
			return nil, err
		}

		qs = append(qs, *q)
	}

	return qs, nil
}
