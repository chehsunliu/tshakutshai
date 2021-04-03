package twse

import "fmt"

// noDataError is an error returned by Fetch functions when query conditions matches nothing. It can be
// no such stock symbol, dates before the stock's IPO, or the TWSE server did not have any data of all
// the stocks before the query date.
type noDataError struct {
	Message string
}

func (e *noDataError) Error() string {
	return fmt.Sprintf("NoData: %s", e.Message)
}

// QuotaExceededError is an error returned by Fetch functions when query the TWSE server too frequently.
// Typically it takes around 1 hour to get back to normal.
type QuotaExceededError struct {
	Message string
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("QuotaExceeded: %s", e.Message)
}

// ConnectionError is an error returned by Fetch functions when having problem to connect to the TWSE server.
type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("ConnectionError: %s", e.Message)
}
