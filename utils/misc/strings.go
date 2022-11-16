package misc

import "strings"

func ExtractColumIntoString(s string, sep string, col int) string {
	cols := ExtractColum(s, sep, col)
	var r string
	for _, field := range cols {
		r += field + " "
	}
	return strings.TrimSpace(r)
}

func ExtractColum(s string, sep string, col int) []string {
	allFields := SplitText(s, sep)
	var fields_ []string
	for _, fields := range allFields {
		if len(fields) <= col {
			continue
		}
		fields_ = append(fields_, fields[col])
	}
	return fields_
}

func SplitText(s string, sep string) [][]string {
	var allFields [][]string
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		var lineFields []string
		fields := strings.Split(line, sep)
		for _, field := range fields {
			field := strings.TrimSpace(field)
			if field == "" {
				continue
			}
			lineFields = append(lineFields, field)
		}
		if len(lineFields) == 0 {
			continue
		}
		allFields = append(allFields, lineFields)
	}
	return allFields
}
