package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type flatJSONObject struct {
	Keys   []string
	Values map[string]string
}

func csvDelimiterFromName(name string) rune {
	switch name {
	case "tab":
		return '\t'
	case "pipe":
		return '|'
	case "semicolon":
		return ';'
	default:
		return ','
	}
}

func scalarCSVValue(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case json.Number:
		return v.String(), nil
	default:
		return "", errors.New("NESTED_JSON_NOT_SUPPORTED")
	}
}

func decodeFlatJSONObject(dec *json.Decoder) (flatJSONObject, error) {
	tok, err := dec.Token()
	if err != nil {
		return flatJSONObject{}, err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return flatJSONObject{}, errors.New("JSON_OBJECT_REQUIRED")
	}

	obj := flatJSONObject{Values: make(map[string]string)}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return flatJSONObject{}, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return flatJSONObject{}, errors.New("JSON_PARSE_FAILED")
		}
		var value any
		if err = dec.Decode(&value); err != nil {
			return flatJSONObject{}, err
		}
		text, err := scalarCSVValue(value)
		if err != nil {
			return flatJSONObject{}, err
		}
		obj.Keys = append(obj.Keys, key)
		obj.Values[key] = text
	}
	if _, err = dec.Token(); err != nil {
		return flatJSONObject{}, err
	}
	return obj, nil
}

func walkJSONObjects(r io.Reader, visit func(flatJSONObject) error) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	first, err := dec.Token()
	if err != nil {
		return fmt.Errorf("JSON_PARSE_FAILED: %w", err)
	}

	if delim, ok := first.(json.Delim); ok && delim == '[' {
		for dec.More() {
			obj, err := decodeFlatJSONObject(dec)
			if err != nil {
				return err
			}
			if err = visit(obj); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return fmt.Errorf("JSON_PARSE_FAILED: %w", err)
		}
		if end != json.Delim(']') {
			return errors.New("JSON_PARSE_FAILED")
		}
	} else if delim, ok := first.(json.Delim); ok && delim == '{' {
		obj := flatJSONObject{Values: make(map[string]string)}
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return fmt.Errorf("JSON_PARSE_FAILED: %w", err)
			}
			key, ok := keyTok.(string)
			if !ok {
				return errors.New("JSON_PARSE_FAILED")
			}
			var value any
			if err = dec.Decode(&value); err != nil {
				return fmt.Errorf("JSON_PARSE_FAILED: %w", err)
			}
			text, err := scalarCSVValue(value)
			if err != nil {
				return err
			}
			obj.Keys = append(obj.Keys, key)
			obj.Values[key] = text
		}
		if _, err = dec.Token(); err != nil {
			return fmt.Errorf("JSON_PARSE_FAILED: %w", err)
		}
		if err = visit(obj); err != nil {
			return err
		}
	} else {
		return errors.New("JSON_TOP_LEVEL_OBJECT_OR_ARRAY_REQUIRED")
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("JSON_PARSE_FAILED")
		}
		return fmt.Errorf("JSON_PARSE_FAILED: %w", err)
	}
	return nil
}

func collectJSONHeaders(r io.Reader) ([]string, int, error) {
	headers := make([]string, 0)
	seen := map[string]struct{}{}
	rows := 0
	err := walkJSONObjects(r, func(obj flatJSONObject) error {
		rows++
		for _, key := range obj.Keys {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			headers = append(headers, key)
		}
		return nil
	})
	return headers, rows, err
}

func convertJSONToDelimited(r io.Reader, w io.Writer, headers []string, delim rune) (ConvertStats, error) {
	writer := csv.NewWriter(w)
	writer.Comma = delim
	if err := writer.Write(headers); err != nil {
		return ConvertStats{}, err
	}
	stats := ConvertStats{Columns: len(headers), Delimiter: delim}
	err := walkJSONObjects(r, func(obj flatJSONObject) error {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = obj.Values[header]
		}
		if err := writer.Write(row); err != nil {
			return err
		}
		stats.Rows++
		return nil
	})
	writer.Flush()
	if err != nil {
		return stats, err
	}
	if err = writer.Error(); err != nil {
		return stats, err
	}
	return stats, nil
}
