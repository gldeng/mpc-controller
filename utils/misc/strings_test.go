package misc

import (
	"fmt"
	"testing"
)

const text = `avalido/mpc-controller                          latest               1dd8f4ee9838   26 hours ago    44MB
<none>                                          <none>               63a7a46800a8   27 hours ago    1.21GB
<none>                                          <none>               6a944206581d   27 hours ago    1.21GB
vektra/mockery                                  latest               0c365eaaf3e4   2 weeks ago     337MB
`

func TestExtractColumIntoString(t *testing.T) {
	col := ExtractColumIntoString(text, " ", 2)
	fmt.Println(col)
}

func TestExtractColum(t *testing.T) {
	col := ExtractColum(text, " ", 2)
	fmt.Printf("Len:%v Val:%v", len(col), col)
}

func TestSplitText(t *testing.T) {
	lines := SplitText(text, " ")
	for _, line := range lines {
		fmt.Printf("Len:%v, Val:%v\n", len(line), line)
	}
}
