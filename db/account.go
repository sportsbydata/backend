package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sportsbydata/backend/scouting"
)

func accountCols() []string {
	return []string{
		`account.id AS "account.id"`,
		`account.first_name AS "account.first_name"`,
		`account.last_name AS "account.last_name"`,
		`account.avatar_url AS "account.avatar_url"`,
		`account.created_at AS "account.created_at"`,
		`account.modified_at AS "account.modified_at"`,
	}
}

func (d *DB) SelectAccounts(ctx context.Context, qr sqlx.QueryerContext, f scouting.AccountFilter) ([]scouting.Account, error) {
	sb := squirrel.Select(accountCols()...).From("account AS account")

	if f.OrganizationID != nil {
		sb = sb.InnerJoin("organization_account ON organization_account.account_id=account.id").
			Where(squirrel.Eq{"organization_account.organization_id": &f.OrganizationID})
	}

	sql, args := sb.MustSql()

	var aa []scouting.Account

	if err := sqlx.SelectContext(ctx, qr, &aa, sql, args...); err != nil {
		return nil, err
	}

	return aa, nil
}

func (d *DB) UpsertAccount(ctx context.Context, ec sqlx.ExecerContext, u scouting.Account) error {
	sb := squirrel.Insert("account").SetMap(map[string]any{
		"id":          u.ID,
		"first_name":  u.FirstName,
		"last_name":   u.LastName,
		"avatar_url":  u.AvatarURL,
		"created_at":  u.CreatedAt,
		"modified_at": u.ModifiedAt,
	}).Suffix("ON CONFLICT (id) DO UPDATE set first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, avatar_url = EXCLUDED.avatar_url, modified_at = EXCLUDED.modified_at")

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return err
}

func (d *DB) UpsertOrganizationAccount(ctx context.Context, ec sqlx.ExecerContext, oid, aid string) error {
	sb := squirrel.Insert("organization_account").SetMap(map[string]any{
		"account_id":      aid,
		"organization_id": oid,
	}).Suffix("ON CONFLICT DO NOTHING")

	sq, args := sb.MustSql()

	_, err := ec.ExecContext(ctx, sq, args...)
	return err
}
