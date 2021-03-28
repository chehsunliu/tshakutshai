package twse

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func isStatOK(rawData map[string]json.RawMessage) (bool, error) {
	rawStat, ok := rawData["stat"]
	if !ok {
		return false, fmt.Errorf("key 'stat' does not exist")
	}

	var stat string
	if err := json.Unmarshal(rawStat, &stat); err != nil {
		return false, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return stat == "OK", nil
}

func retrieveFields(rawData map[string]json.RawMessage, key string) ([]string, error) {
	rawFields, ok := rawData[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' does not exist", key)
	}

	var fields []string
	if err := json.Unmarshal(rawFields, &fields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return fields, nil
}

func retrieveItems(rawData map[string]json.RawMessage, key string) ([][]interface{}, error) {
	rawItems, ok := rawData[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' does not exist", key)
	}

	var items [][]interface{}
	if err := json.Unmarshal(rawItems, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return items, nil
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

func zipFieldsAndItems(rawData map[string]json.RawMessage, fieldsKey, itemsKey string) ([]map[string]interface{}, error) {
	fields, err := retrieveFields(rawData, fieldsKey)
	if err != nil {
		return nil, err
	}

	items, err := retrieveItems(rawData, itemsKey)
	if err != nil {
		return nil, err
	}

	// The TWSE uses the same field name to denote the days of the highest and
	// the lowest prices in yearly quotes. Here I just made the second appearance
	// to be 'name2', the third one to be 'name3' and so on.
	fields = suffixDuplicateFields(fields)
	rawRecords := make([]map[string]interface{}, len(items))

	for i := range rawRecords {
		if len(fields) != len(items[i]) {
			return nil, fmt.Errorf(
				"fields %v has %d elements but item %v has only %d",
				fields, len(fields), items[i], len(items[i]),
			)
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

	return rawRecords, nil
}

func convertToString(rawQuote map[string]interface{}, field string) (string, error) {
	i, ok := rawQuote[field]
	if !ok {
		return "", fmt.Errorf("field '%s' does not exist in %v", field, rawQuote)
	}

	s, ok := i.(string)
	if !ok {
		return "", fmt.Errorf("value %v of field '%s' in %v is not string", i, field, rawQuote)
	}

	return s, nil
}

func convertToFloat64(rawQuote map[string]interface{}, field string) (float64, error) {
	i, ok := rawQuote[field]
	if !ok {
		return 0, fmt.Errorf("field '%s' does not exist in %v", field, rawQuote)
	}

	f, ok := i.(float64)
	if !ok {
		return 0, fmt.Errorf(
			"value %v of field '%s' in %v is not int but %s", i, field, rawQuote, reflect.TypeOf(i))
	}

	return f, nil
}

func convertToStringThenUint64(rawQuote map[string]interface{}, field string) (uint64, error) {
	s, err := convertToString(rawQuote, field)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseUint(strings.Replace(s, ",", "", -1), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("value %v of field '%s' in %v is not uint64: %w", v, field, rawQuote, err)
	}

	return v, nil
}

func convertToStringThenFloat64(rawQuote map[string]interface{}, field string) (float64, error) {
	s, err := convertToString(rawQuote, field)
	if err != nil {
		return 0, err
	}

	// If a stock have no transactions made, its 4 prices will be '--'.
	if s == "--" {
		return 0, nil
	}

	v, err := strconv.ParseFloat(strings.Replace(s, ",", "", -1), 64)
	if err != nil {
		return 0, fmt.Errorf("value %v of field '%s' in %v is not float64: %w", v, field, rawQuote, err)
	}

	return v, nil
}
