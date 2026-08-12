package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const rowsTemplateToken = "{{rows}}"

func splitOutputTemplate(template string) (string, string, error) {
	if strings.TrimSpace(template) == "" {
		template = rowsTemplateToken
	}
	if strings.Count(template, rowsTemplateToken) != 1 {
		return "", "", errors.New("OUTPUT_TEMPLATE_TOKEN_REQUIRED")
	}
	pos := strings.Index(template, rowsTemplateToken)
	inString := false
	escaped := false
	for i := 0; i < pos; i++ {
		c := template[i]
		if escaped {
			escaped = false
			continue
		}
		if inString && c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
		}
	}
	if inString {
		return "", "", errors.New("OUTPUT_TEMPLATE_TOKEN_POSITION")
	}
	prefix := template[:pos]
	suffix := template[pos+len(rowsTemplateToken):]
	if !json.Valid([]byte(prefix + "[]" + suffix)) {
		return "", "", errors.New("OUTPUT_TEMPLATE_INVALID")
	}
	return prefix, suffix, nil
}

func writeOutputTemplate(w io.Writer, template string, writeRows func(io.Writer) (ConvertStats, error)) (ConvertStats, error) {
	prefix, suffix, err := splitOutputTemplate(template)
	if err != nil {
		return ConvertStats{}, err
	}
	if _, err = io.WriteString(w, prefix); err != nil {
		return ConvertStats{}, err
	}
	stats, err := writeRows(w)
	if err != nil {
		return stats, err
	}
	if _, err = io.WriteString(w, suffix); err != nil {
		return stats, err
	}
	return stats, nil
}

func consumeJSONValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return err
			}
			if _, ok := key.(string); !ok {
				return errors.New("invalid object key")
			}
			if err = consumeJSONValue(dec); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return errors.New("invalid object terminator")
		}
	case '[':
		for dec.More() {
			if err = consumeJSONValue(dec); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return errors.New("invalid array terminator")
		}
	default:
		return errors.New("invalid JSON delimiter")
	}
	return nil
}

func validateJSONFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("OUTPUT_TEMPLATE_INVALID: %w", err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.UseNumber()
	if err = consumeJSONValue(dec); err != nil {
		return fmt.Errorf("OUTPUT_TEMPLATE_INVALID: %w", err)
	}
	if _, err = dec.Token(); err != io.EOF {
		if err == nil {
			return errors.New("OUTPUT_TEMPLATE_INVALID")
		}
		return fmt.Errorf("OUTPUT_TEMPLATE_INVALID: %w", err)
	}
	return nil
}
