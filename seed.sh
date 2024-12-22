#!/usr/bin/env bash

curl -f 'http://localhost:8043/account' -H 'Content-Type: application/json' -H "Authorization: Bearer $1"

team_a_uuid=$(curl -f 'http://localhost:8043/team' -H 'Content-Type: application/json' -H "Authorization: Bearer $1" -d '{"name": "Team A"}' | jq -r '.uuid')
team_b_uuid=$(curl -f 'http://localhost:8043/team' -H 'Content-Type: application/json' -H "Authorization: Bearer $1" -d '{"name": "Team B"}' | jq -r '.uuid')
league_uuid=$(curl -f 'http://localhost:8043/league' -H 'Content-Type: application/json' -H "Authorization: Bearer $1" -d "{\"name\": \"League A\", \"team_uuids\": [\"$team_a_uuid\", \"$team_b_uuid\"]}" | jq -r '.uuid')
curl -f -X "PUT" 'http://localhost:8043/organization-league' -H 'Content-Type: application/json' -H "Authorization: Bearer $1" -d "{\"league_uuids\": [\"$league_uuid\"]}"
starts_at=$(date -u -v+2H +"%Y-%m-%dT%H:%M:%SZ")
curl -f -X "POST" 'http://localhost:8043/match' -H 'Content-Type: application/json' -H "Authorization: Bearer $1" -d "{\"league_uuid\": \"$league_uuid\", \"away_team_uuid\": \"$team_a_uuid\", \"home_team_uuid\": \"$team_b_uuid\", \"starts_at\": \"$starts_at\"}"
