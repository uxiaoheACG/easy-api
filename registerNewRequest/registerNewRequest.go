package registerNewRequest

import (
	"backend/svc"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func RegisterNewRequest(data *svc.RequestData, config svc.Config) (*http.Request, error) {
	var bodyBuffer *bytes.Buffer
	var contentType string

	// 优先级：Files > Form > Body
	if len(data.Files) > 0 {
		// multipart/form-data
		bodyBuffer = &bytes.Buffer{}
		writer := multipart.NewWriter(bodyBuffer)

		// 添加表单字段
		for k, v := range data.Form {
			_ = writer.WriteField(k, v)
		}

		// 添加文件
		for field, path := range data.Files {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("打开文件失败: %v", err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile(field, filepath.Base(path))
			if err != nil {
				return nil, fmt.Errorf("创建文件表单字段失败: %v", err)
			}
			if _, err = io.Copy(part, file); err != nil {
				return nil, fmt.Errorf("写入文件内容失败: %v", err)
			}
		}

		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("关闭 multipart writer 失败: %v", err)
		}
		contentType = writer.FormDataContentType()

	} else if len(data.Form) > 0 {
		// application/x-www-form-urlencoded
		form := url.Values{}
		for k, v := range data.Form {
			form.Set(k, v)
		}
		bodyBuffer = bytes.NewBufferString(form.Encode())
		contentType = "application/x-www-form-urlencoded"

	} else if data.Body != nil {
		// 根据 Content-Type 决定编码方式
		ct := strings.ToLower(data.Header["Content-Type"])
		switch {
		case strings.Contains(ct, "xml"):
			// XML
			var xmlBytes []byte
			switch b := data.Body.(type) {
			case string:
				xmlBytes = []byte(b) // 直接传 XML 字符串
			default:
				var err error
				xmlBytes, err = xml.Marshal(b)
				if err != nil {
					return nil, fmt.Errorf("构建 XML 请求体失败: %v", err)
				}
			}
			bodyBuffer = bytes.NewBuffer(xmlBytes)
			contentType = "application/xml"

		case strings.Contains(ct, "plain"):
			// 纯文本
			switch b := data.Body.(type) {
			case string:
				bodyBuffer = bytes.NewBufferString(b)
			default:
				return nil, fmt.Errorf("纯文本请求体必须传 string 类型")
			}
			contentType = "text/plain"

		default:
			// 默认 JSON
			jsonBytes, err := json.Marshal(data.Body)
			if err != nil {
				return nil, fmt.Errorf("构建 JSON 请求体失败: %v", err)
			}
			bodyBuffer = bytes.NewBuffer(jsonBytes)
			contentType = "application/json"
		}

	} else {
		// 无请求体
		bodyBuffer = &bytes.Buffer{}
	}

	// 构建请求
	req, err := http.NewRequest(strings.ToUpper(data.Method), data.Url, bodyBuffer)
	if err != nil {
		config.RequestLog.Println("构建请求失败:", err.Error())
		return nil, err
	}

	// 设置请求头
	for k, v := range data.Header {
		req.Header.Set(k, v)
	}

	// 自动补充 Content-Type（除非用户已手动设置）
	if req.Header.Get("Content-Type") == "" && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 添加 GET 参数
	if len(data.Params) > 0 {
		q := req.URL.Query()
		for k, v := range data.Params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}
