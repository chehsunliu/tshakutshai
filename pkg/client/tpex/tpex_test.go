package tpex_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chehsunliu/tshakutshai/pkg/client/tpex"
	"github.com/chehsunliu/tshakutshai/pkg/internal/tkttest"
)

func TestClient_FetchDayQuotes(t *testing.T) {
	date := time.Date(2021, 3, 30, 0, 0, 0, 0, time.UTC)

	mockResponse := tkttest.NewJsonResponseFromGzipFile("./testdata/quotes-tw-20210330.json.gz", 200)
	mockHttpClient := &tkttest.MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/web/stock/aftertrading/daily_close_quotes/stk_quote_result.php" &&
			u.Query().Get("d") == "110/03/30"
	})).Return(mockResponse, nil)

	client := &tpex.Client{HttpClient: mockHttpClient}
	qs, err := client.FetchDayQuotes(date)

	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 6877, len(qs))

	q := qs["006201"]
	assert.Equal(t, "006201", q.Code)
	assert.Equal(t, "元大富櫃50", q.Name)
	assert.Equal(t, "20210330", q.Date.Format("20060102"))
	assert.Equal(t, uint64(54_765), q.Volume)
	assert.Equal(t, uint64(33), q.Transactions)
	assert.Equal(t, uint64(1_062_607), q.Value)
	assert.Equal(t, 19.51, q.High)
	assert.Equal(t, 19.32, q.Low)
	assert.Equal(t, 19.37, q.Open)
	assert.Equal(t, 19.49, q.Close)

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}

func TestClient_FetchDailyQuotes(t *testing.T) {
	code := "8044"
	date := time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)

	mockResponse := tkttest.NewJsonResponseFromGzipFile("./testdata/quotes-tw-202102-8044.json.gz", 200)
	mockHttpClient := &tkttest.MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/web/stock/aftertrading/daily_trading_info/st43_result.php" &&
			u.Query().Get("d") == "110/02/01" &&
			u.Query().Get("stkno") == code
	})).Return(mockResponse, nil)

	client := &tpex.Client{HttpClient: mockHttpClient}
	qs, err := client.FetchDailyQuotes(code, date.Year(), date.Month())
	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 13, len(qs))

	q := qs[11]
	assert.Equal(t, code, q.Code)
	assert.Equal(t, "網家", q.Name)
	assert.Equal(t, "20210225", q.Date.Format("20060102"))
	assert.Equal(t, uint64(834_000), q.Volume)
	assert.Equal(t, uint64(780), q.Transactions)
	assert.Equal(t, uint64(69_098_000), q.Value)
	assert.Equal(t, 84.00, q.High)
	assert.Equal(t, 82.20, q.Low)
	assert.Equal(t, 83.20, q.Open)
	assert.Equal(t, 82.30, q.Close)

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}

func TestClient_FetchMonthlyQuotes(t *testing.T) {
	code := "8044"
	year := 2020

	mockResponse := tkttest.NewResponseFromGzipFile("./testdata/quotes-tw-2020-8044.csv.gz", 200)
	mockHttpClient := &tkttest.MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.Path == "/web/stock/statistics/monthly/download_st44.php"
	})).Return(mockResponse, nil)

	client := &tpex.Client{HttpClient: mockHttpClient}
	qs, err := client.FetchMonthlyQuotes(code, year)
	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 12, len(qs))

	q := qs[0]
	assert.Equal(t, code, q.Code)
	assert.Equal(t, "", q.Name)
	assert.Equal(t, "20200101", q.Date.Format("20060102"))
	assert.Equal(t, uint64(6_092_000), q.Volume)
	assert.Equal(t, uint64(5_274), q.Transactions)
	assert.Equal(t, uint64(564_646_000), q.Value)
	assert.Equal(t, 96.40, q.High)
	assert.Equal(t, 88.70, q.Low)
}

func TestClient_FetchYearlyQuotes(t *testing.T) {
	code := "8044"

	mockResponse := tkttest.NewResponseFromGzipFile("./testdata/quotes-tw-8044.csv.gz", 200)
	mockHttpClient := &tkttest.MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.Path == "/web/stock/statistics/monthly/download_st42.php"
	})).Return(mockResponse, nil)

	client := &tpex.Client{HttpClient: mockHttpClient}
	qs, err := client.FetchYearlyQuotes(code)
	assert.Nilf(t, err, "%s", err)
	assert.Equal(t, 17, len(qs))

	q := qs[16]
	assert.Equal(t, code, q.Code)
	assert.Equal(t, "", q.Name)
	assert.Equal(t, "20050101", q.Date.Format("20060102"))
	assert.Equal(t, uint64(296_356_000), q.Volume)
	assert.Equal(t, uint64(147_000), q.Transactions)
	assert.Equal(t, uint64(14_075_258_000), q.Value)
	assert.Equal(t, 59.70, q.High)
	assert.Equal(t, 28.20, q.Low)
	assert.Equal(t, time.Date(2005, time.September, 16, 0, 0, 0, 0, time.UTC), q.DateOfHigh)
	assert.Equal(t, time.Date(2005, time.January, 24, 0, 0, 0, 0, time.UTC), q.DateOfLow)
}
