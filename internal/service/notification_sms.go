package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-basic/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// smsProviderAliyun 阿里云短信
	smsProviderAliyun = "aliyun"
	// aliyunSMSDefaultEndpoint 阿里云短信默认端点
	aliyunSMSDefaultEndpoint = "dysmsapi.aliyuncs.com"
	// smsDefaultTemplateParamKey 短信模板变量默认名
	smsDefaultTemplateParamKey = "content"
	// smsRequestTimeout 短信请求超时时间
	smsRequestTimeout = 10 * time.Second
)

// smsAliyunConfig 阿里云短信配置，来源为配置文件 conf.yml 的 sms.aliyun 段（可用 GOTP_SMS_ALIYUN_* 环境变量覆盖）
type smsAliyunConfig struct {
	Endpoint         string // 短信服务端点，如 dysmsapi.aliyuncs.com
	AccessKeyID      string // 阿里云 AccessKeyId
	AccessKeySecret  string // 阿里云 AccessKeySecret
	SignName         string // 已报备通过的短信签名
	TemplateCode     string // 已报备通过的短信模板 ID
	TemplateParamKey string // 模板中承载告警内容的变量名
}

// smsConfig 短信通知配置
type smsConfig struct {
	Enabled  bool
	Provider string
	Aliyun   smsAliyunConfig
}

// loadSMSConfig 从配置文件/环境变量读取短信配置
func loadSMSConfig() smsConfig {
	paramKey := strings.TrimSpace(viper.GetString("sms.aliyun.template_param_key"))
	if paramKey == "" {
		paramKey = smsDefaultTemplateParamKey
	}
	return smsConfig{
		Enabled:  viper.GetBool("sms.enabled"),
		Provider: strings.ToLower(strings.TrimSpace(viper.GetString("sms.provider"))),
		Aliyun: smsAliyunConfig{
			Endpoint:         strings.TrimSpace(viper.GetString("sms.aliyun.endpoint")),
			AccessKeyID:      strings.TrimSpace(viper.GetString("sms.aliyun.access_key_id")),
			AccessKeySecret:  strings.TrimSpace(viper.GetString("sms.aliyun.access_key_secret")),
			SignName:         strings.TrimSpace(viper.GetString("sms.aliyun.sign_name")),
			TemplateCode:     strings.TrimSpace(viper.GetString("sms.aliyun.template_code")),
			TemplateParamKey: paramKey,
		},
	}
}

// validate 校验短信配置是否完整可用
func (c smsConfig) validate() error {
	if !c.Enabled {
		return fmt.Errorf("短信通知未启用（sms.enabled=false）")
	}
	if c.Provider == "" {
		return fmt.Errorf("短信服务商未配置（sms.provider）")
	}
	if c.Provider != smsProviderAliyun {
		return fmt.Errorf("不支持的短信服务商：%s", c.Provider)
	}
	if c.Aliyun.AccessKeyID == "" || c.Aliyun.AccessKeySecret == "" {
		return fmt.Errorf("阿里云短信 AccessKey 未配置（sms.aliyun.access_key_id / access_key_secret）")
	}
	if c.Aliyun.SignName == "" {
		return fmt.Errorf("阿里云短信签名未配置（sms.aliyun.sign_name）")
	}
	if c.Aliyun.TemplateCode == "" {
		return fmt.Errorf("阿里云短信模板未配置（sms.aliyun.template_code）")
	}
	return nil
}

// SendSms 按配置发送一条短信，返回服务商消息 ID
func SendSms(phone, content string) (string, error) {
	cfg := loadSMSConfig()
	if err := cfg.validate(); err != nil {
		return "", err
	}
	if strings.TrimSpace(phone) == "" {
		return "", fmt.Errorf("手机号为空")
	}
	switch cfg.Provider {
	case smsProviderAliyun:
		return sendAliyunSMS(cfg.Aliyun, phone, content)
	default:
		return "", fmt.Errorf("不支持的短信服务商：%s", cfg.Provider)
	}
}

// sendAliyunSMS 调用阿里云短信 SendSms 接口（RPC 风格，HMAC-SHA1 签名，仅使用标准库实现）
func sendAliyunSMS(c smsAliyunConfig, phone, content string) (string, error) {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = aliyunSMSDefaultEndpoint
	}
	scheme := "https://"
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		scheme = ""
	}
	baseURL := scheme + strings.TrimRight(endpoint, "/") + "/"

	paramKey := c.TemplateParamKey
	if paramKey == "" {
		paramKey = smsDefaultTemplateParamKey
	}
	templateParam, err := json.Marshal(map[string]string{paramKey: content})
	if err != nil {
		return "", fmt.Errorf("构造短信模板参数失败: %w", err)
	}

	params := map[string]string{
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"RegionId":         "cn-hangzhou",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   uuid.New(),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"AccessKeyId":      c.AccessKeyID,
		"PhoneNumbers":     phone,
		"SignName":         c.SignName,
		"TemplateCode":     c.TemplateCode,
		"TemplateParam":    string(templateParam),
	}

	// 计算签名：StringToSign = GET&%2F&<规范化查询串>
	canonical := aliyunCanonicalQuery(params)
	stringToSign := "GET&" + aliyunPercentEncode("/") + "&" + aliyunPercentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(c.AccessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	params["Signature"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))

	requestURL := baseURL + "?" + aliyunCanonicalQuery(params)
	client := &http.Client{Timeout: smsRequestTimeout}
	resp, err := client.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("请求阿里云短信服务失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取阿里云短信响应失败: %w", err)
	}

	var result struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestId string `json:"RequestId"`
		BizId     string `json:"BizId"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析阿里云短信响应失败: %w, body=%s", err, string(body))
	}
	if result.Code != "OK" {
		return "", fmt.Errorf("阿里云短信发送失败: Code=%s Message=%s RequestId=%s", result.Code, result.Message, result.RequestId)
	}
	logrus.Info("阿里云短信发送受理成功, phone=", phone, " BizId=", result.BizId)
	return result.BizId, nil
}

// aliyunCanonicalQuery 按阿里云 RPC 签名要求生成规范化查询串（参数名排序 + RFC3986 编码）
func aliyunCanonicalQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, aliyunPercentEncode(k)+"="+aliyunPercentEncode(params[k]))
	}
	return strings.Join(parts, "&")
}

// aliyunPercentEncode 阿里云签名要求的 RFC3986 百分号编码
func aliyunPercentEncode(s string) string {
	e := url.QueryEscape(s)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "*", "%2A")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}
