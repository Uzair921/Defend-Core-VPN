package engine

import (
        "encoding/json"
        "fmt"
        "net/http"
        "time"

        "defendcore-vpn/internal/policy"
)

type PolicyClient struct {
        backendURL string
        apiKey     string
        http       *http.Client
}

func NewPolicyClient(backendURL, apiKey string) *PolicyClient {
        return &PolicyClient{
                backendURL: backendURL,
                apiKey:     apiKey,
                http:       &http.Client{Timeout: 5 * time.Second},
        }
}

func (c *PolicyClient) FetchPolicy(userID, deviceID string) (*policy.PolicySet, error) {
        url := fmt.Sprintf("%s/api/v1/vpn/policies/%s", c.backendURL, userID)
        if deviceID != "" {
                url += "?device_id=" + deviceID
        }

        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
                return nil, err
        }
        req.Header.Set("X-API-Key", c.apiKey)

        resp, err := c.http.Do(req)
        if err != nil {
                return nil, fmt.Errorf("policy fetch: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return nil, fmt.Errorf("policy fetch status: %d", resp.StatusCode)
        }

        var set policy.PolicySet
        if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
                return nil, fmt.Errorf("decode policy: %w", err)
        }
        return &set, nil
}
