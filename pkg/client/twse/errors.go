package twse

import "fmt"

type NoDataError struct {
	Message string
}

func (e *NoDataError) Error() string {
	return fmt.Sprintf("NoData: %s", e.Message)
}

type QuotaExceededError struct {
	Message string
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("QuotaExceeded: %s", e.Message)
}

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("ConnectionError: %s", e.Message)
}
