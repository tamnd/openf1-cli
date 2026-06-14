package openf1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(baseURL string) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = baseURL
	cfg.Rate = 0
	cfg.Retries = 0
	return NewClientConfig(cfg)
}

func TestGetSessions(t *testing.T) {
	fixture := []Session{
		{
			SessionKey:       9662,
			MeetingKey:       1242,
			Year:             2024,
			MeetingName:      "Abu Dhabi Grand Prix",
			CountryName:      "Abu Dhabi",
			CircuitShortName: "abu_dhabi",
			SessionName:      "Race",
			SessionType:      "Race",
			DateStart:        "2024-11-24T13:00:00+00:00",
			DateEnd:          "2024-11-24T15:00:00+00:00",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions" {
			t.Errorf("unexpected path %q, want /sessions", r.URL.Path)
		}
		if r.URL.Query().Get("year") != "2024" {
			t.Errorf("year param = %q, want 2024", r.URL.Query().Get("year"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetSessions(context.Background(), SessionFilter{Year: 2024})
	if err != nil {
		t.Fatalf("GetSessions: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	s := results[0]
	if s.SessionKey != 9662 {
		t.Errorf("SessionKey = %d, want 9662", s.SessionKey)
	}
	if s.SessionName != "Race" {
		t.Errorf("SessionName = %q, want Race", s.SessionName)
	}
	if s.CountryName != "Abu Dhabi" {
		t.Errorf("CountryName = %q, want Abu Dhabi", s.CountryName)
	}
}

func TestGetMeetings(t *testing.T) {
	fixture := []Meeting{
		{
			MeetingKey:   1242,
			Year:         2024,
			MeetingName:  "Abu Dhabi Grand Prix",
			CountryName:  "Abu Dhabi",
			CountryCode:  "UAE",
			DateStart:    "2024-11-21",
			OfficialName: "Formula 1 Etihad Airways Abu Dhabi Grand Prix 2024",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meetings" {
			t.Errorf("unexpected path %q, want /meetings", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetMeetings(context.Background(), MeetingFilter{Year: 2024})
	if err != nil {
		t.Fatalf("GetMeetings: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].MeetingKey != 1242 {
		t.Errorf("MeetingKey = %d, want 1242", results[0].MeetingKey)
	}
}

func TestGetDrivers(t *testing.T) {
	fixture := []Driver{
		{
			DriverNumber: 55,
			FullName:     "Carlos SAINZ",
			NameAcronym:  "SAI",
			TeamName:     "Ferrari",
			TeamColour:   "E8002D",
			CountryCode:  "ESP",
			HeadshotURL:  "https://example.com/sainz.png",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/drivers" {
			t.Errorf("unexpected path %q, want /drivers", r.URL.Path)
		}
		if r.URL.Query().Get("session_key") != "9158" {
			t.Errorf("session_key = %q, want 9158", r.URL.Query().Get("session_key"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetDrivers(context.Background(), DriverFilter{SessionKey: 9158})
	if err != nil {
		t.Fatalf("GetDrivers: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	d := results[0]
	if d.DriverNumber != 55 {
		t.Errorf("DriverNumber = %d, want 55", d.DriverNumber)
	}
	if d.FullName != "Carlos SAINZ" {
		t.Errorf("FullName = %q, want Carlos SAINZ", d.FullName)
	}
}

func TestGetLaps(t *testing.T) {
	dur := 92.5
	fixture := []Lap{
		{
			LapNumber:    1,
			DriverNumber: 55,
			LapDuration:  &dur,
			I1Speed:      266,
			I2Speed:      249,
			STSpeed:      310,
			IsPitOutLap:  true,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/laps" {
			t.Errorf("unexpected path %q, want /laps", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetLaps(context.Background(), LapFilter{SessionKey: 9158, DriverNumber: 55})
	if err != nil {
		t.Fatalf("GetLaps: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	l := results[0]
	if l.LapNumber != 1 {
		t.Errorf("LapNumber = %d, want 1", l.LapNumber)
	}
	if l.LapDuration == nil || *l.LapDuration != 92.5 {
		t.Errorf("LapDuration = %v, want 92.5", l.LapDuration)
	}
	if !l.IsPitOutLap {
		t.Error("IsPitOutLap = false, want true")
	}
}

func TestGetLapsNullDuration(t *testing.T) {
	// LapDuration is null in the first lap of a race typically
	fixture := []map[string]any{
		{
			"lap_number":     1,
			"driver_number":  55,
			"lap_duration":   nil,
			"i1_speed":       266,
			"i2_speed":       249,
			"st_speed":       310,
			"is_pit_out_lap": true,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetLaps(context.Background(), LapFilter{SessionKey: 9158})
	if err != nil {
		t.Fatalf("GetLaps null duration: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].LapDuration != nil {
		t.Errorf("LapDuration = %v, want nil", results[0].LapDuration)
	}
}

func TestGetStints(t *testing.T) {
	fixture := []Stint{
		{
			StintNumber:    1,
			DriverNumber:   55,
			LapStart:       1,
			LapEnd:         20,
			Compound:       "HARD",
			TyreAgeAtStart: 0,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stints" {
			t.Errorf("unexpected path %q, want /stints", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetStints(context.Background(), StintFilter{SessionKey: 9158, DriverNumber: 55})
	if err != nil {
		t.Fatalf("GetStints: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	s := results[0]
	if s.Compound != "HARD" {
		t.Errorf("Compound = %q, want HARD", s.Compound)
	}
}

func TestGetPit(t *testing.T) {
	dur := 23.5
	fixture := []PitStop{
		{
			DriverNumber: 55,
			LapNumber:    21,
			PitDuration:  &dur,
			Date:         "2023-09-15T09:30:53",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pit" {
			t.Errorf("unexpected path %q, want /pit", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetPit(context.Background(), PitFilter{SessionKey: 9158, DriverNumber: 55})
	if err != nil {
		t.Fatalf("GetPit: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	p := results[0]
	if p.DriverNumber != 55 {
		t.Errorf("DriverNumber = %d, want 55", p.DriverNumber)
	}
	if p.PitDuration == nil || *p.PitDuration != 23.5 {
		t.Errorf("PitDuration = %v, want 23.5", p.PitDuration)
	}
}

func TestGetRaceControl(t *testing.T) {
	sector := 7
	fixture := []RaceControlMsg{
		{
			Date:     "2023-09-14T10:02:42+00:00",
			Category: "Flag",
			Flag:     "YELLOW",
			Scope:    "Sector",
			Sector:   &sector,
			Message:  "YELLOW IN TRACK SECTOR 7",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/race_control" {
			t.Errorf("unexpected path %q, want /race_control", r.URL.Path)
		}
		if r.URL.Query().Get("flag") != "YELLOW" {
			t.Errorf("flag param = %q, want YELLOW", r.URL.Query().Get("flag"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	results, err := c.GetRaceControl(context.Background(), RaceControlFilter{SessionKey: 9158, Flag: "YELLOW"})
	if err != nil {
		t.Fatalf("GetRaceControl: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	m := results[0]
	if m.Flag != "YELLOW" {
		t.Errorf("Flag = %q, want YELLOW", m.Flag)
	}
	if m.Sector == nil || *m.Sector != 7 {
		t.Errorf("Sector = %v, want 7", m.Sector)
	}
}

func TestClientRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClientConfig(cfg)

	start := time.Now()
	results, err := c.GetSessions(context.Background(), SessionFilter{})
	if err != nil {
		t.Fatalf("GetSessions after retries: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d", len(results))
	}
	if hits != 3 {
		t.Errorf("server hit %d times, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestClientUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua == "" {
			t.Error("no User-Agent header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.GetSessions(context.Background(), SessionFilter{})
	if err != nil {
		t.Fatalf("GetSessions: %v", err)
	}
}
