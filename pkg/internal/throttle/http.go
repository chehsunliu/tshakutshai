package throttle

import (
	"net/http"
	"sync"
	"time"
)

type HttpClient struct {
	*http.Client

	minInterval time.Duration
	mutex       sync.Mutex
	last        time.Time
}

func NewHttpClient(minInterval time.Duration) *HttpClient {
	return &HttpClient{
		Client:      &http.Client{},
		minInterval: minInterval,
	}
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	elapsed := time.Now().Sub(c.last)
	if elapsed < c.minInterval {
		time.Sleep(c.minInterval - elapsed)
	}

	resp, err := c.Client.Do(req)
	c.last = time.Now()
	return resp, err
}
