// domain.go exposes OpenF1 as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/openf1-cli/openf1"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// openf1:// URIs by routing to the operations Register installs. The same
// Domain also builds the standalone openf1 binary (see cli.NewApp), so the
// binary and a host share one source of truth.
package openf1

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the openf1 driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "openf1",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "openf1",
			Short:  "Formula 1 timing and session data from OpenF1",
			Long: `openf1 reads Formula 1 live timing and historical session data
from the OpenF1 public API (api.openf1.org/v1). No API key required.

Commands cover sessions, race weekends, drivers, lap timing, tyre stints,
pit stops, and race control messages.`,
			Site: "https://" + Host,
			Repo: "https://github.com/tamnd/openf1-cli",
		},
	}
}

// Register installs the client factory and every OpenF1 operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name: "sessions", Group: "read",
		Summary: "List F1 sessions (Race, Qualifying, Sprint, etc.)",
	}, listSessions)

	kit.Handle(app, kit.OpMeta{
		Name: "meetings", Group: "read",
		Summary: "List F1 race weekends",
	}, listMeetings)

	kit.Handle(app, kit.OpMeta{
		Name: "drivers", Group: "read",
		Summary: "List drivers in a session",
	}, listDrivers)

	kit.Handle(app, kit.OpMeta{
		Name: "laps", Group: "read",
		Summary: "List lap timing data for a session",
	}, listLaps)

	kit.Handle(app, kit.OpMeta{
		Name: "stints", Group: "read",
		Summary: "List tyre stint data for a session",
	}, listStints)

	kit.Handle(app, kit.OpMeta{
		Name: "pit", Group: "read",
		Summary: "List pit stop data for a session",
	}, listPit)

	kit.Handle(app, kit.OpMeta{
		Name: "racecontrol", Group: "read",
		Summary: "List race control messages for a session",
	}, listRaceControl)
}

// newClient builds the client from the host-resolved config, so a host and the
// standalone binary pace and identify themselves the same way.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	dcfg := DefaultConfig()
	if cfg.UserAgent != "" {
		dcfg.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		dcfg.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		dcfg.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		dcfg.Timeout = cfg.Timeout
	}
	return NewClientConfig(dcfg), nil
}

// --- input structs ---

type sessionsInput struct {
	Year        int     `kit:"flag" help:"filter by year (e.g. 2024)"`
	SessionName string  `kit:"flag,name=session-name" help:"filter by session name (Race, Qualifying, Sprint)"`
	Circuit     string  `kit:"flag" help:"filter by circuit short name"`
	Limit       int     `kit:"flag,inherit" help:"max results"`
	Client      *Client `kit:"inject"`
}

