package scouting

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/guregu/null/v5"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/sbd"
)

type Match struct {
	UUID           uuid.UUID             `db:"match.uuid"`
	LeagueUUID     uuid.UUID             `db:"match.league_uuid"`
	AwayTeamUUID   uuid.UUID             `db:"match.away_team_uuid"`
	HomeTeamUUID   uuid.UUID             `db:"match.home_team_uuid"`
	CreatedBy      string                `db:"match.created_by"`
	HomeScore      null.Value[uint]      `db:"match.home_score"`
	AwayScore      null.Value[uint]      `db:"match.away_score"`
	OrganizationID string                `db:"match.organization_id"`
	StartsAt       time.Time             `db:"match.starts_at"`
	FinishedAt     null.Value[time.Time] `db:"match.finished_at"`
	CreatedAt      time.Time             `db:"match.created_at"`
	ModifiedAt     time.Time             `db:"match.modified_at"`
}

type NewMatch struct {
	LeagueUUID   uuid.UUID `json:"league_uuid"`
	AwayTeamUUID uuid.UUID `json:"away_team_uuid"`
	HomeTeamUUID uuid.UUID `json:"home_team_uuid"`
	StartsAt     time.Time `json:"starts_at"`
}

func (nm *NewMatch) ToMatch(oid, aid string) Match {
	tnow := time.Now()

	return Match{
		UUID:           uuid.Must(uuid.NewV7()),
		LeagueUUID:     nm.LeagueUUID,
		CreatedBy:      aid,
		AwayTeamUUID:   nm.AwayTeamUUID,
		HomeTeamUUID:   nm.HomeTeamUUID,
		OrganizationID: oid,
		StartsAt:       nm.StartsAt,
		CreatedAt:      tnow,
		ModifiedAt:     tnow,
	}
}

func (m *Match) Validate() error {
	if m.StartsAt.Before(time.Now()) {
		return errors.New("starts at cannot be before now")
	}

	return nil
}

type Mode string

const (
	ModeAttack        Mode = "attack"
	ModeDefence       Mode = "defence"
	ModeAttackDefence Mode = "attack_defence"
)

type Submode string

const (
	SubmodeAllRules    Submode = "all_rules"
	SubmodeAnyRules    Submode = "any_rules"
	SubmodeOurRules    Submode = "our_rules"
	SubmodeNotOurRules Submode = "not_our_rules"
	SubmodePlays       Submode = "plays"
)

var errInternal = errors.New("internal error")

func CreateMatch(ctx context.Context, sdb *sqlx.DB, oid, aid string, nm NewMatch) (Match, error) {
	logger := slog.With(
		slog.String("league_uuid", nm.LeagueUUID.String()),
		slog.String("home_team_uuid", nm.HomeTeamUUID.String()),
		slog.String("away_team_uuid", nm.AwayTeamUUID.String()),
		slog.String("organization_id", oid),
		slog.String("account_id", aid),
	)

	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		logger.Error("beginning transaction", slog.Any("error", err))

		return Match{}, errInternal
	}

	defer tx.Rollback()

	leagues, err := SelectLeagues(ctx, tx, LeagueFilter{
		LeagueUUID:     nm.LeagueUUID,
		OrganizationID: oid,
	})
	switch {
	case err == nil && len(leagues) > 0:
		// OK.
	case err == nil && len(leagues) == 0:
		return Match{}, sbd.NewNotFoundError("league")
	default:
		logger.Error("selecting leagues", slog.Any("error", err))

		return Match{}, errInternal
	}

	league := leagues[0]

	teams, err := SelectTeams(ctx, tx, TeamFilter{
		LeagueUUID: league.UUID,
		UUIDs:      []uuid.UUID{nm.HomeTeamUUID, nm.AwayTeamUUID},
	})
	if err != nil {
		logger.Error("selecting league teams", slog.Any("error", err))

		return Match{}, errInternal
	}

	if len(teams) != 2 {
		return Match{}, sbd.NewValidationError("team not found in league")
	}

	m := nm.ToMatch(oid, aid)

	if err := m.Validate(); err != nil {
		return Match{}, sbd.NewValidationError(err.Error())
	}

	if err = insertMatch(ctx, tx, m); err != nil {
		logger.Error("inserting match", slog.Any("error", err))

		return Match{}, errors.Join(err, errInternal)
	}

	if err = tx.Commit(); err != nil {
		logger.Error("commiting", slog.Any("error", err))

		return Match{}, errInternal
	}

	return m, nil
}

type ScoutReport struct {
}

