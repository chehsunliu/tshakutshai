package twse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chehsunliu/tshakutshai/pkg/client/twse"
)

var client = twse.NewClient(time.Second * 3)

func TestClient_FetchDayQuotes(t *testing.T) {
	date := time.Date(2021, time.March, 29, 0, 0, 0, 0, time.UTC)

	quotes, err := client.FetchDayQuotes(date)
	assert.Nilf(t, err, "%+v", err)
	assert.Greater(t, len(quotes), 20000)

	q2330 := quotes["2330"]
	assert.Equal(t, "2330", q2330.Code)
	assert.Equal(t, "台積電", q2330.Name)
	assert.Equal(t, "20210329", q2330.Date.Format("20060102"))
	assert.Equal(t, uint64(40_573_913), q2330.Volume)
	assert.Equal(t, uint64(44_500), q2330.Transactions)
	assert.Equal(t, uint64(24_302_039_570), q2330.Value)
	assert.Equal(t, 602.00, q2330.High)
	assert.Equal(t, 596.00, q2330.Low)
	assert.Equal(t, 599.00, q2330.Open)
	assert.Equal(t, 599.00, q2330.Close)
	assert.Equal(t, time.Time{}, q2330.DateOfHigh)
	assert.Equal(t, time.Time{}, q2330.DateOfLow)

	q2454 := quotes["2454"]
	assert.Equal(t, "聯發科", q2454.Name)
	assert.Equal(t, 941.00, q2454.Close)

	q00684R := quotes["00684R"]
	assert.Equal(t, "00684R", q00684R.Code)
	assert.Equal(t, "期元大美元指反1", q00684R.Name)
	assert.Equal(t, "20210329", q00684R.Date.Format("20060102"))
	assert.Equal(t, uint64(0), q00684R.Volume)
	assert.Equal(t, uint64(0), q00684R.Transactions)
	assert.Equal(t, uint64(0), q00684R.Value)
	assert.Equal(t, 0.00, q00684R.High)
	assert.Equal(t, 0.00, q00684R.Low)
	assert.Equal(t, 0.00, q00684R.Open)
	assert.Equal(t, 0.00, q00684R.Close)
}

func TestClient_FetchDailyQuotes(t *testing.T) {
	quotes, err := client.FetchDailyQuotes("2330", 2021, time.February)
	assert.Nilf(t, err, "%+v", err)

	q0 := quotes[0]
	assert.Equal(t, "2330", q0.Code)
	assert.Equal(t, "", q0.Name)
	assert.Equal(t, "20210201", q0.Date.Format("20060102"))
	assert.Equal(t, uint64(70_161_939), q0.Volume)
	assert.Equal(t, uint64(81_346), q0.Transactions)
	assert.Equal(t, uint64(42_004_241_697), q0.Value)
	assert.Equal(t, 612.00, q0.High)
	assert.Equal(t, 587.00, q0.Low)
	assert.Equal(t, 595.00, q0.Open)
	assert.Equal(t, 611.00, q0.Close)

	dates := make([]string, 0)
	prices := make([]float64, 0)

	for _, q := range quotes {
		dates = append(dates, q.Date.Format("20060102"))
		prices = append(prices, q.Close)
	}

	assert.Equal(t, []string{
		"20210201",
		"20210202",
		"20210203",
		"20210204",
		"20210205",
		"20210217",
		"20210218",
		"20210219",
		"20210222",
		"20210223",
		"20210224",
		"20210225",
		"20210226",
	}, dates)
	assert.Equal(t, []float64{
		611.00,
		632.00,
		630.00,
		627.00,
		632.00,
		663.00,
		660.00,
		652.00,
		650.00,
		641.00,
		625.00,
		635.00,
		606.00,
	}, prices)
}

func TestClient_FetchMonthlyQuotes(t *testing.T) {
	quotes, err := client.FetchMonthlyQuotes("0050", 2020)
	assert.Nilf(t, err, "%+v", err)
	assert.Equal(t, 12, len(quotes))

	q12 := quotes[11]
	assert.Equal(t, "0050", q12.Code)
	assert.Equal(t, "", q12.Name)
	assert.Equal(t, "20201201", q12.Date.Format("20060102"))
	assert.Equal(t, uint64(13_1351_140), q12.Volume)
	assert.Equal(t, uint64(113_776), q12.Transactions)
	assert.Equal(t, uint64(15_518_022_408), q12.Value)
	assert.Equal(t, 122.40, q12.High)
	assert.Equal(t, 113.35, q12.Low)
	assert.Equal(t, 0.00, q12.Open)
	assert.Equal(t, 0.00, q12.Close)
	assert.Equal(t, time.Time{}, q12.DateOfHigh)
	assert.Equal(t, time.Time{}, q12.DateOfLow)
}

func TestClient_FetchYearlyQuotes(t *testing.T) {
	quotes, err := client.FetchYearlyQuotes("2412")
	assert.Nilf(t, err, "%+v", err)
	assert.Greater(t, len(quotes), 20)

	q := quotes[19]
	assert.Equal(t, "2412", q.Code)
	assert.Equal(t, "", q.Name)
	assert.Equal(t, "20190101", q.Date.Format("20060102"))
	assert.Equal(t, uint64(1_679_098_996), q.Volume)
	assert.Equal(t, uint64(765_529), q.Transactions)
	assert.Equal(t, uint64(185_060_260_891), q.Value)
	assert.Equal(t, 114.00, q.High)
	assert.Equal(t, 106.00, q.Low)
	assert.Equal(t, 0.00, q.Open)
	assert.Equal(t, 0.00, q.Close)
	assert.Equal(t, "20191126", q.DateOfHigh.Format("20060102"))
	assert.Equal(t, "20190222", q.DateOfLow.Format("20060102"))
}
