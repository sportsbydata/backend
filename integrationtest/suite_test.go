package integrationtest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/sportsbydata/backend/db"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	pool        *dockertest.Pool
	postgresRes *dockertest.Resource
	sdb         *sqlx.DB
}

func (s *Suite) SetupSuite() {
	var err error

	s.pool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	var dsn string

	s.postgresRes, dsn, err = setupPostgres(&s.Suite, s.pool)
	s.Require().NoError(err)

	err = s.pool.Retry(func() error {
		sdb, err := db.Connect(context.Background(), dsn)
		if err != nil {
			return err
		}

		if err = db.Migrate(sdb.DB); err != nil {
			return err
		}

		s.sdb = sdb

		return nil
	})
	s.Require().NoError(err)
}

func TestDBSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(Suite))
}

func (s *Suite) TearDownTest() {
	tables := []string{
		"organization_league",
		"league_team",
		"match_scout",
		"match",
		"league",
		"team",
		"organization_account",
		"account",
		"organization",
	}

	for _, table := range tables {
		_, err := s.sdb.ExecContext(context.Background(), fmt.Sprintf("DELETE FROM %s", table))
		s.Require().NoError(err)
	}
}

func (s *Suite) TearDownSuite() {
	s.Require().NoError(s.sdb.Close())
	s.Require().NoError(s.postgresRes.Close())
}

func setupPostgres(s *suite.Suite, pool *dockertest.Pool) (*dockertest.Resource, string, error) {
	res, err := pool.Run("postgres", "latest", []string{
		"POSTGRES_USER=root",
		"POSTGRES_PASSWORD=password",
		"POSTGRES_DB=scouting",
	})
	s.Require().NoError(err)

	dsn := "postgres://root:password@" + hostPort(res, "5432/tcp") + "/scouting"

	return res, dsn, nil
}

func hostPort(r *dockertest.Resource, portID string) string {
	containersURL := os.Getenv("CONTAINERS_HOST")
	if containersURL != "" {
		return fmt.Sprintf("%s:%s", containersURL, r.GetPort(portID))
	}

	return r.GetHostPort(portID)
}

func (s *Suite) selectCount(table string, pred any) int {
	s.T().Helper()

	sql, args := squirrel.Select("COUNT(*)").From(table).Where(pred).MustSql()

	row := s.sdb.QueryRowxContext(context.Background(), sql, args...)
	s.Require().NoError(row.Err())

	var cnt int
	s.Require().NoError(row.Scan(&cnt))

	return cnt
}
