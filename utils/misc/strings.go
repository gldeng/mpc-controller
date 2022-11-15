package misc

import "strings"

func ExtractColumIntoString(s string, sep string, col int) string {
	lines := strings.Split(s, "\n")
	var r string
	for _, line := range lines {
		fields := strings.Split(line, sep)
		r += fields[col] + " "
	}
	r = strings.TrimRight(r, " ")
	return r
}
