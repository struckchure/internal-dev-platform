package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type APIClient struct {
	baseURL      string
	httpClient   *http.Client
	AccessToken  string
	RefreshToken string
	tokenPath    string
}

func NewAPIClient(baseURL string) *APIClient {
	client := &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		tokenPath: defaultTokenPath(),
	}
	client.loadTokens()
	return client
}

func (c *APIClient) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimRight(baseURL, "/")
}

func (c *APIClient) Request(method string, path string, query map[string]string, body any, auth bool) (string, int, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", 0, err
	}

	q := u.Query()
	for k, v := range query {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	var reader io.Reader
	if body != nil {
		b, mErr := json.Marshal(body)
		if mErr != nil {
			return "", 0, mErr
		}
		reader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, u.String(), reader)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth && c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", res.StatusCode, err
	}
	if len(respBody) == 0 {
		return fmt.Sprintf(`{"status":"%s"}`, res.Status), res.StatusCode, nil
	}

	var pretty bytes.Buffer
	if json.Indent(&pretty, respBody, "", "  ") == nil {
		return pretty.String(), res.StatusCode, nil
	}

	return string(respBody), res.StatusCode, nil
}

func (c *APIClient) parseMaybeJSON(raw string) any {
	var result any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]any{"raw": raw}
	}
	return result
}

func (c *APIClient) applyTokensFromBody(resp string) bool {
	var env loginEnvelope
	if err := json.Unmarshal([]byte(resp), &env); err == nil && env.Tokens.AccessToken != "" {
		c.AccessToken = env.Tokens.AccessToken
		c.RefreshToken = env.Tokens.RefreshToken
		_ = c.saveTokens()
		return true
	}

	var nested map[string]json.RawMessage
	if err := json.Unmarshal([]byte(resp), &nested); err == nil {
		if raw, ok := nested["tokens"]; ok {
			var tokens Tokens
			if json.Unmarshal(raw, &tokens) == nil && tokens.AccessToken != "" {
				c.AccessToken = tokens.AccessToken
				c.RefreshToken = tokens.RefreshToken
				_ = c.saveTokens()
				return true
			}
		}
	}
	return false
}

func (c *APIClient) IsAuthenticated() bool {
	return c.AccessToken != ""
}

func (c *APIClient) TokenPreview() string {
	if c.AccessToken == "" {
		return ""
	}
	t := c.AccessToken
	if len(t) <= 12 {
		return t
	}
	return t[:6] + "…" + t[len(t)-4:]
}

func (c *APIClient) Register(firstName string, lastName string, email string, password string) (string, int, error) {
	resp, status, err := c.Request(http.MethodPost, "/api/v1/auth/register/", nil, map[string]any{
		"firstName": firstName,
		"lastName":  lastName,
		"email":     email,
		"password":  password,
	}, false)
	if err != nil {
		return resp, status, err
	}
	c.applyTokensFromBody(resp)
	return resp, status, nil
}

func (c *APIClient) Login(email string, password string) (string, int, error) {
	resp, status, err := c.Request(http.MethodPost, "/api/v1/auth/login/", nil, map[string]any{
		"email":    email,
		"password": password,
	}, false)
	if err != nil {
		return resp, status, err
	}
	c.applyTokensFromBody(resp)
	return resp, status, nil
}

func (c *APIClient) Refresh() (string, int, error) {
	resp, status, err := c.Request(http.MethodPost, "/api/v1/auth/refresh-access-token/", nil, map[string]any{
		"refreshToken": c.RefreshToken,
	}, false)
	if err != nil {
		return resp, status, err
	}
	c.applyTokensFromBody(resp)
	return resp, status, nil
}

func (c *APIClient) Logout() {
	c.AccessToken = ""
	c.RefreshToken = ""
	_ = c.saveTokens()
}

func defaultTokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".idp-tui-auth.json"
	}
	return filepath.Join(home, ".idp-tui-auth.json")
}

func (c *APIClient) saveTokens() error {
	data, err := json.MarshalIndent(Tokens{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.tokenPath, data, 0o600)
}

func (c *APIClient) loadTokens() {
	raw, err := os.ReadFile(c.tokenPath)
	if err != nil {
		return
	}
	var tokens Tokens
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return
	}
	c.AccessToken = tokens.AccessToken
	c.RefreshToken = tokens.RefreshToken
}

func (c *APIClient) GetProfile() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/user/profile/", nil, nil, true)
}

func (c *APIClient) UpdateProfile(payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPatch, "/api/v1/user/profile/", nil, payload, true)
}

func (c *APIClient) ListMachines() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/machine/", map[string]string{
		"take": "100",
		"skip": "0",
	}, nil, true)
}

func (c *APIClient) CreateMachine(payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPost, "/api/v1/machine/", nil, payload, true)
}

func (c *APIClient) GetMachine(machineID string) (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/machine/"+machineID, nil, nil, true)
}

func (c *APIClient) UpdateMachine(machineID string, payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPatch, "/api/v1/machine/"+machineID, nil, payload, true)
}

func (c *APIClient) DeleteMachine(machineID string) (string, int, error) {
	return c.Request(http.MethodDelete, "/api/v1/machine/"+machineID, nil, nil, true)
}

func (c *APIClient) ListNetworks() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/network/", nil, nil, true)
}

func (c *APIClient) CreateNetwork(payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPost, "/api/v1/network/", nil, payload, true)
}

func (c *APIClient) DeleteNetwork(networkID string) (string, int, error) {
	return c.Request(http.MethodDelete, "/api/v1/network/"+networkID, nil, nil, true)
}

func (c *APIClient) ListRepoConnections() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/repo-connection/", nil, nil, true)
}

func (c *APIClient) CreateRepoConnection(payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPost, "/api/v1/repo-connection/", nil, payload, true)
}

func (c *APIClient) GetRepoConnection(connectionID string) (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/repo-connection/"+connectionID, nil, nil, true)
}

func (c *APIClient) UpdateRepoConnection(connectionID string, payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPatch, "/api/v1/repo-connection/"+connectionID, nil, payload, true)
}

func (c *APIClient) DeleteRepoConnection(connectionID string) (string, int, error) {
	return c.Request(http.MethodDelete, "/api/v1/repo-connection/"+connectionID, nil, nil, true)
}

func (c *APIClient) ListDeployments(machineID string) (string, int, error) {
	query := map[string]string{}
	if machineID != "" {
		query["machineId"] = machineID
	}
	return c.Request(http.MethodGet, "/api/v1/deployments/", query, nil, true)
}

func (c *APIClient) DeployRepo(payload map[string]any) (string, int, error) {
	return c.Request(http.MethodPost, "/api/v1/deployments/deploy", nil, payload, true)
}

func (c *APIClient) GetDeployment(deploymentID string) (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/deployments/"+deploymentID, nil, nil, true)
}

func (c *APIClient) ListDeploymentLogs(deploymentID string) (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/deployments/"+deploymentID+"/logs", nil, nil, true)
}

func (c *APIClient) ListRepos(page string, size string) (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/gh/repos", map[string]string{
		"pageNumber": page,
		"pageSize":   size,
	}, nil, true)
}

func (c *APIClient) AuthorizeGithub() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/gh/authorize", nil, nil, true)
}

func (c *APIClient) UpdateAppAccess() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/gh/update-app-access", nil, nil, true)
}

func (c *APIClient) ListAccountConnections() (string, int, error) {
	return c.Request(http.MethodGet, "/api/v1/gh/account-connections", nil, nil, true)
}
