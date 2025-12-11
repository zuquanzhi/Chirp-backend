package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// Sender defines the interface for sending SMS
type Sender interface {
	Send(ctx context.Context, phone, code, purpose string) error
}

// ConsoleSender is a mock sender that logs to console (for dev/test)
type ConsoleSender struct{}

func (s *ConsoleSender) Send(ctx context.Context, phone, code, purpose string) error {
	log.Printf("[SMS] To: %s, Code: %s, Purpose: %s", phone, code, purpose)
	return nil
}

// AliyunSender is a placeholder for Aliyun SMS implementation
type AliyunSender struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	RegionID        string
}

func NewAliyunSender(ak, sk, signName, templateCode string) *AliyunSender {
	return &AliyunSender{
		AccessKeyID:     ak,
		AccessKeySecret: sk,
		SignName:        signName,
		TemplateCode:    templateCode,
		RegionID:        "cn-hangzhou",
	}
}

func (s *AliyunSender) Send(ctx context.Context, phone, code, purpose string) error {
	log.Printf("[SMS] Aliyun sending phone=%s purpose=%s code=%s", phone, purpose, code)

	client, err := dysmsapi.NewClientWithAccessKey(s.RegionID, s.AccessKeyID, s.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("init aliyun client: %w", err)
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = s.SignName
	request.TemplateCode = s.TemplateCode

	// TemplateParam must be a JSON string, e.g., {"code":"123456"}
	// Adjust the key "code" based on your actual Aliyun template variable name
	params, _ := json.Marshal(map[string]string{"code": code})
	request.TemplateParam = string(params)

	response, err := client.SendSms(request)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}

	if response.Code != "OK" {
		return fmt.Errorf("aliyun sms error: %s - %s", response.Code, response.Message)
	}

	log.Printf("[SMS] Aliyun sent phone=%s purpose=%s requestId=%s bizId=%s", phone, purpose, response.RequestId, response.BizId)

	return nil
}
