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

func TestClient_GetQuotesOfDay(t *testing.T) {
	date := time.Date(2021, 3, 25, 0, 0, 0, 0, time.UTC)

	mockResponse := NewResponseFromFile("./testdata/quotes-en-20210325.json.gz", 200)
	mockHttpClient := &MockHttpClient{}
	mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		u := req.URL
		return u.Path == "/en/exchangeReport/MI_INDEX" && u.Query().Get("date") == "20210325"
	})).Return(mockResponse, nil)

	client := &Client{http: mockHttpClient}
	_, err := client.GetQuotesOfDay(date)

	assert.Nil(t, err)
	mockHttpClient.AssertNumberOfCalls(t, "Do", 1)
}
