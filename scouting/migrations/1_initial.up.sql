CREATE TABLE IF NOT EXISTS organization (
    id TEXT PRIMARY KEY NOT NULL,
    scouting_config JSONB NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS team (
    uuid UUID PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS league (
    uuid UUID PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS league_team (
    team_uuid UUID NOT NULL REFERENCES team(uuid),
    league_uuid UUID NOT NULL REFERENCES league(uuid),

    PRIMARY KEY(league_uuid, team_uuid)
);

CREATE TABLE IF NOT EXISTS organization_league (
    organization_id TEXT NOT NULL REFERENCES organization(id),
    league_uuid UUID NOT NULL REFERENCES league(uuid),

    PRIMARY KEY(organization_id, league_uuid)
);

CREATE TABLE account (
    id TEXT PRIMARY KEY NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    avatar_url TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_account (
    account_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organization(id),

    PRIMARY KEY(organization_id, account_id)
);

CREATE TABLE IF NOT EXISTS match (
    uuid UUID PRIMARY KEY NOT NULL,
    league_uuid UUID NOT NULL REFERENCES league(uuid),
    home_team_uuid UUID NOT NULL REFERENCES team(uuid),
    away_team_uuid UUID NOT NULL REFERENCES team(uuid),
    created_by TEXT NOT NULL,
    home_score INTEGER,
    away_score INTEGER,
    organization_id TEXT NOT NULL REFERENCES organization(id),

    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS match_scout (
    match_uuid UUID NOT NULL REFERENCES match(uuid),
    account_id TEXT NOT NULL REFERENCES account(id),
    mode TEXT NOT NULL,
    submode TEXT NOT NULL,
    finished_at TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (match_uuid, account_id),
    UNIQUE (match_uuid, mode, submode)
);
