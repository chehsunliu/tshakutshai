package tkttest

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

func newResponseFromGzipFile(filepath string, statusCode int, header http.Header) *http.Response {
	f, err := os.Open(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to open %s: %w", filepath, err))
	}

	reader, err := gzip.NewReader(f)
	if err != nil {
		panic(fmt.Errorf("failed to create GZIP reader: %w", err))
	}

	return &http.Response{
		Body:       reader,
		StatusCode: statusCode,
		Header:     header,
	}
}

func NewJsonResponseFromGzipFile(filepath string, statusCode int) *http.Response {
	header := http.Header{}
	header.Set("content-type", "application/json; charset=utf-8")
	return newResponseFromGzipFile(filepath, statusCode, header)
}

func NewResponseFromGzipFile(filepath string, statusCode int) *http.Response {
	return newResponseFromGzipFile(filepath, statusCode, http.Header{})
}

func NewResponseFromString(content string, statusCode int) *http.Response {
	header := http.Header{}
	header.Set("content-type", "application/json; charset=utf-8")

	return &http.Response{
		Body:       io.NopCloser(strings.NewReader(content)),
		StatusCode: statusCode,
		Header:     header,
	}
}
