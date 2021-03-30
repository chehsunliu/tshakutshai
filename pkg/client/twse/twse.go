package twse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	tkthttp "github.com/chehsunliu/tshakutshai/pkg/http"
)

// Quote is the basic unit returned by the Fetch functions.
type Quote struct {
	// Code/symbol of a stock, e.g. 0050 and 2330.
	Code string
	// Name is the Chinese stock name and only available in Client.FetchDayQuotes.
	Name string
	// Date represents the date in Client.FetchDayQuotes and Client.FetchDailyQuotes. You should ignore
	// the day field in Client.FetchMonthlyQuotes and even the month field in Client.FetchYearlyQuotes.
	Date time.Time

	Volume       uint64
	Transactions uint64
	Value        uint64

	// If no transactions are made, i.e. Transactions equals to zero, they will all zeros. Note that
	// Open and Close are only meaningful in Client.FetchDayQuotes and Client.FetchDailyQuotes.
	High  float64
	Low   float64
	Open  float64
	Close float64

	// These two fields are only used in Client.FetchYearlyQuotes.
	DateOfHigh time.Time
	DateOfLow  time.Time
}

// Client is a crawler gathering data from the TWSE server.
type Client struct {
	// HttpClient is the actual object that interacts with the TWSE server. It must not be nil; otherwise,
	// it will panic during fetching data.
	HttpClient tkthttp.Client
}

// NewClient returns a new Client, which intervals between each query are not less than minInterval.
func NewClient(minInterval time.Duration) *Client {
	return &Client{HttpClient: tkthttp.NewThrottledClient(minInterval)}
}

