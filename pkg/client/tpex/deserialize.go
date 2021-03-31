package tpex

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
