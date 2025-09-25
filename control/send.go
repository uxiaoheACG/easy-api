package control

import (
	"backend/svc"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var client = &http.Client{}

// 结果结构体
type RequestResult struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
	Error      string `json:"error,omitempty"`
}

func Send(config svc.Config, ReqDates svc.RequestData) {
	var wg sync.WaitGroup
	results := make([]RequestResult, ReqDates.Frequency)
	uniqueResults := make(map[string]RequestResult)

	// 保存原始请求，用于多次复制
	originalReq := config.Request

	sem := make(chan struct{}, 200)
	for i := 0; i < ReqDates.Frequency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			sem <- struct{}{} //获取令牌

			// 每次复制请求
			reqCopy, err := cloneRequest(originalReq)
			if err != nil {
				results[idx] = RequestResult{Error: fmt.Sprintf("复制请求失败: %v", err)}
				return
			}

			results[idx] = doRequestWithReq(reqCopy, ReqDates.TimeOut)
			<-sem //释放令牌
		}(i)
	}
	wg.Wait()

	// 统计
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Error == "" && r.StatusCode >= 200 && r.StatusCode < 300 {
			successCount++
		} else {
			failCount++
		}
		key := fmt.Sprintf("%d-%s-%s", r.StatusCode, r.Body, r.Error)
		if _, exists := uniqueResults[key]; !exists {
			uniqueResults[key] = r
		}
	}

	// 转换 map -> slice
	deduped := make([]RequestResult, 0, len(uniqueResults))
	for _, r := range uniqueResults {
		deduped = append(deduped, r)
	}

	// ===== 日志统一输出 =====
	var logBuilder strings.Builder
	logBuilder.WriteString("\n===== 请求汇总日志 =====\n")
	logBuilder.WriteString(fmt.Sprintf("请求方法: %s\n", config.Request.Method))
	logBuilder.WriteString(fmt.Sprintf("请求URL: %s\n", config.Request.URL.String()))
	logBuilder.WriteString(fmt.Sprintf("请求头: %v\n", config.Request.Header))

	if config.Request.Body != nil && config.Request.GetBody != nil {
		bodyCopy, _ := config.Request.GetBody()
		bodyContent, _ := io.ReadAll(bodyCopy)
		logBuilder.WriteString(fmt.Sprintf("请求体: %s\n", string(bodyContent)))
	}

	logBuilder.WriteString(fmt.Sprintf("总请求数: %d\n", ReqDates.Frequency))
	logBuilder.WriteString(fmt.Sprintf("成功: %d, 失败: %d\n", successCount, failCount))
	logBuilder.WriteString("去重后的响应结果:\n")
	for _, r := range deduped {
		if r.Error != "" {
			logBuilder.WriteString(fmt.Sprintf("  错误: %s\n", r.Error))
		} else {
			logBuilder.WriteString(fmt.Sprintf("  状态码: %d, 响应体: %s\n", r.StatusCode, r.Body))
		}
	}
	logBuilder.WriteString("===== 汇总结束 =====\n")

	config.RequestLog.Println(logBuilder.String())

	// ===== 返回给前端 =====
	config.C.JSON(200, gin.H{
		"total":         ReqDates.Frequency,
		"success":       successCount,
		"fail":          failCount,
		"uniqueResults": deduped,
		"allResults":    results,
	})
}

// doRequestWithReq 发送单个请求
func doRequestWithReq(req *http.Request, timeout int) RequestResult {

	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return RequestResult{Error: fmt.Sprintf("发送请求失败: %v", err)}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return RequestResult{
			StatusCode: resp.StatusCode,
			Error:      fmt.Sprintf("读取响应失败: %v", err),
		}
	}

	return RequestResult{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
	}
}

// cloneRequest 复制请求，每次都会得到新的 Body
func cloneRequest(req *http.Request) (*http.Request, error) {
	var bodyCopy io.ReadCloser
	if req.Body != nil {
		if req.GetBody != nil {
			newBody, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			bodyCopy = newBody
		} else {
			// 无法复制 Body
			bodyCopy = nil
		}
	}

	req2 := req.Clone(req.Context())
	req2.Body = bodyCopy
	return req2, nil
}
