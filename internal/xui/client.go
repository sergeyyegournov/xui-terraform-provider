package xui

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Client talks to a 3x-ui panel using the same cookie session as the web UI.
type Client struct {
	baseURL *url.URL
	http    *http.Client
	user    string
	pass    string
}

// NewClient builds an HTTP client; baseURL must include the panel path prefix (e.g. https://host:port/<uuid>/).
func NewClient(rawBaseURL, username, password string, insecureSkipVerify bool) (*Client, error) {
	u, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base_url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base_url must include scheme and host")
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	return &Client{
		baseURL: u,
		http: &http.Client{
			Jar:       jar,
			Transport: tr,
		},
		user: username,
		pass: password,
	}, nil
}

func (c *Client) join(elem ...string) (string, error) {
	return c.baseURL.JoinPath(elem...).String(), nil
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login establishes a session cookie. Call before other API methods.
func (c *Client) Login() error {
	endpoint, err := c.join("login")
	if err != nil {
		return err
	}
	body, err := json.Marshal(loginBody{Username: c.user, Password: c.pass})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	var msg APIResponse
	if err := json.Unmarshal(b, &msg); err != nil {
		return fmt.Errorf("login: decode response: %w; body=%s", err, truncate(b, 512))
	}
	if !msg.Success {
		return fmt.Errorf("login failed: %s", msg.Msg)
	}
	return nil
}

func truncate(b []byte, n int) string {
	s := string(b)
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func (c *Client) postJSON(path []string, payload any) (*APIResponse, error) {
	endpoint, err := c.join(path...)
	if err != nil {
		return nil, err
	}
	var body []byte
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	return c.requestJSON(http.MethodPost, endpoint, body)
}

func (c *Client) get(path []string) (*APIResponse, error) {
	endpoint, err := c.join(path...)
	if err != nil {
		return nil, err
	}
	return c.requestJSON(http.MethodGet, endpoint, nil)
}

func (c *Client) requestJSON(method, endpoint string, body []byte) (*APIResponse, error) {
	if err := c.Login(); err != nil {
		return nil, err
	}
	doOnce := func() ([]byte, int, error) {
		var rdr io.Reader
		if body != nil {
			rdr = bytes.NewReader(body)
		}
		req, err := http.NewRequest(method, endpoint, rdr)
		if err != nil {
			return nil, 0, err
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, 0, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		return b, resp.StatusCode, err
	}
	b, status, err := doOnce()
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		if err := c.Login(); err != nil {
			return nil, err
		}
		b, _, err = doOnce()
		if err != nil {
			return nil, err
		}
	}
	var msg APIResponse
	if err := json.Unmarshal(b, &msg); err != nil {
		return nil, fmt.Errorf("%s %s: %w; body=%s", method, endpoint, err, truncate(b, 512))
	}
	if !msg.Success {
		return nil, fmt.Errorf("%s %s: %s", method, endpoint, msg.Msg)
	}
	return &msg, nil
}

// ListInbounds returns raw obj JSON (array of inbounds).
func (c *Client) ListInbounds() (json.RawMessage, error) {
	msg, err := c.get([]string{"panel", "api", "inbounds", "list"})
	if err != nil {
		return nil, err
	}
	return msg.Obj, nil
}

// GetInbound returns one inbound as JSON object.
func (c *Client) GetInbound(id int) (json.RawMessage, error) {
	msg, err := c.get([]string{"panel", "api", "inbounds", "get", fmt.Sprintf("%d", id)})
	if err != nil {
		return nil, err
	}
	return msg.Obj, nil
}

// AddInbound creates an inbound; returns created inbound JSON in Obj.
func (c *Client) AddInbound(payload map[string]any) (json.RawMessage, error) {
	msg, err := c.postJSON([]string{"panel", "api", "inbounds", "add"}, payload)
	if err != nil {
		return nil, err
	}
	return msg.Obj, nil
}

// UpdateInbound updates inbound by id (id in URL and body).
func (c *Client) UpdateInbound(id int, payload map[string]any) (json.RawMessage, error) {
	msg, err := c.postJSON([]string{"panel", "api", "inbounds", "update", fmt.Sprintf("%d", id)}, payload)
	if err != nil {
		return nil, err
	}
	return msg.Obj, nil
}

// DeleteInbound removes an inbound.
func (c *Client) DeleteInbound(id int) error {
	_, err := c.postJSON([]string{"panel", "api", "inbounds", "del", fmt.Sprintf("%d", id)}, map[string]any{})
	return err
}

// AddInboundClient appends clients in settings to an existing inbound (see 3x-ui AddInboundClient).
func (c *Client) AddInboundClient(inboundID int, settingsWithClientsJSON string) error {
	payload := map[string]any{
		"id":       inboundID,
		"settings": settingsWithClientsJSON,
	}
	_, err := c.postJSON([]string{"panel", "api", "inbounds", "addClient"}, payload)
	return err
}

// UpdateInboundClient updates a single client; clientID is the VLESS UUID string.
func (c *Client) UpdateInboundClient(clientID string, inboundPayload map[string]any) error {
	_, err := c.postJSON([]string{"panel", "api", "inbounds", "updateClient", clientID}, inboundPayload)
	return err
}

// DeleteInboundClient removes a client UUID from an inbound.
func (c *Client) DeleteInboundClient(inboundID int, clientID string) error {
	_, err := c.postJSON([]string{"panel", "api", "inbounds", fmt.Sprintf("%d", inboundID), "delClient", clientID}, map[string]any{})
	return err
}
