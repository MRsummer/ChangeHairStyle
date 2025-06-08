/*
Copyright (year) Beijing Volcano Engine Technology Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package volcengine

import (
        "bytes"
        "crypto/hmac"
        "crypto/sha256"
        "encoding/hex"
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "net/http/httputil"
        "net/url"
        "strings"
        "time"
)

// Client 火山引擎API客户端
type Client struct {
        AccessKeyID     string
        SecretAccessKey string
        Addr            string
        Path            string
        Service         string
        Region          string
        Action          string
        Version         string
}

// NewClient 创建新的火山引擎API客户端
func NewClient(accessKeyID, secretAccessKey string) *Client {
        return &Client{
                AccessKeyID:     accessKeyID,
                SecretAccessKey: secretAccessKey,
                Addr:            "https://visual.volcengineapi.com",
                Path:            "/",
                Service:         "cv",
                Region:          "cn-north-1",
                Action:          "CVProcess",
                Version:         "2022-08-31",
        }
}

func hmacSHA256(key []byte, content string) []byte {
        mac := hmac.New(sha256.New, key)
        mac.Write([]byte(content))
        return mac.Sum(nil)
}

func getSignedKey(secretKey, date, region, service string) []byte {
        kDate := hmacSHA256([]byte(secretKey), date)
        kRegion := hmacSHA256(kDate, region)
        kService := hmacSHA256(kRegion, service)
        kSigning := hmacSHA256(kService, "request")

        return kSigning
}

func hashSHA256(data []byte) []byte {
        hash := sha256.New()
        if _, err := hash.Write(data); err != nil {
                log.Printf("input hash err:%s", err.Error())
        }

        return hash.Sum(nil)
}

// DoRequest 发送请求到火山引擎API
func (c *Client) DoRequest(method string, queries url.Values, body []byte) ([]byte, int, error) {
        // 1. 构建请求
        queries.Set("Action", c.Action)
        queries.Set("Version", c.Version)
        requestAddr := fmt.Sprintf("%s%s?%s", c.Addr, c.Path, queries.Encode())
        log.Printf("请求地址: %s\n", requestAddr)

        request, err := http.NewRequest(method, requestAddr, bytes.NewBuffer(body))
        if err != nil {
                return nil, 0, fmt.Errorf("bad request: %w", err)
        }

        // 2. 构建签名材料
        now := time.Now()
        date := now.UTC().Format("20060102T150405Z")
        authDate := date[:8]
        request.Header.Set("X-Date", date)

        payload := hex.EncodeToString(hashSHA256(body))
        request.Header.Set("X-Content-Sha256", payload)
        request.Header.Set("Content-Type", "application/json")

        queryString := strings.Replace(queries.Encode(), "+", "%20", -1)
        signedHeaders := []string{"host", "x-date", "x-content-sha256", "content-type"}
        var headerList []string
        for _, header := range signedHeaders {
                if header == "host" {
                        headerList = append(headerList, header+":"+request.Host)
                } else {
                        v := request.Header.Get(header)
                        headerList = append(headerList, header+":"+strings.TrimSpace(v))
                }
        }
        headerString := strings.Join(headerList, "\n")

        canonicalString := strings.Join([]string{
                method,
                c.Path,
                queryString,
                headerString + "\n",
                strings.Join(signedHeaders, ";"),
                payload,
        }, "\n")
        log.Printf("规范字符串:\n%s\n", canonicalString)

        hashedCanonicalString := hex.EncodeToString(hashSHA256([]byte(canonicalString)))
        log.Printf("哈希后的规范字符串: %s\n", hashedCanonicalString)

        credentialScope := authDate + "/" + c.Region + "/" + c.Service + "/request"
        signString := strings.Join([]string{
                "HMAC-SHA256",
                date,
                credentialScope,
                hashedCanonicalString,
        }, "\n")
        log.Printf("签名字符串:\n%s\n", signString)

        // 3. 构建认证请求头
        signedKey := getSignedKey(c.SecretAccessKey, authDate, c.Region, c.Service)
        signature := hex.EncodeToString(hmacSHA256(signedKey, signString))
        log.Printf("签名: %s\n", signature)

        authorization := "HMAC-SHA256" +
                " Credential=" + c.AccessKeyID + "/" + credentialScope +
                ", SignedHeaders=" + strings.Join(signedHeaders, ";") +
                ", Signature=" + signature

        request.Header.Set("Authorization", authorization)
        log.Printf("认证头: %s\n", authorization)

        // 4. 打印请求，发起请求
        requestRaw, err := httputil.DumpRequest(request, true)
        if err != nil {
                return nil, 0, fmt.Errorf("dump request err: %w", err)
        }

        log.Printf("完整请求:\n%s\n", string(requestRaw))

        response, err := http.DefaultClient.Do(request)
        if err != nil {
                return nil, 0, fmt.Errorf("do request err: %w", err)
        }
        defer response.Body.Close()

        // 5. 读取响应内容
        responseBody, err := io.ReadAll(response.Body)
        if err != nil {
                return nil, 0, fmt.Errorf("read response body err: %w", err)
        }

        // 6. 打印响应
        log.Printf("响应状态码: %d\n", response.StatusCode)
        log.Printf("响应内容:\n%s\n", string(responseBody))

        return responseBody, response.StatusCode, nil
}

// ProcessCV 处理计算机视觉请求
func (c *Client) ProcessCV(reqKey string, params map[string]interface{}) ([]byte, int, error) {
        // 构建请求体
        reqBody := map[string]interface{}{
                "req_key": reqKey,
        }
        // 合并其他参数
        for k, v := range params {
                reqBody[k] = v
        }

        reqBodyStr, err := json.Marshal(reqBody)
        if err != nil {
                return nil, 0, fmt.Errorf("marshal request body err: %w", err)
        }

        return c.DoRequest("POST", url.Values{}, reqBodyStr)
}

// GenerateHairStyle 生成新的发型图片
func (c *Client) GenerateHairStyle(imageURL string, prompt string) (string, error) {
        // 准备请求参数
        params := map[string]interface{}{
                "image_urls": []string{imageURL},
                "prompt":     prompt,
                "return_url": true,
        }

        // 发送请求
        response, statusCode, err := c.ProcessCV("byteedit_v2.0", params)
        if err != nil {
                return "", fmt.Errorf("请求失败: %v", err)
        }

        if statusCode != 200 {
                return "", fmt.Errorf("API返回错误状态码: %d", statusCode)
        }

        // 解析响应
        var result map[string]interface{}
        if err := json.Unmarshal(response, &result); err != nil {
                return "", fmt.Errorf("解析响应失败: %v", err)
        }

        // 提取图片URL
        if data, ok := result["data"].(map[string]interface{}); ok {
                if imageUrls, ok := data["image_urls"].([]interface{}); ok && len(imageUrls) > 0 {
                        if url, ok := imageUrls[0].(string); ok {
                                return url, nil
                        }
                }
        }

        return "", fmt.Errorf("未找到生成的图片URL")
} 