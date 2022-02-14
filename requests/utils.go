package requests

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// UrlParse 调用url.Parse，增加了对%的处理
func UrlParse(sourceUrl string) (*url.URL, error) {
	u, err := url.Parse(sourceUrl)
	if err != nil {
		u, err = url.Parse(escapePercentSign(sourceUrl))
	}
	if err != nil {
		return nil, errors.Wrap(err, "parse url error")
	}
	return u, nil
}

// escapePercentSign 把url中的%替换为%25
func escapePercentSign(raw string) string {
	s := strings.ReplaceAll(raw, "%", "%25")

	return s
}
