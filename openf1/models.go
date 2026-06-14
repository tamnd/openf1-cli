package openf1

// Session represents one F1 session (Race, Qualifying, Sprint, etc.)
// as returned by the /sessions endpoint.
type Session struct {
	SessionKey       int    `json:"session_key" kit:"id" table:"session_key"`
	MeetingKey       int    `json:"meeting_key" table:"meeting_key"`
	Year             int    `json:"year" table:"year"`
	MeetingName      string `json:"meeting_name" table:"meeting_name"`
	CountryName      string `json:"country_name" table:"country_name"`
	CircuitShortName string `json:"circuit_short_name" table:"circuit_short_name"`
	SessionName      string `json:"session_name" table:"session_name"`
	SessionType      string `json:"session_type" table:"session_type"`
	DateStart        string `json:"date_start" table:"date_start"`
	DateEnd          string `json:"date_end" table:"date_end"`
}

// Meeting represents one F1 race weekend as returned by the /meetings endpoint.
type Meeting struct {
	MeetingKey   int    `json:"meeting_key" kit:"id" table:"meeting_key"`
	Year         int    `json:"year" table:"year"`
	MeetingName  string `json:"meeting_name" table:"meeting_name"`
	CountryName  string `json:"country_name" table:"country_name"`
	CountryCode  string `json:"country_code" table:"country_code"`
	DateStart    string `json:"date_start" table:"date_start"`
	OfficialName string `json:"meeting_official_name" table:"official_name"`
}

// Driver represents a driver entry in a session as returned by /drivers.
type Driver struct {
	DriverNumber int    `json:"driver_number" kit:"id" table:"driver_number"`
	FullName     string `json:"full_name" table:"full_name"`
	NameAcronym  string `json:"name_acronym" table:"acronym"`
	TeamName     string `json:"team_name" table:"team"`
	TeamColour   string `json:"team_colour" table:"team_colour"`
	CountryCode  string `json:"country_code" table:"country"`
	HeadshotURL  string `json:"headshot_url" table:"headshot_url,url"`
}

// Lap represents one lap's timing data as returned by /laps.
// LapDuration may be null when no clean lap time was recorded.
type Lap struct {
	LapNumber    int      `json:"lap_number" kit:"id" table:"lap_number"`
	DriverNumber int      `json:"driver_number" table:"driver_number"`
	LapDuration  *float64 `json:"lap_duration,omitempty" table:"lap_duration"`
	I1Speed      int      `json:"i1_speed" table:"i1_speed"`
	I2Speed      int      `json:"i2_speed" table:"i2_speed"`
	STSpeed      int      `json:"st_speed" table:"st_speed"`
	IsPitOutLap  bool     `json:"is_pit_out_lap" table:"pit_out"`
}

// Stint represents one tyre stint as returned by /stints.
type Stint struct {
	StintNumber    int    `json:"stint_number" kit:"id" table:"stint_number"`
	DriverNumber   int    `json:"driver_number" table:"driver_number"`
	LapStart       int    `json:"lap_start" table:"lap_start"`
	LapEnd         int    `json:"lap_end" table:"lap_end"`
	Compound       string `json:"compound" table:"compound"`
	TyreAgeAtStart int    `json:"tyre_age_at_start" table:"tyre_age"`
}

// PitStop represents one pit stop event as returned by /pit.
// PitDuration may be null for in-laps or laps where the stop time was not recorded.
type PitStop struct {
	DriverNumber int      `json:"driver_number" kit:"id" table:"driver_number"`
	LapNumber    int      `json:"lap_number" table:"lap_number"`
	PitDuration  *float64 `json:"pit_duration,omitempty" table:"pit_duration"`
	Date         string   `json:"date" table:"date"`
}

// RaceControlMsg represents one race control message as returned by /race_control.
// DriverNumber, LapNumber, and Sector are nullable in the API.
type RaceControlMsg struct {
	Date         string `json:"date" table:"date"`
	DriverNumber *int   `json:"driver_number,omitempty" table:"driver_number"`
	LapNumber    *int   `json:"lap_number,omitempty" table:"lap_number"`
	Category     string `json:"category" table:"category"`
	Flag         string `json:"flag,omitempty" table:"flag"`
	Scope        string `json:"scope,omitempty" table:"scope"`
	Sector       *int   `json:"sector,omitempty" table:"sector"`
	Message      string `json:"message" table:"message"`
}
