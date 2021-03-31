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
	_, err := client.FetchDayQuotes(date)

	assert.Nilf(t, err, "%+v", err)

	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}
