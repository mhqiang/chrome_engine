package chrome_engine

import (
	"chrome_engine/config"
	"chrome_engine/model"
	"net/http"
	"net/url"
	"testing"

	"github.com/mhqiang/logger"
)

func initLog() {
	var config logger.Config

	config.MaxLogSize = 100
	config.ServiceName = "test"

	// config.NotDisplayLine = true
	logger.Init(&config)
}

func TestBrowser(t *testing.T) {
	initLog()
	b := InitBrowser(`/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome`, true,
		nil, "", true)
	defer b.Close()

	reqURL := "http://10.220.170.86:8181/#/login?redirect=%2Fmonitor%2Findex"
	u, err := url.Parse(reqURL)
	logger.Info("-----", u, err)
	req := model.Request{
		Method: http.MethodGet,
	}
	req.URL = &model.URL{
		URL: *u,
	}

	tab := NewTab(b, req, TabConfig{
		TabRunTimeout:           config.TabRunTimeout,
		DomContentLoadedTimeout: config.DomContentLoadedTimeout,
		EventTriggerMode:        config.EventTriggerSync,
		EventTriggerInterval:    config.EventTriggerInterval,
		BeforeExitDelay:         config.BeforeExitDelay,
		EncodeURLWithCharset:    true,
		IgnoreKeywords:          config.DefaultIgnoreKeywords,
		//CustomFormValues:        config.InputTextMap,
		//CustomFormKeywordValues: t.crawlerTask.Config.CustomFormKeywordValues,
	})
	tab.Start()

	for _, req := range tab.ResultList {
		logger.Info(req)
	}
	select {}

}
