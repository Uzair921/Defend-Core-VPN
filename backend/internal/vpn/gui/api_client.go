package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL string
	http    *http.Client
	token   string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type LoginResponse struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c *APIClient) Login(email, password, mfaCode string) (*LoginResponse, error) {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	if mfaCode != "" {
		body["mfa_code"] = mfaCode
	}

	data, _ := json.Marshal(body)
	resp, err := c.http.Post(c.baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s", string(b))
	}

	var out LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode login: %w", err)
	}
	c.token = out.AccessToken
	return &out, nil
}

type Device struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
	Platform  string `json:"platform"`
	Status    string `json:"status"`
}

func (c *APIClient) ListDevices() ([]Device, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/api/v1/devices", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Devices []Device `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Devices, nil
}

type DeviceConfig struct {
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	ServerHost   string `json:"server_host"`
	ServerPort   int    `json:"server_port"`
	AssignedIP   string `json:"assigned_ip"`
	PublicKey    string `json:"public_key"`
	PrivateKey   string `json:"private_key"`
	ServerPubKey string `json:"server_public_key,omitempty"`
	DNS          string `json:"dns,omitempty"`
}

func (c *APIClient) GetDeviceConfig(deviceID string) (*DeviceConfig, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/api/v1/devices/"+deviceID, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get config failed: %s", string(b))
	}

	var out DeviceConfig
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MyService represents a VPN service assigned to the user's organization.
type MyService struct {
ID           string   `json:"id"`
Name         string   `json:"name"`
Slug         string   `json:"slug"`
ServiceType  string   `json:"service_type"`
Description  string   `json:"description,omitempty"`
Subnet       string   `json:"subnet"`
ServerIP     string   `json:"server_ip"`
DNSServers   []string `json:"dns_servers"`
MaxClients   int      `json:"max_clients"`
Status       string   `json:"status"`
CustomRoutes []string `json:"custom_routes,omitempty"`
TypeName     string   `json:"type_name,omitempty"`
TypeIcon     string   `json:"type_icon,omitempty"`
}

// MyOrganization represents the user's organization.
type MyOrganization struct {
ID                string `json:"id"`
Name              string `json:"name"`
Slug              string `json:"slug"`
Email             string `json:"email"`
Status            string `json:"status"`
Plan              string `json:"plan"`
MaxUsers          int    `json:"max_users"`
MaxServices       int    `json:"max_services"`
MonthlyPriceCents *int   `json:"monthly_price_cents,omitempty"`
}

// ListMyServices fetches VPN services available to the authenticated user.
func (c *APIClient) ListMyServices() ([]MyService, error) {
req, _ := http.NewRequest("GET", c.baseURL+"/api/v1/vpn/services", nil)
req.Header.Set("Authorization", "Bearer "+c.token)

resp, err := c.http.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
return nil, fmt.Errorf("list services failed (%d): %s", resp.StatusCode, string(b))
}

var out struct {
Services []MyService `json:"services"`
}
if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
return nil, err
}
return out.Services, nil
}

// GetMyOrganization fetches the current user's organization.
func (c *APIClient) GetMyOrganization() (*MyOrganization, error) {
req, _ := http.NewRequest("GET", c.baseURL+"/api/v1/org/me", nil)
req.Header.Set("Authorization", "Bearer "+c.token)

resp, err := c.http.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

if resp.StatusCode == http.StatusForbidden {
return nil, nil // User is not a member of any org
}
if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
return nil, fmt.Errorf("get org failed (%d): %s", resp.StatusCode, string(b))
}

var org MyOrganization
if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
return nil, err
}
return &org, nil
}

// ServiceConfig is the config returned for a specific service.
type ServiceConfig struct {
ServiceID       string   `json:"service_id"`
ServiceName     string   `json:"service_name"`
ServiceType     string   `json:"service_type"`
ServerHost      string   `json:"server_host"`
ServerPort      int      `json:"server_port"`
ServerPublicKey string   `json:"server_public_key"`
AssignedIP      string   `json:"assigned_ip"`
Routes          []string `json:"routes"`
DNSServers      []string `json:"dns_servers"`
MTU             int      `json:"mtu"`
}

// GetServiceConfig fetches the client config for a specific VPN service.
func (c *APIClient) GetServiceConfig(serviceID string) (*ServiceConfig, error) {
req, _ := http.NewRequest("GET", c.baseURL+"/api/v1/vpn/services/"+serviceID+"/config", nil)
req.Header.Set("Authorization", "Bearer "+c.token)

resp, err := c.http.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
return nil, fmt.Errorf("get service config failed (%d): %s", resp.StatusCode, string(b))
}

var cfg ServiceConfig
if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
return nil, err
}
return &cfg, nil
}
