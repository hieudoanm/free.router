package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const BaseURL = "https://openrouter.ai/api/v1"

type Model struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ContextLength int     `json:"context_length"`
	Pricing       Pricing `json:"pricing"`
}

type Pricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Request    string `json:"request"`
}

type modelsResponse struct {
	Data []Model `json:"data"`
}

func FetchFreeModels() ([]Model, error) {
	resp, err := http.Get(BaseURL + "/models")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter returned HTTP %d", resp.StatusCode)
	}

	var payload modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	var free []Model
	for _, m := range payload.Data {
		if isFree(m.Pricing) {
			free = append(free, m)
		}
	}

	sort.Slice(free, func(i, j int) bool {
		return free[i].ID < free[j].ID
	})

	return free, nil
}

func isFree(p Pricing) bool {
	return (p.Prompt == "0" || p.Prompt == "") &&
		(p.Completion == "0" || p.Completion == "")
}

func ResolveModel(query string, models []Model) *Model {
	q := strings.ToLower(query)

	for i, m := range models {
		if strings.ToLower(m.ID) == q {
			return &models[i]
		}
	}

	for i, m := range models {
		if strings.ToLower(m.ID) == q+":free" {
			return &models[i]
		}
	}

	var idMatches []Model
	for _, m := range models {
		if strings.Contains(strings.ToLower(m.ID), q) {
			idMatches = append(idMatches, m)
		}
	}
	if len(idMatches) == 1 {
		return &idMatches[0]
	}
	if len(idMatches) > 1 {
		sort.Slice(idMatches, func(i, j int) bool {
			return len(idMatches[i].ID) < len(idMatches[j].ID)
		})
		return &idMatches[0]
	}

	var nameMatches []Model
	for _, m := range models {
		if strings.Contains(strings.ToLower(m.Name), q) {
			nameMatches = append(nameMatches, m)
		}
	}
	if len(nameMatches) > 0 {
		sort.Slice(nameMatches, func(i, j int) bool {
			return len(nameMatches[i].ID) < len(nameMatches[j].ID)
		})
		return &nameMatches[0]
	}

	return nil
}

// ── Probing ───────────────────────────────────────────────────────────────────

type ProbeStatus int

const (
	StatusOK ProbeStatus = iota
	StatusRateLimited
	StatusRestricted
	StatusError
)

type ProbeResult struct {
	Model   Model
	Status  ProbeStatus
	Message string
	Latency int64
}

func ProbeModel(m Model, apiKey string) ProbeResult {
	payload := map[string]any{
		"model": m.ID,
		"messages": []map[string]any{
			{"role": "user", "content": "hi"},
		},
		"max_tokens": 1,
		"provider": map[string]any{
			"allow_fallbacks": false,
			"data_collection": "allow",
		},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ProbeResult{Model: m, Status: StatusError, Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/freerouter/freerouter")
	req.Header.Set("X-Title", "freerouter")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return ProbeResult{Model: m, Status: StatusError, Message: err.Error()}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return ProbeResult{Model: m, Status: StatusOK, Latency: elapsed}
	case http.StatusTooManyRequests:
		return ProbeResult{Model: m, Status: StatusRateLimited, Message: extractMessage(raw)}
	case http.StatusNotFound:
		return ProbeResult{Model: m, Status: StatusRestricted, Message: extractMessage(raw)}
	default:
		msg := extractMessage(raw)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return ProbeResult{Model: m, Status: StatusError, Message: msg}
	}
}

func extractMessage(raw []byte) string {
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Error.Message != "" {
		msg := envelope.Error.Message
		if len(msg) > 80 {
			msg = msg[:79] + "…"
		}
		return msg
	}
	return ""
}
