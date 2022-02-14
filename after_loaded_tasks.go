package chrome_engine

import (
	"chrome_engine/config"
	"chrome_engine/js"
	"context"

	"github.com/mhqiang/logger"
	"github.com/mhqiang/tools"

	"fmt"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

/**
根据NODE节点执行JS的代码
err := EvaluateAsDevTools(snippet(submitJS, cashX(true), sel, nodes[0]), &res).Do(ctx)

具体环境实现在 chromedp.submit 函数中 参考即可写出
*/

/**
在页面Loaded之后执行
同时等待 afterDOMRun 之后执行
*/
func (tab *Tab) AfterLoadedRun() {
	defer tab.WG.Done()
	logger.Debug("afterLoadedRun start")
	tab.formSubmitWG.Add(2)
	tab.loadedWG.Add(3)
	tab.removeLis.Add(1)

	go tab.formSubmit()
	tab.formSubmitWG.Wait()
	logger.Debug("formSubmit end")

	if tab.config.EventTriggerMode == config.EventTriggerAsync {
		go tab.triggerJavascriptProtocol()
		go tab.triggerInlineEvents()
		go tab.triggerDom2Events()
		tab.loadedWG.Wait()
	} else if tab.config.EventTriggerMode == config.EventTriggerSync {
		tab.triggerInlineEvents()
		time.Sleep(tab.config.EventTriggerInterval)
		tab.triggerDom2Events()
		time.Sleep(tab.config.EventTriggerInterval)
		tab.triggerJavascriptProtocol()
	}

	// 事件触发之后 需要等待一点时间让浏览器成功发出ajax请求 更新DOM
	time.Sleep(tab.config.BeforeExitDelay)

	go tab.RemoveDOMListener()
	tab.removeLis.Wait()
	logger.Debug("afterLoadedRun end")
}

/**
自动化点击提交表单
*/
func (tab *Tab) formSubmit() {

	logger.Debug("formSubmit start")

	// 首先对form表单设置target
	tab.setFormToFrame()

	// 接下来尝试三种方式提交表单
	go tab.clickSubmit()
	go tab.clickAllButton()
}

/**
设置form的target指向一个frame
*/
func (tab *Tab) setFormToFrame() {
	// 首先新建 frame
	nameStr := tools.RandSeq(8)
	tab.Evaluate(fmt.Sprintf(js.NewFrameTemplate, nameStr, nameStr))

	// 接下来将所有的 form 节点target都指向它
	ctx := tab.GetExecutor()
	formNodes, formErr := tab.GetNodeIDs(`form`)
	if formErr != nil || len(formNodes) == 0 {
		logger.Debug("setFormToFrame: get form element err")
		if formErr != nil {
			logger.Debug(formErr)
		}
		return
	}
	tCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	_ = chromedp.SetAttributeValue(formNodes, "target", nameStr, chromedp.ByNodeID).Do(tCtx)
}

/**
点击按钮 type=submit
*/
func (tab *Tab) clickSubmit() {
	defer tab.formSubmitWG.Done()

	// 首先点击按钮 type=submit
	ctx := tab.GetExecutor()

	// 获取所有的form节点 直接执行submit
	formNodes, formErr := tab.GetNodeIDs(`form`)
	if formErr != nil || len(formNodes) == 0 {
		logger.Debug("clickSubmit: get form element err")
		if formErr != nil {
			logger.Debug(formErr)
		}
		return
	}
	tCtx1, cancel1 := context.WithTimeout(ctx, time.Second*2)
	defer cancel1()
	_ = chromedp.Submit(formNodes, chromedp.ByNodeID).Do(tCtx1)

	// 获取所有的input标签
	inputNodes, inputErr := tab.GetNodeIDs(`form input[type=submit]`)
	if inputErr != nil || len(inputNodes) == 0 {
		logger.Debug("clickSubmit: get form input element err")
		if inputErr != nil {
			logger.Debug(inputErr)
		}
		return
	}
	tCtx2, cancel2 := context.WithTimeout(ctx, time.Second*2)
	defer cancel2()
	_ = chromedp.Click(inputNodes, chromedp.ByNodeID).Do(tCtx2)
}

/**
click all button
*/
func (tab *Tab) clickAllButton() {
	defer tab.formSubmitWG.Done()

	// 获取所有的form中的button节点
	ctx := tab.GetExecutor()
	// 获取所有的button标签
	btnNodeIDs, bErr := tab.GetNodeIDs(`form button`)
	if bErr != nil || len(btnNodeIDs) == 0 {
		logger.Debug("clickAllButton: get form button element err")
		if bErr != nil {
			logger.Debug(bErr)
		}
		return
	}
	tCtx, cancel1 := context.WithTimeout(ctx, time.Second*2)
	defer cancel1()
	_ = chromedp.Click(btnNodeIDs, chromedp.ByNodeID).Do(tCtx)

	// 使用JS的click方法进行点击
	var btnNodes []*cdp.Node
	tCtx2, cancel2 := context.WithTimeout(ctx, time.Second*2)
	defer cancel2()
	err := chromedp.Nodes(btnNodeIDs, &btnNodes, chromedp.ByNodeID).Do(tCtx2)
	if err != nil {
		return
	}
	for _, node := range btnNodes {
		_ = tab.EvaluateWithNode(js.FormNodeClickJS, node)
	}
}

/**
触发内联事件
*/
func (tab *Tab) triggerInlineEvents() {
	defer tab.loadedWG.Done()
	logger.Debug("triggerInlineEvents start")
	tab.Evaluate(fmt.Sprintf(js.TriggerInlineEventJS, tab.config.EventTriggerInterval.Seconds()*1000))
	logger.Debug("triggerInlineEvents end")
}

/**
触发DOM2级事件
*/
func (tab *Tab) triggerDom2Events() {
	defer tab.loadedWG.Done()
	logger.Debug("triggerDom2Events start")
	tab.Evaluate(fmt.Sprintf(js.TriggerDom2EventJS, tab.config.EventTriggerInterval.Seconds()*1000))
	logger.Debug("triggerDom2Events end")
}

/**
a标签的href值为伪协议，
*/
func (tab *Tab) triggerJavascriptProtocol() {
	defer tab.loadedWG.Done()
	logger.Debug("clickATagJavascriptProtocol start")
	tab.Evaluate(fmt.Sprintf(js.TriggerJavascriptProtocol, tab.config.EventTriggerInterval.Seconds()*1000,
		tab.config.EventTriggerInterval.Seconds()*1000))
	logger.Debug("clickATagJavascriptProtocol end")
}

/**
移除DOM节点变化监听
*/
func (tab *Tab) RemoveDOMListener() {
	defer tab.removeLis.Done()
	logger.Debug("RemoveDOMListener start")
	// 移除DOM节点变化监听
	tab.Evaluate(js.RemoveDOMListenerJS)
	logger.Debug("RemoveDOMListener end")
}
