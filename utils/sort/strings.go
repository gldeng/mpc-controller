package sort

import (
	"sort"
	"strings"
)

func SortMultiLineStr(s string) string {
	ss := strings.Split(s, "\n")
	sort.Strings(ss)
	news := strings.Join(ss, "\n")
	return news
}
