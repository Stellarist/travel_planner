package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"example.com/travel_planner/backend/config"
)

// CallModel sends a prompt to the configured model and returns the raw content string
func CallModel(ctx context.Context, prompt string) (string, error) {
	cfg := config.Global.Model
	if cfg.ApiKey == "" || cfg.BaseURL == "" || cfg.Model == "" {
		return "", errors.New("model not configured")
	}

	payload := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	reqURL := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"

	// 增加超时时间到 90 秒，因为生成行程需要较长时间
	reqCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, "POST", reqURL, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+cfg.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// HTTP Client 超时设置略大于 context 超时
	client := &http.Client{Timeout: 95 * time.Second}

	startTime := time.Now()

	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("请求失败 (耗时: %.2f秒): %v\n", time.Since(startTime).Seconds(), err)
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("model API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("model error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", errors.New("no choices returned")
	}

	content := extractJSON(result.Choices[0].Message.Content)
	return content, nil
}

// extractJSON strips code fences and returns the JSON-like portion
func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}
