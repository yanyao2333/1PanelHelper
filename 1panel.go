package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PanelApiStruct struct {
	BaseUrl string
	Session string
}

// GetCronjobRecords 获取定时任务记录
func (p *PanelApiStruct) GetCronjobRecords(request CronjobRecordRequest) (*CronjobRecordResponse, error) {
	if p.Session == "" {
		return nil, errors.New("你还没有登录")
	}
	endTime := time.Now().Format(time.RFC3339)
	startTime := time.Now().Add(-time.Hour * 24 * time.Duration(request.Days)).Format(time.RFC3339)
	type reqStruct struct {
		Page      int    `json:"page"`
		PageSize  int    `json:"pageSize"`
		CronjobID int    `json:"cronjobID"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
		Status    string `json:"status"`
	}
	reqData := reqStruct{
		Page:      request.Page,
		PageSize:  request.PageSize,
		CronjobID: request.CronjobID,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "",
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("序列化 JSON 数据错误: %w", err)
	}

	req, err := http.NewRequest("POST", p.BaseUrl+"/api/v1/cronjobs/search/records", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", p.Session)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("意外的状态码: %d, 响应体: %s", resp.StatusCode, string(body))
	}

	var recordResp CronjobRecordResponse
	err = json.Unmarshal(body, &recordResp)
	if err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %w", err)
	}

	if recordResp.Code != 200 {
		return nil, errors.New(fmt.Sprintf("api返回内容状态码非200,请自行检查: %v", recordResp))
	}

	return &recordResp, nil
}

// GetCronjobList 获取定时任务列表
func (p *PanelApiStruct) GetCronjobList(request SearchRequest) (*SearchResponse, error) {
	if p.Session == "" {
		return nil, errors.New("你还没有登录")
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化 JSON 数据错误: %w", err)
	}

	req, err := http.NewRequest("POST", p.BaseUrl+"/api/v1/cronjobs/search", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", p.Session)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("意外的状态码: %d, 响应体: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	err = json.Unmarshal(body, &searchResp)
	if err != nil {
		return nil, fmt.Errorf("反序列化响应失败: %w", err)
	}

	if searchResp.Code != 200 {
		return nil, errors.New(fmt.Sprintf("api返回内容状态码非200,请自行检查: %v", searchResp))
	}

	return &searchResp, nil
}

// LoginWithout2FA 普通登录
func (p *PanelApiStruct) LoginWithout2FA(username, password, entranceCode string) error {
	requestData := struct {
		Name          string `json:"name"`
		Password      string `json:"password"`
		IgnoreCaptcha bool   `json:"ignoreCaptcha"`
		Captcha       string `json:"captcha"`
		CaptchaID     string `json:"captchaID"`
		AuthMethod    string `json:"authMethod"`
		Language      string `json:"language"`
	}{
		Name:          username,
		Password:      password,
		IgnoreCaptcha: true,
		Captcha:       "",
		CaptchaID:     "1145141919810",
		AuthMethod:    "session",
		Language:      "zh",
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化 JSON 数据错误: %w", err)
	}
	req, err := http.NewRequest("POST", p.BaseUrl+"/api/v1/auth/login", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	entranceCodeB64 := base64.StdEncoding.EncodeToString([]byte(entranceCode))
	req.Header.Set("entrancecode", entranceCodeB64)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("意外的状态码: %d, 响应体: %s", resp.StatusCode, string(body))
	}
	if resp.Header.Get("Set-Cookie") == "" {
		return fmt.Errorf("找不到Set-Cookie字段,请根据返回内容自行排查原因: %s", string(body))
	}
	p.Session = resp.Header.Get("Set-Cookie")
	return nil

}
