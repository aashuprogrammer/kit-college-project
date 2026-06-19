package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CashfreeClient struct {
	appID       string
	secretKey   string
	environment string
	baseURL     string
	httpClient  *http.Client
}

type CashfreeCustomerDetails struct {
	CustomerID    string `json:"customer_id"`
	CustomerName  string `json:"customer_name,omitempty"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
}

type CashfreeOrderMeta struct {
	ReturnURL string `json:"return_url"`
}

type CashfreeCreateOrderRequest struct {
	OrderID         string                  `json:"order_id"`
	OrderAmount     float64                 `json:"order_amount"`
	OrderCurrency   string                  `json:"order_currency"`
	CustomerDetails CashfreeCustomerDetails `json:"customer_details"`
	OrderMeta       CashfreeOrderMeta       `json:"order_meta"`
}

type CashfreeOrderResponse struct {
	OrderID          string                  `json:"order_id"`
	CFOrderID        interface{}             `json:"cf_order_id"`
	OrderAmount      float64                 `json:"order_amount"`
	OrderCurrency    string                  `json:"order_currency"`
	OrderStatus      string                  `json:"order_status"` // e.g. ACTIVE, PAID, EXPIRED
	PaymentSessionID string                  `json:"payment_session_id"`
	CustomerDetails  CashfreeCustomerDetails `json:"customer_details"`
}

type CashfreeErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Type    string `json:"type"`
}

func NewCashfreeClient(appID, secretKey, environment string) *CashfreeClient {
	var baseURL string
	if strings.ToLower(environment) == "production" {
		baseURL = "https://api.cashfree.com/pg"
	} else {
		baseURL = "https://sandbox.cashfree.com/pg"
	}

	return &CashfreeClient{
		appID:       appID,
		secretKey:   secretKey,
		environment: environment,
		baseURL:     baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateOrder calls the Cashfree POST /orders endpoint to initialize a payment session
func (c *CashfreeClient) CreateOrder(ctx context.Context, orderID string, amount float64, customerID, customerName, customerEmail, customerPhone string, returnURL string) (*CashfreeOrderResponse, error) {
	reqBody := CashfreeCreateOrderRequest{
		OrderID:       orderID,
		OrderAmount:   amount,
		OrderCurrency: "INR",
		CustomerDetails: CashfreeCustomerDetails{
			CustomerID:    customerID,
			CustomerName:  customerName,
			CustomerEmail: customerEmail,
			CustomerPhone: customerPhone,
		},
		OrderMeta: CashfreeOrderMeta{
			ReturnURL: returnURL,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/orders", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cashfree api call: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var cfErr CashfreeErrorResponse
		if err := json.Unmarshal(respBytes, &cfErr); err == nil && cfErr.Message != "" {
			return nil, fmt.Errorf("cashfree api error: %s (code: %s, type: %s)", cfErr.Message, cfErr.Code, cfErr.Type)
		}
		return nil, fmt.Errorf("cashfree api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var orderResp CashfreeOrderResponse
	if err := json.Unmarshal(respBytes, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &orderResp, nil
}

// GetOrder calls the Cashfree GET /orders/{order_id} endpoint to verify payment status
func (c *CashfreeClient) GetOrder(ctx context.Context, orderID string) (*CashfreeOrderResponse, error) {
	reqURL := fmt.Sprintf("%s/orders/%s", c.baseURL, orderID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cashfree status call: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var cfErr CashfreeErrorResponse
		if err := json.Unmarshal(respBytes, &cfErr); err == nil && cfErr.Message != "" {
			return nil, fmt.Errorf("cashfree status api error: %s (code: %s)", cfErr.Message, cfErr.Code)
		}
		return nil, fmt.Errorf("cashfree status api returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var orderResp CashfreeOrderResponse
	if err := json.Unmarshal(respBytes, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &orderResp, nil
}

// VerifyWebhookSignature verifies the payload signature sent by Cashfree
func (c *CashfreeClient) VerifyWebhookSignature(signature, timestamp, rawBody string) bool {
	data := timestamp + rawBody
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(data))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (c *CashfreeClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-client-id", c.appID)
	req.Header.Set("x-client-secret", c.secretKey)
	req.Header.Set("x-api-version", "2023-08-01")
}
