// Package openf1 is the library behind the openf1 command line:
// the HTTP client, request shaping, and the typed data models for the OpenF1
// API (api.openf1.org/v1).
//
// OpenF1 is a free, open-source API providing Formula 1 live timing and
// historical session data. No API key required. Data includes sessions,
// meetings, drivers, lap timing, stints, pit stops, race control messages,
// and telemetry.
package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Host is the API host this client talks to.
const Host = "api.openf1.org"

// BaseURL is the API base URL.
const BaseURL = "https://api.openf1.org/v1"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   BaseURL,
		UserAgent: "openf1-cli/0.1 (tamnd87@gmail.com)",
		Rate:      200 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to the OpenF1 API over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with the default config.
func NewClient() *Client {
	cfg := DefaultConfig()
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// NewClientConfig returns a Client configured with cfg.
func NewClientConfig(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// get fetches rawURL and returns the body bytes with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// buildURL constructs an API URL from an endpoint path and query params.
func (c *Client) buildURL(path string, params url.Values) string {
	u := c.cfg.BaseURL + "/" + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}

// addIntParam adds a param only when v is non-zero.
func addIntParam(params url.Values, key string, v int) {
	if v != 0 {
		params.Set(key, strconv.Itoa(v))
	}
}

// addStrParam adds a param only when v is non-empty.
func addStrParam(params url.Values, key, v string) {
	if v != "" {
		params.Set(key, v)
	}
}

// --- Sessions ---

// SessionFilter holds optional filter params for the sessions endpoint.
type SessionFilter struct {
	Year        int
	SessionName string // "Race", "Qualifying", "Sprint", etc.
	Circuit     string // circuit_short_name
}

// GetSessions fetches sessions matching the filter.
func (c *Client) GetSessions(ctx context.Context, f SessionFilter) ([]Session, error) {
	params := url.Values{}
	addIntParam(params, "year", f.Year)
	addStrParam(params, "session_name", f.SessionName)
	addStrParam(params, "circuit_short_name", f.Circuit)
	rawURL := c.buildURL("sessions", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []Session
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode sessions: %w", err)
	}
	return out, nil
}

// --- Meetings ---

// MeetingFilter holds optional filter params for the meetings endpoint.
type MeetingFilter struct {
	Year int
}

// GetMeetings fetches meetings matching the filter.
func (c *Client) GetMeetings(ctx context.Context, f MeetingFilter) ([]Meeting, error) {
	params := url.Values{}
	addIntParam(params, "year", f.Year)
	rawURL := c.buildURL("meetings", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []Meeting
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode meetings: %w", err)
	}
	return out, nil
}

// --- Drivers ---

// DriverFilter holds optional filter params for the drivers endpoint.
type DriverFilter struct {
	SessionKey   int
	DriverNumber int
}

// GetDrivers fetches drivers matching the filter.
func (c *Client) GetDrivers(ctx context.Context, f DriverFilter) ([]Driver, error) {
	params := url.Values{}
	addIntParam(params, "session_key", f.SessionKey)
	addIntParam(params, "driver_number", f.DriverNumber)
	rawURL := c.buildURL("drivers", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []Driver
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode drivers: %w", err)
	}
	return out, nil
}

// --- Laps ---

// LapFilter holds optional filter params for the laps endpoint.
type LapFilter struct {
	SessionKey   int
	DriverNumber int
	LapNumber    int
}

// GetLaps fetches laps matching the filter.
func (c *Client) GetLaps(ctx context.Context, f LapFilter) ([]Lap, error) {
	params := url.Values{}
	addIntParam(params, "session_key", f.SessionKey)
	addIntParam(params, "driver_number", f.DriverNumber)
	addIntParam(params, "lap_number", f.LapNumber)
	rawURL := c.buildURL("laps", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []Lap
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode laps: %w", err)
	}
	return out, nil
}

// --- Stints ---

// StintFilter holds optional filter params for the stints endpoint.
type StintFilter struct {
	SessionKey   int
	DriverNumber int
}

// GetStints fetches stints matching the filter.
func (c *Client) GetStints(ctx context.Context, f StintFilter) ([]Stint, error) {
	params := url.Values{}
	addIntParam(params, "session_key", f.SessionKey)
	addIntParam(params, "driver_number", f.DriverNumber)
	rawURL := c.buildURL("stints", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []Stint
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode stints: %w", err)
	}
	return out, nil
}

// --- Pit ---

// PitFilter holds optional filter params for the pit endpoint.
type PitFilter struct {
	SessionKey   int
	DriverNumber int
}

// GetPit fetches pit stops matching the filter.
func (c *Client) GetPit(ctx context.Context, f PitFilter) ([]PitStop, error) {
	params := url.Values{}
	addIntParam(params, "session_key", f.SessionKey)
	addIntParam(params, "driver_number", f.DriverNumber)
	rawURL := c.buildURL("pit", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []PitStop
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode pit: %w", err)
	}
	return out, nil
}

// --- RaceControl ---

// RaceControlFilter holds optional filter params for the race_control endpoint.
type RaceControlFilter struct {
	SessionKey int
	Flag       string
}

// GetRaceControl fetches race control messages matching the filter.
func (c *Client) GetRaceControl(ctx context.Context, f RaceControlFilter) ([]RaceControlMsg, error) {
	params := url.Values{}
	addIntParam(params, "session_key", f.SessionKey)
	addStrParam(params, "flag", f.Flag)
	rawURL := c.buildURL("race_control", params)

	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var out []RaceControlMsg
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode race_control: %w", err)
	}
	return out, nil
}