func SubmitScoutReport(ctx context.Context, sdb *sqlx.DB, oid, aid string, matchUUID uuid.UUID, sr ScoutReport) (MatchScout, error) {
	logger := slog.With(
		slog.String("organization_id", oid),
		slog.String("match_uuid", matchUUID.String()),
	)

	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		logger.Error("beginning tx", slog.Any("error", err))

		return MatchScout{}, errInternal
	}

	mss, err := SelectMatchScouts(ctx, tx, MatchScoutFilter{
		MatchUUID:           &matchUUID,
		MatchOrganizationID: &oid,
	})
	if err != nil {
		logger.Error("selecting match scouts", slog.Any("error", err))

		return MatchScout{}, errInternal
	}

	var ms *MatchScout

	for _, m := range mss {
		if m.AccountID == aid {
			ms = &m

			break
		}
	}

	if ms == nil {
		return MatchScout{}, sbd.NewValidationError("match scout not found")
	}

	if ms.FinishedAt.Valid {
		return MatchScout{}, sbd.NewValidationError("match scout already finished")
	}

	tnow := time.Now()

	ms.FinishedAt = null.NewValue(tnow, true)

	if err = updateMatchScout(ctx, tx, *ms); err != nil {
		logger.Error("updating match scout", slog.Any("error", err))

		return MatchScout{}, errInternal
	}

	defer tx.Rollback()

	if err = tx.Commit(); err != nil {
		logger.Error("commiting", slog.Any("error", err))

		return MatchScout{}, errInternal
	}

	return *ms, nil
}

type MatchFinishRequest struct {
	HomeScore uint `json:"home_score"`
	AwayScore uint `json:"away_score"`
}

func FinishMatch(
	ctx context.Context,
	sdb *sqlx.DB,
	oid string,
	matchUUID uuid.UUID,
	fr MatchFinishRequest,
) (Match, error) {
	logger := slog.With(
		slog.String("organization_id", oid),
		slog.String("match_uuid", matchUUID.String()),
	)

	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		logger.Error("beginning tx", slog.Any("error", err))

		return Match{}, errInternal
	}

	defer tx.Rollback()

	mm, err := SelectMatches(ctx, tx, MatchFilter{
		UUID:           matchUUID,
		Active:         true,
		OrganizationID: oid,
	}, true)
	switch {
	case err == nil && len(mm) > 0:
		// OK.
	case err == nil && len(mm) == 0:
		return Match{}, sbd.NewNotFoundError("match")
	default:
		logger.Error("selecting organization matches", slog.Any("error", err))

		return Match{}, errInternal
	}

	m := mm[0]

	mss, err := SelectMatchScouts(ctx, tx, MatchScoutFilter{
		MatchUUID:           &m.UUID,
		MatchOrganizationID: &oid,
	})
	if err != nil {
		logger.Error("selecting match scouts", slog.Any("error", err))

		return Match{}, errInternal
	}

	if err := validateMatchFinish(m, mss); err != nil {
		return Match{}, sbd.NewValidationError(err.Error())
	}

	now := time.Now()

	m.HomeScore = null.NewValue(fr.HomeScore, true)
	m.AwayScore = null.NewValue(fr.AwayScore, true)
	m.FinishedAt = null.NewValue(now, true)
	m.ModifiedAt = now

	if err = updateMatch(ctx, tx, m); err != nil {
		logger.Error("updating match", slog.Any("error", err))

		return Match{}, errInternal
	}

	if err = tx.Commit(); err != nil {
		logger.Error("commiting", slog.Any("error", err))

		return Match{}, errInternal
	}

	return m, nil
}

type MatchFilter struct {
	Active         bool
	UUID           uuid.UUID
	OrganizationID string
}

func validateMatchFinish(m Match, mss []MatchScout) error {
	if m.FinishedAt.Valid {
		return errors.New("match already finished")
	}

	for _, ms := range mss {
		if !ms.FinishedAt.Valid {
			return errors.New("not all scouts have finished")
		}
	}

	return nil
}

func matchScoutable(_ Match, aid string, mss []MatchScout, sr NewMatchScout) error {
	for _, ms := range mss {
		if ms.AccountID == aid {
			return errors.New("account already scouting this match")
		}
	}

	if !modeSubmodeValid(sr.Mode, sr.Submode) {
		return errors.New("mode and submode combination not valid")
	}

	for _, ms := range mss {
		if modesSubmodesConflicts(modeSubmode{ms.Mode, ms.Submode}, modeSubmode{sr.Mode, sr.Submode}) {
			return errors.New("mode and submode conflicts with other scouts")
		}
	}

	return nil
}

type modeSubmode struct {
	mode    Mode
	submode Submode
}

