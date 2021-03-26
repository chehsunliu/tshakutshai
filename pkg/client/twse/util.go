package twse

import (
	"encoding/json"
	"fmt"
)

func retrieveFields(raw map[string]json.RawMessage, key string) ([]string, error) {
	rawFields, ok := raw[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' does not exist", key)
	}

	var fields []string
	if err := json.Unmarshal(rawFields, &fields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return fields, nil
}

func retrieveItems(raw map[string]json.RawMessage, key string) ([][]interface{}, error) {
	rawItems, ok := raw[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' does not exist", key)
	}

	var items [][]interface{}
	if err := json.Unmarshal(rawItems, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return items, nil
}

func zipFieldsAndItems(fields []string, items [][]interface{}) []map[string]interface{} {
	rawRecords := make([]map[string]interface{}, len(items))

	for i := range rawRecords {

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
