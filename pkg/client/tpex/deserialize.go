package tpex

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func deserializeSliceOfSlicesOfStrings(rawData map[string]json.RawMessage, key string) [][]string {
	rawItems, ok := rawData[key]
	if !ok {
		panic(fmt.Sprintf("key '%s' does not exist", key))
	}

	var items [][]string
	if err := json.Unmarshal(rawItems, &items); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %s", err))
	}

	return items
}

func deserializeString(rawData map[string]json.RawMessage, key string) string {
	rawItem, ok := rawData[key]
	if !ok {
		panic(fmt.Sprintf("key '%s' does not exist", key))
	}

	var s string
	if err := json.Unmarshal(rawItem, &s); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %s", err))
	}

	return s
}

func stringToUint64(s string) uint64 {
	v, err := strconv.ParseUint(strings.Replace(s, ",", "", -1), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("value %v is not uint64: %s", s, err))
	}

	return v
}

func stringToFloat64(s string) float64 {
	if strings.TrimSpace(s) == "---" {
		return 0
	}

	v, err := strconv.ParseFloat(strings.Replace(s, ",", "", -1), 64)
	if err != nil {
		panic(fmt.Sprintf("value %v is not float64: %s", s, err))
	}

	return v
}

func stringToDate(s string) time.Time {
	rawDate := strings.SplitN(s, "/", 2)
	if len(rawDate) != 2 {
		panic(fmt.Sprintf("the format of '%s' is unexpected", s))
	}

	rocYear, err := strconv.ParseInt(rawDate[0], 0, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s to int: %v", rawDate[0], err))
	}

	t, err := time.Parse("01/02", rawDate[1])
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s to month and day: %v", rawDate[1], err))
	}

	return time.Date(int(rocYear+1911), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