func modesSubmodesConflicts(sm1 modeSubmode, sm2 modeSubmode) bool {
	conflicting := make(map[Mode]map[Submode]struct{})

	pp, ok := conflicting[sm1.mode]
	if !ok {
		pp = make(map[Submode]struct{})
	}

	switch sm1.submode {
	case SubmodeOurRules:
		pp[SubmodeOurRules] = struct{}{}
		pp[SubmodeAnyRules] = struct{}{}
		pp[SubmodeAllRules] = struct{}{}
	case SubmodeNotOurRules:
		pp[SubmodeNotOurRules] = struct{}{}
		pp[SubmodeAnyRules] = struct{}{}
		pp[SubmodeAllRules] = struct{}{}
	case SubmodeAnyRules, SubmodeAllRules:
		pp[SubmodeOurRules] = struct{}{}
		pp[SubmodeAnyRules] = struct{}{}
		pp[SubmodeAllRules] = struct{}{}
		pp[SubmodeNotOurRules] = struct{}{}
	case SubmodePlays:
		pp[SubmodePlays] = struct{}{}
	}

	if sm1.mode == ModeAttackDefence {
		conflicting[ModeAttack] = pp
		conflicting[ModeDefence] = pp
	} else {
		conflicting[sm1.mode] = pp
	}

	var modes []Mode

	if sm2.mode == ModeAttackDefence {
		modes = []Mode{ModeAttack, ModeDefence}
	} else {
		modes = []Mode{sm2.mode}
	}

	for _, m := range modes {
		pairs, ok := conflicting[m]
		if ok {
			if _, ok := pairs[sm2.submode]; ok {
				return true
			}
		}
	}

	return false
}

func modeSubmodeValid(m Mode, sm Submode) bool {
	switch {
	case m == ModeAttack && sm == SubmodeAllRules:
		return true
	case m == ModeAttack && sm == SubmodeOurRules:
		return true
	case m == ModeDefence && sm == SubmodeOurRules:
		return true
	case m == ModeAttack && sm == SubmodeNotOurRules:
		return true
	case m == ModeDefence && sm == SubmodeNotOurRules:
		return true
	case m == ModeDefence && sm == SubmodeAllRules:
		return true
	case m == ModeAttack && sm == SubmodeAnyRules:
		return true
	case m == ModeDefence && sm == SubmodeAnyRules:
		return true
	case m == ModeAttack && sm == SubmodePlays:
		return true
	case m == ModeDefence && sm == SubmodePlays:
		return true
	case m == ModeAttackDefence && sm == SubmodeAnyRules:
		return true
	case m == ModeAttackDefence && sm == SubmodePlays:
		return true
	}

	return false
}

type MatchScoutFilter struct {
	MatchUUID           *uuid.UUID
	MatchOrganizationID *string
}

type MatchScout struct {
	MatchUUID  uuid.UUID             `db:"match_scout.match_uuid"`
	AccountID  string                `db:"match_scout.account_id"`
	Mode       Mode                  `db:"match_scout.mode"`
	Submode    Submode               `db:"match_scout.submode"`
	FinishedAt null.Value[time.Time] `db:"match_scout.finished_at"`
}

type NewMatchScout struct {
	Mode    Mode    `json:"mode"`
	Submode Submode `json:"submode"`
}

func ScoutMatch(ctx context.Context, sdb *sqlx.DB, oid, aid string, matchUUID uuid.UUID, sr NewMatchScout) error {
	logger := slog.With(
		slog.String("account_id", aid),
		slog.String("organization_id", oid),
		slog.String("match_uuid", matchUUID.String()),
	)

	tx, err := sdb.BeginTxx(ctx, nil)
	if err != nil {
		logger.Error("beginning tx", slog.Any("error", err))

		return errInternal
	}

	defer tx.Rollback()

	mm, err := SelectMatches(ctx, tx, MatchFilter{
		UUID:           matchUUID,
		Active:         true,
		OrganizationID: oid,
	}, true)
	switch {
	case err == nil && len(mm) > 0:
		// OK.
	case err == nil && len(mm) == 0:
		return sbd.NewNotFoundError("match")
	default:
		logger.Error("selecting matches", slog.Any("error", err))

		return errInternal
	}

	m := mm[0]

	mss, err := SelectMatchScouts(ctx, tx, MatchScoutFilter{
		MatchUUID: &m.UUID,
	})
	if err != nil {
		logger.Error("selecting match scouts", slog.Any("error", err))

		return errInternal
	}

	if err := matchScoutable(m, aid, mss, sr); err != nil {
		return sbd.NewValidationError(err.Error())
	}

	ms := MatchScout{
		MatchUUID: m.UUID,
		AccountID: aid,
		Mode:      sr.Mode,
		Submode:   sr.Submode,
	}

	if err = insertMatchScout(ctx, tx, ms); err != nil {
		logger.Error("inserting match scout", slog.Any("error", err))

		return errInternal
	}

	if err = tx.Commit(); err != nil {
		logger.Error("commiting", slog.Any("error", err))

		return errInternal
	}

	return nil
}