type meetingsInput struct {
	Year   int     `kit:"flag" help:"filter by year (e.g. 2024)"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type driversInput struct {
	SessionKey int     `kit:"flag,name=session-key" help:"session key (required)"`
	Driver     int     `kit:"flag" help:"filter by driver number"`
	Limit      int     `kit:"flag,inherit" help:"max results"`
	Client     *Client `kit:"inject"`
}

type lapsInput struct {
	SessionKey int     `kit:"flag,name=session-key" help:"session key (required)"`
	Driver     int     `kit:"flag" help:"filter by driver number"`
	Lap        int     `kit:"flag" help:"filter by lap number"`
	Limit      int     `kit:"flag,inherit" help:"max results"`
	Client     *Client `kit:"inject"`
}

type stintsInput struct {
	SessionKey int     `kit:"flag,name=session-key" help:"session key (required)"`
	Driver     int     `kit:"flag" help:"filter by driver number"`
	Limit      int     `kit:"flag,inherit" help:"max results"`
	Client     *Client `kit:"inject"`
}

type pitInput struct {
	SessionKey int     `kit:"flag,name=session-key" help:"session key (required)"`
	Driver     int     `kit:"flag" help:"filter by driver number"`
	Limit      int     `kit:"flag,inherit" help:"max results"`
	Client     *Client `kit:"inject"`
}

type raceControlInput struct {
	SessionKey int     `kit:"flag,name=session-key" help:"session key (required)"`
	Flag       string  `kit:"flag" help:"filter by flag (YELLOW, RED, SAFETY CAR, etc.)"`
	Limit      int     `kit:"flag,inherit" help:"max results"`
	Client     *Client `kit:"inject"`
}

// --- handlers ---

func listSessions(ctx context.Context, in sessionsInput, emit func(Session) error) error {
	results, err := in.Client.GetSessions(ctx, SessionFilter{
		Year:        in.Year,
		SessionName: in.SessionName,
		Circuit:     in.Circuit,
	})
	if err != nil {
		return err
	}
	for i, s := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func listMeetings(ctx context.Context, in meetingsInput, emit func(Meeting) error) error {
	results, err := in.Client.GetMeetings(ctx, MeetingFilter{Year: in.Year})
	if err != nil {
		return err
	}
	for i, m := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(m); err != nil {
			return err
		}
	}
	return nil
}

func listDrivers(ctx context.Context, in driversInput, emit func(Driver) error) error {
	if in.SessionKey == 0 {
		return errs.Usage("--session-key is required for drivers")
	}
	results, err := in.Client.GetDrivers(ctx, DriverFilter{
		SessionKey:   in.SessionKey,
		DriverNumber: in.Driver,
	})
	if err != nil {
		return err
	}
	for i, d := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(d); err != nil {
			return err
		}
	}
	return nil
}

func listLaps(ctx context.Context, in lapsInput, emit func(Lap) error) error {
	if in.SessionKey == 0 {
		return errs.Usage("--session-key is required for laps")
	}
	results, err := in.Client.GetLaps(ctx, LapFilter{
		SessionKey:   in.SessionKey,
		DriverNumber: in.Driver,
		LapNumber:    in.Lap,
	})
	if err != nil {
		return err
	}
	for i, l := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(l); err != nil {
			return err
		}
	}
	return nil
}

func listStints(ctx context.Context, in stintsInput, emit func(Stint) error) error {
	if in.SessionKey == 0 {
		return errs.Usage("--session-key is required for stints")
	}
	results, err := in.Client.GetStints(ctx, StintFilter{
		SessionKey:   in.SessionKey,
		DriverNumber: in.Driver,
	})
	if err != nil {
		return err
	}
	for i, s := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func listPit(ctx context.Context, in pitInput, emit func(PitStop) error) error {
	if in.SessionKey == 0 {
		return errs.Usage("--session-key is required for pit")
	}
	results, err := in.Client.GetPit(ctx, PitFilter{
		SessionKey:   in.SessionKey,
		DriverNumber: in.Driver,
	})
	if err != nil {
		return err
	}
	for i, p := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(p); err != nil {
			return err
		}
	}
	return nil
}

func listRaceControl(ctx context.Context, in raceControlInput, emit func(RaceControlMsg) error) error {
	if in.SessionKey == 0 {
		return errs.Usage("--session-key is required for racecontrol")
	}
	results, err := in.Client.GetRaceControl(ctx, RaceControlFilter{
		SessionKey: in.SessionKey,
		Flag:       in.Flag,
	})
	if err != nil {
		return err
	}
	for i, m := range results {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(m); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: the URI-native string functions ---

// Classify is not deeply meaningful for an API-only domain with no browsable
// page hierarchy, so we return an error to signal this domain is query-driven.
func (Domain) Classify(input string) (uriType, id string, err error) {
	return "", "", errs.Usage("openf1 is a query-driven domain; use commands like `openf1 sessions` or `openf1 drivers`")
}

// Locate returns an error since OpenF1 sessions are not web-browsable pages.
func (Domain) Locate(uriType, id string) (string, error) {
	return "", errs.Usage("openf1 has no web resource type %q", uriType)
}
