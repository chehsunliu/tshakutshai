package twse

import (
	"encoding/json"
	"fmt"
)

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

func zipFieldsAndItems(fields []string, items [][]interface{}) ([]map[string]interface{}, error) {
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
