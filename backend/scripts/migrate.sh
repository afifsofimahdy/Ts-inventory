#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIG_DIR="$ROOT_DIR/migrations"
SEED_DIR="$ROOT_DIR/seed"

if [[ -f "$ROOT_DIR/.env" ]]; then
  # shellcheck disable=SC1091
  source "$ROOT_DIR/.env"
fi

DB_URL="${DATABASE_URL:-}"
if [[ -z "$DB_URL" ]]; then
  echo "DATABASE_URL is required (set in backend/.env)" >&2
  exit 1
fi

PSQL=(psql "$DB_URL" -v ON_ERROR_STOP=1 -X)

ensure_table() {
  "${PSQL[@]}" -c "CREATE TABLE IF NOT EXISTS schema_migrations (version INT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT NOW());"
}

current_version() {
  "${PSQL[@]}" -tA -c "SELECT COALESCE(MAX(version), 0) FROM schema_migrations;"
}

apply_up() {
  ensure_table
  local cur
  cur=$(current_version)
  local -a files=()
  if [[ -d "$MIG_DIR" ]]; then
    while IFS= read -r -d '' f; do
      files+=("$f")
    done < <(find "$MIG_DIR" -maxdepth 1 -name "*.up.sql" -print0 | sort -z)
  fi
  if (( ${#files[@]} == 0 )); then
    echo "no migrations found"
    return 0
  fi
  for f in "${files[@]}"; do
    local base version
    base=$(basename "$f")
    version=${base%%_*}
    if (( version > cur )); then
      "${PSQL[@]}" -c "BEGIN;" \
        -f "$f" \
        -c "INSERT INTO schema_migrations(version) VALUES ($version);" \
        -c "COMMIT;"
      echo "applied $version"
    fi
  done
}

apply_down() {
  ensure_table
  local steps=${1:-1}
  if ! [[ "$steps" =~ ^[0-9]+$ ]]; then
    echo "steps must be number" >&2
    exit 1
  fi
  local versions
  versions=$("${PSQL[@]}" -tA -c "SELECT version FROM schema_migrations ORDER BY version DESC;" | head -n "$steps")
  if [[ -z "$versions" ]]; then
    echo "no migrations to rollback"
    return 0
  fi
  while IFS= read -r ver; do
    [[ -z "$ver" ]] && continue
    local match=""
    if [[ -d "$MIG_DIR" ]]; then
      while IFS= read -r -d '' f; do
        local base vstr
        base=$(basename "$f")
        vstr=${base%%_*}
        if [[ "$vstr" =~ ^[0-9]+$ ]] && (( 10#$vstr == ver )); then
          match="$f"
          break
        fi
      done < <(find "$MIG_DIR" -maxdepth 1 -name "*.down.sql" -print0 | sort -z)
    fi
    if [[ -z "$match" ]]; then
      echo "missing down file for version $ver" >&2
      exit 1
    fi
    "${PSQL[@]}" -c "BEGIN;" \
      -f "$match" \
      -c "DELETE FROM schema_migrations WHERE version=$ver;" \
      -c "COMMIT;"
    echo "rolled back $ver"
  done <<< "$versions"
}

run_seed() {
  local -a files=()
  if [[ -d "$SEED_DIR" ]]; then
    while IFS= read -r -d '' f; do
      files+=("$f")
    done < <(find "$SEED_DIR" -maxdepth 1 -name "*.sql" -print0 | sort -z)
  fi
  if (( ${#files[@]} == 0 )); then
    echo "no seed files found"
    return 0
  fi
  for f in "${files[@]}"; do
    "${PSQL[@]}" -f "$f"
    echo "seeded $(basename "$f")"
  done
}

status() {
  ensure_table
  local cur
  cur=$(current_version)
  echo "current_version=$cur"
}

cmd=${1:-}
case "$cmd" in
  up) apply_up ;;
  down) apply_down "${2:-1}" ;;
  seed) run_seed ;;
  status) status ;;
  *)
    echo "Usage: $0 {up|down [steps]|seed|status}" >&2
    exit 1
    ;;
esac
