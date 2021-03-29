package twse

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func retrieveStat(rawData map[string]json.RawMessage) string {
	rawStat, ok := rawData["stat"]
	if !ok {
		panic("key 'stat' does not exist")
	}

	var stat string
	if err := json.Unmarshal(rawStat, &stat); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %s", err))
	}

	return stat
}

func retrieveFields(rawData map[string]json.RawMessage, key string) []string {
	rawFields, ok := rawData[key]
	if !ok {
		panic(fmt.Sprintf("key '%s' does not exist", key))
	}

	var fields []string
	if err := json.Unmarshal(rawFields, &fields); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %s", err))
	}

	return fields
}

func retrieveItems(rawData map[string]json.RawMessage, key string) [][]interface{} {
	rawItems, ok := rawData[key]
	if !ok {
		panic(fmt.Sprintf("key '%s' does not exist", key))
	}

	var items [][]interface{}
	if err := json.Unmarshal(rawItems, &items); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %s", err))
	}

	return items
}

func suffixDuplicateFields(fields []string) []string {
	fixed := map[string]int{}
	unfixed := map[string]int{}
	for _, field := range fields {
		fixed[field]++
		unfixed[field]++
	}

	for i := range fields {
		if fixed[fields[i]] == 1 {
			continue
		}

		diff := fixed[fields[i]] - unfixed[fields[i]]
		unfixed[fields[i]]--

		if diff > 0 {
			fields[i] = fmt.Sprintf("%s%d", fields[i], diff+1)
		}
	}

	return fields
}

func zipFieldsAndItems(rawData map[string]json.RawMessage, fieldsKey, itemsKey string) []map[string]interface{} {
	fields := retrieveFields(rawData, fieldsKey)
	items := retrieveItems(rawData, itemsKey)

	// The TWSE uses the same field name to denote the days of the highest and
	// the lowest prices in yearly quotes. Here I just made the second appearance
	// to be 'name2', the third one to be 'name3' and so on.
	fields = suffixDuplicateFields(fields)
	rawRecords := make([]map[string]interface{}, len(items))

	for i := range rawRecords {
		if len(fields) != len(items[i]) {
			panic(fmt.Sprintf("fields %v has %d elements but item %v has only %d", fields, len(fields), items[i], len(items[i])))
		}

		rawRecord := map[string]interface{}{}
		for j := range fields {
			if j >= len(items[i]) {
				break
			}

			rawRecord[fields[j]] = items[i][j]
		}

		rawRecords[i] = rawRecord
	}

	return rawRecords
}

func convertToString(rawQuote map[string]interface{}, field string) string {
	i, ok := rawQuote[field]
	if !ok {
		panic(fmt.Sprintf("field '%s' does not exist in %v", field, rawQuote))
	}

	s, ok := i.(string)
	if !ok {
		panic(fmt.Sprintf("value %v of field '%s' in %v is not string", i, field, rawQuote))
	}

	return s
}

func convertToFloat64(rawQuote map[string]interface{}, field string) float64 {
	i, ok := rawQuote[field]
	if !ok {
		panic(fmt.Sprintf("field '%s' does not exist in %v", field, rawQuote))
	}

	f, ok := i.(float64)
	if !ok {
		panic(fmt.Sprintf("value %v of field '%s' in %v is not int but %s", i, field, rawQuote, reflect.TypeOf(i)))
	}

	return f
}

func convertToStringThenUint64(rawQuote map[string]interface{}, field string) uint64 {
	s := convertToString(rawQuote, field)

	v, err := strconv.ParseUint(strings.Replace(s, ",", "", -1), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("value %v of field '%s' in %v is not uint64: %s", v, field, rawQuote, err))
	}

	return v
}

func convertToStringThenFloat64(rawQuote map[string]interface{}, field string) float64 {
	s := convertToString(rawQuote, field)

	// If a stock have no transactions made, its 4 prices will be '--'.
	if s == "--" {
		return 0
	}

	v, err := strconv.ParseFloat(strings.Replace(s, ",", "", -1), 64)
	if err != nil {
		panic(fmt.Sprintf("value %v of field '%s' in %v is not float64: %s", v, field, rawQuote, err))
	}

	return v
}
