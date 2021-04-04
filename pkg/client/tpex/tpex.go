package tpex

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	tkthttp "github.com/chehsunliu/tshakutshai/pkg/http"
)

var invalidCsvChars = regexp.MustCompile(`[a-zA-Z]+`)

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
	DateOfHigh   time.Time
	DateOfLow    time.Time
}

type Client struct {
	HttpClient tkthttp.Client
}

func NewClient(minInterval time.Duration) *Client {
	return &Client{HttpClient: tkthttp.NewThrottledClient(&http.Client{}, minInterval)}
}

func (c *Client) fetchJSON(p string, rawQuery url.Values) (map[string]json.RawMessage, error) {
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

func (c *Client) fetchPlainText(p string, rawQuery, formValues url.Values) (string, error) {
	if c.HttpClient == nil {
		panic("Client.HttpClient should not be nil")
	}

	u := url.URL{Scheme: "https", Host: "www.tpex.org.tw", Path: p, RawQuery: rawQuery.Encode()}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(formValues.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body), nil
}

func (c *Client) fetchDayQuotes(date time.Time) (map[string]json.RawMessage, error) {
	rawQuery := url.Values{}
	rawQuery.Set("d", fmt.Sprintf("%d/%s", date.Year()-1911, date.Format("01/02")))
	rawQuery.Set("l", "zh-tw")
	return c.fetchJSON("/web/stock/aftertrading/daily_close_quotes/stk_quote_result.php", rawQuery)
}

func (c *Client) fetchDailyQuotes(code string, year int, month time.Month) (map[string]json.RawMessage, error) {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	rawQuery := url.Values{}
	rawQuery.Set("d", fmt.Sprintf("%d/%s", date.Year()-1911, date.Format("01/02")))
	rawQuery.Set("l", "zh-tw")
	rawQuery.Set("stkno", code)
	return c.fetchJSON("/web/stock/aftertrading/daily_trading_info/st43_result.php", rawQuery)
}

func (c *Client) fetchMonthlyQuotes(code string, year int) (string, error) {
	rawQuery := url.Values{}
	rawQuery.Set("l", "en-us")
	formValues := url.Values{}
	formValues.Set("yy", strconv.Itoa(year))
	formValues.Set("stk_no", code)
	return c.fetchPlainText("/web/stock/statistics/monthly/download_st44.php", rawQuery, formValues)
}

func (c *Client) fetchYearlyQuotes(code string) (string, error) {
	rawQuery := url.Values{}
	rawQuery.Set("l", "en-us")
	formValues := url.Values{}
	formValues.Set("stk_no", code)
	return c.fetchPlainText("/web/stock/statistics/monthly/download_st42.php", rawQuery, formValues)
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
			Code:         code,
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

func filterOutInvalidLines(text string) string {
	rawTextSplit := strings.Split(text, "\n")
	dataLines := make([]string, 0)

	isInAddingDataLineStage := false
	for _, line := range rawTextSplit {
		isMatched := invalidCsvChars.MatchString(line)

		if !isInAddingDataLineStage {
			if isMatched {
				continue
			} else {
				isInAddingDataLineStage = true
			}
		} else if isMatched {
			break
		}

		dataLines = append(dataLines, line)
	}

	return strings.Join(dataLines, "\n")
}

func convertRawMonthlyQuote(code string, raw []string) Quote {
	year, err := strconv.Atoi(raw[0])
	if err != nil {
		panic(fmt.Sprintf("failed to parse year %s: %s", raw[0], err))
	}

	month, err := strconv.Atoi(raw[1])
	if err != nil {
		panic(fmt.Sprintf("failed to parse year %s: %s", raw[1], err))
	}

	high, err := strconv.ParseFloat(raw[2], 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse high %s: %s", raw[2], err))
	}

	low, err := strconv.ParseFloat(raw[3], 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse low %s: %s", raw[3], err))
	}

	transactions, err := strconv.ParseUint(strings.ReplaceAll(raw[5], ",", ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse transactions %s: %s", raw[5], err))
	}

	value, err := strconv.ParseUint(strings.ReplaceAll(raw[6], ",", ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse value %s: %s", raw[6], err))
	}

	volume, err := strconv.ParseUint(strings.ReplaceAll(raw[7], ",", ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse volume %s: %s", raw[7], err))
	}

	return Quote{
		Code:         code,
		Date:         time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC),
		Volume:       volume * 1000,
		Transactions: transactions,
		Value:        value * 1000,
		High:         high,
		Low:          low,
	}
}

func (c *Client) FetchMonthlyQuotes(code string, year int) ([]Quote, error) {
	rawText, err := c.fetchMonthlyQuotes(code, year)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(filterOutInvalidLines(rawText)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	qs := make([]Quote, 0)

	for _, record := range records {
		qs = append(qs, convertRawMonthlyQuote(code, record))
	}

	return qs, nil
}

func convertRawYearlyQuote(code string, raw []string) Quote {
	year, err := strconv.Atoi(raw[0])
	if err != nil {
		panic(fmt.Sprintf("failed to parse year %s: %s", raw[0], err))
	}

	volume, err := strconv.ParseUint(strings.ReplaceAll(raw[1], ",", ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse volume %s: %s", raw[1], err))
	}

	value, err := strconv.ParseUint(strings.ReplaceAll(raw[2], ",", ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse value %s: %s", raw[2], err))
	}

	transactions, err := strconv.ParseUint(strings.ReplaceAll(raw[3], ",", ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse transactions %s: %s", raw[3], err))
	}

	high, err := strconv.ParseFloat(raw[4], 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse high %s: %s", raw[4], err))
	}

	dateOfHigh, err := time.Parse("01/02", raw[5])
	if err != nil {
		panic(fmt.Sprintf("failed to parse date of high %s: %s", raw[5], err))
	}

	low, err := strconv.ParseFloat(raw[6], 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse low %s: %s", raw[6], err))
	}

	dateOfLow, err := time.Parse("01/02", raw[7])
	if err != nil {
		panic(fmt.Sprintf("failed to parse date of higlowh %s: %s", raw[7], err))
	}

	return Quote{
		Code:         code,
		Date:         time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
		Volume:       volume * 1000,
		Transactions: transactions * 1000,
		Value:        value * 1000,
		High:         high,
		Low:          low,
		DateOfHigh:   time.Date(year, dateOfHigh.Month(), dateOfHigh.Day(), 0, 0, 0, 0, time.UTC),
		DateOfLow:    time.Date(year, dateOfLow.Month(), dateOfLow.Day(), 0, 0, 0, 0, time.UTC),
	}
}

func (c *Client) FetchYearlyQuotes(code string) ([]Quote, error) {
	rawText, err := c.fetchYearlyQuotes(code)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(filterOutInvalidLines(rawText)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	qs := make([]Quote, 0)

	for _, record := range records {
		qs = append(qs, convertRawYearlyQuote(code, record))
	}

	return qs, nil
}
