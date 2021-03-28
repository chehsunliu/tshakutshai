package twse

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	r0, r1 := args.Get(0), args.Error(1)
	if r0 == nil {
		return nil, r1
	} else {
		return r0.(*http.Response), r1
	}
}

func NewResponseFromFile(filepath string, statusCode int) *http.Response {
	f, err := os.Open(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to open %s: %w", filepath, err))
	}

	reader, err := gzip.NewReader(f)
	if err != nil {
		panic(fmt.Errorf("failed to create GZIP reader: %w", err))
	}

	return &http.Response{Body: reader, StatusCode: statusCode}
}

func TestClient_FetchDayQuotes(t *testing.T) {
	date := time.Date(2021, 3, 24, 0, 0, 0, 0, time.UTC)

	mockResponse := NewResponseFromFile("./testdata/quotes-tw-20210324.json.gz", 200)
	mockHttpClient := &MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/exchangeReport/MI_INDEX" && u.Query().Get("date") == "20210324"
	})).Return(mockResponse, nil)

	client := &Client{http: mockHttpClient}
	quotes, err := client.FetchDayQuotes(date)

	assert.Nilf(t, err, "%+v", err)
	assert.Greater(t, len(quotes), 20000)

	assert.Equal(t, Quote{
		Code:         "0050",
		Name:         "元大台灣50",
		Date:         date,
		Volume:       11_082_813,
		Transactions: 20_959,
		Value:        1_459_923_222,
		Open:         131.80,
		High:         132.45,
		Low:          131.30,
		Close:        131.50,
	}, quotes["0050"])

	assert.Equal(t, Quote{
		Code:         "2330",
		Name:         "台積電",
		Date:         date,
		Volume:       115_318_351,
		Transactions: 242_138,
		Value:        66_559_451_738,
		Open:         571.00,
		High:         582.00,
		Low:          571.00,
		Close:        576.00,
	}, quotes["2330"])

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}

func TestClient_FetchDailyQuotes(t *testing.T) {
	code := "2330"
	date := time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)

	mockResponse := NewResponseFromFile("./testdata/quotes-tw-202102-2330.json.gz", 200)
	mockHttpClient := &MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/exchangeReport/STOCK_DAY" &&
			u.Query().Get("date") == "20210201" &&
			u.Query().Get("stockNo") == code
	})).Return(mockResponse, nil)

	client := &Client{http: mockHttpClient}
	quotes, err := client.FetchDailyQuotes(code, date.Year(), date.Month())

	assert.Nilf(t, err, "%+v", err)
	assert.Greater(t, len(quotes), 10)

	assert.Equal(t, Quote{
		Code:         "2330",
		Name:         "",
		Date:         date,
		Volume:       70_161_939,
		Transactions: 81_346,
		Value:        42_004_241_697,
		Open:         595.00,
		High:         612.00,
		Low:          587.00,
		Close:        611.00,
	}, quotes[0])

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}

func TestClient_FetchMonthlyQuotes(t *testing.T) {
	code := "2454"
	year := 2020

	mockResponse := NewResponseFromFile("./testdata/quotes-tw-2020-2454.json.gz", 200)
	mockHttpClient := &MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/exchangeReport/FMSRFK" &&
			u.Query().Get("date") == "20200101" &&
			u.Query().Get("stockNo") == code
	})).Return(mockResponse, nil)

	client := &Client{http: mockHttpClient}
	qs, err := client.FetchMonthlyQuotes(code, year)

	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 12, len(qs))

	assert.Equal(t, MonthlyQuote{
		Code:         code,
		Year:         year,
		Month:        time.April,
		Volume:       218_553_058,
		Transactions: 146_711,
		Value:        80_262_421_295,
		High:         415.50,
		Low:          325.50,
	}, qs[3])

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}

func TestClient_FetchYearlyQuotes(t *testing.T) {
	code := "0050"

	mockResponse := NewResponseFromFile("./testdata/quotes-tw-0050.json.gz", 200)
	mockHttpClient := &MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/exchangeReport/FMNPTK" && u.Query().Get("stockNo") == code
	})).Return(mockResponse, nil)

	client := &Client{http: mockHttpClient}
	qs, err := client.FetchYearlyQuotes(code)

	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 18, len(qs))

	assert.Equal(t, YearlyQuote{
		Code:         code,
		Year:         2020,
		Volume:       2_564_396_277,
		Value:        234_459_163_641,
		Transactions: 1_413_186,
		High:         122.40,
		Low:          67.25,
		DateOfHigh:   time.Date(2020, time.December, 31, 0, 0, 0, 0, time.UTC),
		DateOfLow:    time.Date(2020, time.March, 19, 0, 0, 0, 0, time.UTC),
	}, qs[17])

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}
