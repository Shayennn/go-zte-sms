package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ZTEConnector struct {
	client  *http.Client
	baseURL string
}

func NewZTEConnector(baseURL string) (*ZTEConnector, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar: jar,
		// No timeout here; timeouts are handled per request with contexts
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &ZTEConnector{
		client:  client,
		baseURL: baseURL,
	}, nil
}

// helper function to create HTTP requests with necessary headers
func (zte *ZTEConnector) newRequest(ctx context.Context, method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", zte.baseURL)
	return req, nil
}

func (zte *ZTEConnector) getNumericConfig(ctx context.Context, configName string) (int, error) {
	url := zte.baseURL + "js/config/config.js"
	req, err := zte.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request for config.js: %w", err)
	}

	resp, err := zte.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch config.js: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("config.js returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read config.js response: %w", err)
	}
	body := string(bodyBytes)

	re := regexp.MustCompile(regexp.QuoteMeta(configName) + `\s*:\s*(\d),`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return 0, fmt.Errorf("failed to find config %s", configName)
	}

	return strconv.Atoi(matches[1])
}

func (zte *ZTEConnector) getLoginLD(ctx context.Context) (string, error) {
	params := url.Values{
		"isTest": []string{"false"},
		"cmd":    []string{"LD"},
		"_":      []string{fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))},
	}

	fullURL := zte.baseURL + "goform/goform_get_cmd_process?" + params.Encode()
	req, err := zte.newRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for LD: %w", err)
	}

	resp, err := zte.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get LD: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LD request returned status %d", resp.StatusCode)
	}

	var response struct {
		LD string `json:"LD"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode LD response: %w", err)
	}

	if response.LD == "" {
		return "", fmt.Errorf("LD is null")
	}

	return response.LD, nil
}

func getLoginErrorMessage(code string) string {
	switch code {
	case "0":
		return "Login OK"
	case "1":
		return "Login Fail"
	case "2":
		return "Duplicate User"
	case "3":
		return "Bad Password"
	default:
		return "Unknown error"
	}
}

func sha256Hex(str string) string {
	hash := sha256.Sum256([]byte(str))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (zte *ZTEConnector) Login(password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	supportSHA256, err := zte.getNumericConfig(ctx, "WEB_ATTR_IF_SUPPORT_SHA256")
	if err != nil {
		return fmt.Errorf("failed to get WEB_ATTR_IF_SUPPORT_SHA256: %w", err)
	}

	var hashPassword string
	switch supportSHA256 {
	case 0:
		hashPassword = base64.StdEncoding.EncodeToString([]byte(password))
	case 1:
		hashPassword = sha256Hex(base64.StdEncoding.EncodeToString([]byte(password)))
	case 2:
		LD, err := zte.getLoginLD(ctx)
		if err != nil {
			return fmt.Errorf("failed to get LD: %w", err)
		}
		hashPassword = sha256Hex(sha256Hex(password) + LD)
	default:
		return fmt.Errorf("unsupported WEB_ATTR_IF_SUPPORT_SHA256: %d", supportSHA256)
	}

	params := url.Values{
		"isTest":   []string{"false"},
		"goformId": []string{"LOGIN"},
		"password": []string{hashPassword},
	}

	fullURL := zte.baseURL + "goform/goform_set_cmd_process"
	req, err := zte.newRequest(ctx, http.MethodPost, fullURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := zte.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login request returned status %d", resp.StatusCode)
	}

	var response struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if response.Result != "0" {
		return fmt.Errorf("login failed: %s", getLoginErrorMessage(response.Result))
	}

	return nil
}

func (zte *ZTEConnector) GetSMS(page, perPage, memStore, tag int) ([]ZTESMS, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	params := url.Values{
		"isTest":        []string{"false"},
		"cmd":           []string{"sms_data_total"},
		"page":          []string{strconv.Itoa(page)},
		"data_per_page": []string{strconv.Itoa(perPage)},
		"mem_store":     []string{strconv.Itoa(memStore)},
		"tags":          []string{strconv.Itoa(tag)},
		"order_by":      []string{"order by id desc"},
		"_":             []string{fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))},
	}

	fullURL := zte.baseURL + "goform/goform_get_cmd_process?" + params.Encode()
	req, err := zte.newRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetSMS request: %w", err)
	}

	resp, err := zte.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform GetSMS request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetSMS request returned status %d", resp.StatusCode)
	}

	var response struct {
		Messages []ZTEMessage `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode GetSMS response: %w", err)
	}

	var smsList []ZTESMS
	for _, msg := range response.Messages {
		sms, err := NewZTESMS(msg)
		if err != nil {
			log.Printf("Error parsing SMS ID %s: %v", msg.ID, err)
			continue
		}
		smsList = append(smsList, sms)
	}

	return smsList, nil
}
