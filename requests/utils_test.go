package requests

import (
	"fmt"
	"testing"
)

func TestUrlParse(t *testing.T) {

	// c := `http://www.gxwmhq.com/search?a=tes\%dada&b=YWRzZGFzYQ==`

	c := `https://yuntan.360.cn/monitor/index?token=UlNKbEVrMmxUSHVhWXgwN3R1eXJUOGxpUkJyWXlsTmNnOFElMkJybXh4VHpVJTNE`

	u, err := UrlParse(c)
	fmt.Println(u.RawQuery, err)
}
