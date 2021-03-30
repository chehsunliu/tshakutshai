package http

import (
	"net/http"
	"sync"
	"time"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type ThrottledClient struct {
	*http.Client

	minInterval time.Duration
	mutex       sync.Mutex
	last        time.Time
}

func NewThrottledClient(minInterval time.Duration) *ThrottledClient {
	return &ThrottledClient{
		Client:      &http.Client{},
		minInterval: minInterval,
	}
}

func (c *ThrottledClient) Do(req *http.Request) (*http.Response, error) {
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