func (c *Client) fetch(p string, rawQuery url.Values) (map[string]json.RawMessage, error) {
	if c.HttpClient == nil {
		panic("ThrottledClient should not be nil")
	}

	u := url.URL{Scheme: "https", Host: "www.twse.com.tw", Path: p, RawQuery: rawQuery.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		panic(err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, &QuotaExceededError{"empty reply from server"}
		} else {
			return nil, &ConnectionError{fmt.Sprintf("failed to query: %s", err)}
		}
	}
	defer resp.Body.Close()

	contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		panic(err)
	}

	if contentType != "application/json" {
		return nil, &QuotaExceededError{fmt.Sprintf("received unexpected content type '%s'", contentType)}
	}

	rawData := map[string]json.RawMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		panic(err)
	}

	if stat := retrieveStat(rawData); stat != "OK" {
		return nil, &NoDataError{fmt.Sprintf("expected stat 'OK' but got '%s'", stat)}
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

func convertRawQuote(rawDayQuote map[string]interface{}) *Quote {
	return &Quote{
		Volume:       convertToStringThenUint64(rawDayQuote, "成交股數"),
		Transactions: convertToStringThenUint64(rawDayQuote, "成交筆數"),
		Value:        convertToStringThenUint64(rawDayQuote, "成交金額"),
		Open:         convertToStringThenFloat64(rawDayQuote, "開盤價"),
		High:         convertToStringThenFloat64(rawDayQuote, "最高價"),
		Low:          convertToStringThenFloat64(rawDayQuote, "最低價"),
		Close:        convertToStringThenFloat64(rawDayQuote, "收盤價"),
	}
}

func convertRawDayQuote(rawDayQuote map[string]interface{}, date time.Time) *Quote {
	q := convertRawQuote(rawDayQuote)
	q.Code = convertToString(rawDayQuote, "證券代號")
	q.Name = convertToString(rawDayQuote, "證券名稱")
	q.Date = date
	return q
}

func convertRawDailyQuote(rawDailyQuote map[string]interface{}, code string, year int, month time.Month) *Quote {
	rawDate := convertToString(rawDailyQuote, "日期")

	splitRawDate := strings.Split(rawDate, "/")
	if len(splitRawDate) != 3 {
		panic(fmt.Sprintf("'%s' of %v is ill-formatted", rawDate, rawDailyQuote))
	}

	day, err := strconv.ParseInt(splitRawDate[2], 10, 32)
	if err != nil {
		panic(fmt.Sprintf("ill-formatted date '%s' in %v", rawDate, rawDailyQuote))
	}

	q := convertRawQuote(rawDailyQuote)
	q.Code = code
	q.Date = time.Date(year, month, int(day), 0, 0, 0, 0, time.UTC)
	return q
}

func convertRawMonthlyQuote(rawMonthlyQuote map[string]interface{}, code string, year int) *Quote {
	rawMonth := convertToFloat64(rawMonthlyQuote, "月份")

	t, err := time.Parse("01", fmt.Sprintf("%02d", int(rawMonth)))
	if err != nil {
		panic(fmt.Sprintf("%f is not a legal month: %s", rawMonth, err))
	}

	return &Quote{
		Code:         code,
		Date:         time.Date(year, t.Month(), 1, 0, 0, 0, 0, time.UTC),
		Volume:       convertToStringThenUint64(rawMonthlyQuote, "成交股數(B)"),
		Transactions: convertToStringThenUint64(rawMonthlyQuote, "成交筆數"),
		Value:        convertToStringThenUint64(rawMonthlyQuote, "成交金額(A)"),
		High:         convertToStringThenFloat64(rawMonthlyQuote, "最高價"),
		Low:          convertToStringThenFloat64(rawMonthlyQuote, "最低價"),
	}
}

func convertRawYearlyQuote(rawYearlyQuote map[string]interface{}, code string) *Quote {
	rawYear := convertToFloat64(rawYearlyQuote, "年度")
	year := 1911 + int(rawYear)

	rawDateOfHigh := convertToString(rawYearlyQuote, "日期")
	dateOfHigh, err := time.Parse("2006/1/02", fmt.Sprintf("%d/%s", year, rawDateOfHigh))
	if err != nil {
		panic(fmt.Sprintf("%s is not a legal month/day: %s", rawDateOfHigh, err))
	}

	rawDateOfLow := convertToString(rawYearlyQuote, "日期2")
	dateOfLow, err := time.Parse("2006/1/02", fmt.Sprintf("%d/%s", year, rawDateOfLow))
	if err != nil {
		panic(fmt.Sprintf("%s is not a legal month/day: %s", rawDateOfLow, err))
	}

	return &Quote{
		Code:         code,
		Date:         time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
		Volume:       convertToStringThenUint64(rawYearlyQuote, "成交股數"),
		Transactions: convertToStringThenUint64(rawYearlyQuote, "成交筆數"),
		Value:        convertToStringThenUint64(rawYearlyQuote, "成交金額"),
		High:         convertToStringThenFloat64(rawYearlyQuote, "最高價"),
		Low:          convertToStringThenFloat64(rawYearlyQuote, "最低價"),
		DateOfHigh:   dateOfHigh,
		DateOfLow:    dateOfLow,
	}
}

// FetchDayQuotes returns a map that maps stock symbols to their corresponding quotes on that date.
func (c *Client) FetchDayQuotes(date time.Time) (map[string]Quote, error) {
	rawData, err := c.fetchDayQuotes(date)
	if err != nil {
		return nil, err
	}

	rawDayQuotes := zipFieldsAndItems(rawData, "fields9", "data9")

	qs := map[string]Quote{}
	for _, rawDayQuote := range rawDayQuotes {
		q := convertRawDayQuote(rawDayQuote, date)
		qs[q.Code] = *q
	}

	return qs, nil
}

// FetchDailyQuotes return a Quote slice containing daily quotes on the month of the year.
func (c *Client) FetchDailyQuotes(code string, year int, month time.Month) ([]Quote, error) {
	rawData, err := c.fetchDailyQuotes(code, year, month)
	if err != nil {
		return nil, err
	}

	rawDailyQuotes := zipFieldsAndItems(rawData, "fields", "data")

	qs := make([]Quote, 0)
	for _, rawDailyQuote := range rawDailyQuotes {
		qs = append(qs, *convertRawDailyQuote(rawDailyQuote, code, year, month))
	}

	return qs, nil
}

// FetchMonthlyQuotes return a Quote slice containing monthly quotes of the year.
func (c *Client) FetchMonthlyQuotes(code string, year int) ([]Quote, error) {
	rawData, err := c.fetchMonthlyQuotes(code, year)
	if err != nil {
		return nil, err
	}

	rawMonthlyQuotes := zipFieldsAndItems(rawData, "fields", "data")

	qs := make([]Quote, 0)
	for _, rawMonthlyQuote := range rawMonthlyQuotes {
		qs = append(qs, *convertRawMonthlyQuote(rawMonthlyQuote, code, year))
	}

	return qs, nil
}

// FetchYearlyQuotes return a Quote slice containing yearly quotes of all time.
func (c *Client) FetchYearlyQuotes(code string) ([]Quote, error) {
	rawData, err := c.fetchYearlyQuotes(code)
	if err != nil {
		return nil, err
	}

	rawYearlyQuotes := zipFieldsAndItems(rawData, "fields", "data")

	qs := make([]Quote, 0)
	for _, rawYearlyQuote := range rawYearlyQuotes {
		qs = append(qs, *convertRawYearlyQuote(rawYearlyQuote, code))
	}

	return qs, nil
}
