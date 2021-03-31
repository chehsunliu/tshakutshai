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

	mockResponse := tkttest.NewResponseFromFile("./testdata/quotes-tw-20210330.json.gz", 200)
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
