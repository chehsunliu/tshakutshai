package http

import (
	"net/http"
	"sync"
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
	}

	return r0.(*http.Response), r1
}

func TestThrottledClient_Do(t *testing.T) {
	minInterval := time.Millisecond * 250
	mockHttpClient := &MockHttpClient{}
	mockHttpClient.On("Do", mock.Anything).Return(nil, nil)

	client := NewThrottledClient(mockHttpClient, minInterval)
	t0 := time.Now()

	var wg sync.WaitGroup
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func() {
			client.Do(&http.Request{})
			wg.Done()
		}()
	}
	wg.Wait()

	elapsed := time.Now().Sub(t0)
	assert.True(t, elapsed >= minInterval*(4-1))
	mockHttpClient.AssertNumberOfCalls(t, "Do", 4)
}
