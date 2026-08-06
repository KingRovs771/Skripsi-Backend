#!/usr/bin/env bash
# =============================================================================
# backup_sindas.sh — Skrip Backup PostgreSQL untuk Sistem SINDAS
# =============================================================================
#
# Penggunaan:
#   ./backup_sindas.sh weekly [--job-id=<uuid>]
#   ./backup_sindas.sh annual [--job-id=<uuid>]
#
# Prasyarat di VPS:
#   sudo apt-get install -y postgresql-client gnupg
#
# Setup sekali (simpan passphrase GPG di file terproteksi):
#   echo "passphrase-rahasia-anda" | sudo tee /etc/sindas/gpg_passphrase > /dev/null
#   sudo chmod 600 /etc/sindas/gpg_passphrase
#   sudo chown postgres:postgres /etc/sindas/gpg_passphrase
#
# =============================================================================

set -euo pipefail

# ─── Konfigurasi — SESUAIKAN dengan environment Anda ─────────────────────────

DB_HOST="${SINDAS_DB_HOST:-localhost}"
DB_PORT="${SINDAS_DB_PORT:-5432}"
DB_NAME="${SINDAS_DB_NAME:-sindas_db}"   # Nama database PostgreSQL Anda
DB_USER="${SINDAS_DB_USER:-postgres}"    # User PostgreSQL
DB_PASS="${SINDAS_DB_PASS:-}"            # Password — kosong jika pakai peer/trust auth

BACKUP_DIR="/var/backups/sindas"
LOG_FILE="/var/log/sindas_backup.log"
GPG_PASSPHRASE_FILE="/etc/sindas/gpg_passphrase"  # File berisi passphrase GPG

# Retensi
WEEKLY_RETENTION_DAYS=90    # Simpan backup mingguan 3 bulan
ANNUAL_RETENTION_DAYS=1825  # Simpan backup tahunan 5 tahun (5 × 365)

# ─── Parse Argumen ───────────────────────────────────────────────────────────

JOB_ID=""
BACKUP_TYPE=""

for arg in "$@"; do
  if [[ "$arg" == "weekly" || "$arg" == "annual" ]]; then
    BACKUP_TYPE="$arg"
  elif [[ "$arg" =~ --job-id=(.+) ]]; then
    JOB_ID="${BASH_REMATCH[1]}"
  fi
done

if [[ -z "$BACKUP_TYPE" ]]; then
  echo "ERROR: Argumen harus 'weekly' atau 'annual'" >&2
  echo "Contoh: $0 weekly" >&2
  exit 1
fi

# ─── Fungsi logging ───────────────────────────────────────────────────────────

log() {
  local level="$1"
  shift
  local msg="$*"
  local ts
  ts="$(date '+%Y-%m-%d %H:%M:%S')"
  local line="[$ts] [$level] $msg"
  echo "$line"
  echo "$line" >> "$LOG_FILE"
}

# ─── Fungsi Update Status Database ───────────────────────────────────────────

update_job_status() {
  local status="$1"
  local file_path="${2:-}"
  local file_size="${3:-}"
  local error_msg="${4:-}"

  if [[ -z "${JOB_ID:-}" ]]; then
    return 0
  fi

  log "INFO" "Mengupdate status job $JOB_ID di database menjadi $status..."
  
  local sql=""
  if [[ "$status" == "RUNNING" ]]; then
    sql="UPDATE backup_jobs SET status = 'RUNNING', started_at = NOW() WHERE job_uid = '$JOB_ID';"
  elif [[ "$status" == "SUCCESS" ]]; then
    sql="UPDATE backup_jobs SET status = 'SUCCESS', finished_at = NOW(), file_path = '$file_path', file_size = '$file_size' WHERE job_uid = '$JOB_ID';"
  elif [[ "$status" == "FAILED" ]]; then
    local escaped_error_msg
    escaped_error_msg=$(echo "$error_msg" | sed "s/'/''/g")
    sql="UPDATE backup_jobs SET status = 'FAILED', finished_at = NOW(), error_message = '$escaped_error_msg' WHERE job_uid = '$JOB_ID';"
  fi

  if [[ -z "$sql" ]]; then
    return 0
  fi

  # Nonaktifkan ERR trap sementara agar tidak looping jika psql gagal
  trap - ERR
  set +e

  # Gunakan PGPASSWORD dari env var SINDAS_DB_PASS jika tersedia
  PGPASSWORD="${DB_PASS:-}" PGPASSFILE="" \
    psql \
      -h "$DB_HOST" \
      -p "$DB_PORT" \
      -U "$DB_USER" \
      -d "$DB_NAME" \
      -c "$sql" \
      -X \
      -t 2>>"$LOG_FILE" || log "WARNING" "Gagal mengupdate database via psql — Go goroutine akan mendeteksi dan menangani ini"

  set -e
  trap 'on_error $LINENO' ERR
}

# ─── Error Handler Trap ───────────────────────────────────────────────────────

on_error() {
  local exit_code="$?"
  local line_number="$1"
  log "ERROR" "Terjadi error pada baris $line_number dengan exit code $exit_code."
  update_job_status "FAILED" "" "" "Error pada baris $line_number dengan exit code $exit_code"
  
  if [[ -f "${DUMP_FILE:-}" ]]; then
    rm -f "$DUMP_FILE"
  fi
  exit "$exit_code"
}

trap 'on_error $LINENO' ERR

# ─── Setup direktori ──────────────────────────────────────────────────────────

SUBDIR="$BACKUP_DIR/$BACKUP_TYPE"
mkdir -p "$SUBDIR"
chmod 750 "$SUBDIR"

# ─── Nama file backup ─────────────────────────────────────────────────────────

TIMESTAMP="$(date '+%Y%m%d_%H%M%S')"
DUMP_FILE="$SUBDIR/backup_sindas_${BACKUP_TYPE}_${TIMESTAMP}.dump"
ENCRYPTED_FILE="${DUMP_FILE}.gpg"

# CATATAN: Status RUNNING sudah di-set oleh Go goroutine sebelum script ini dipanggil.
# Script bash hanya perlu mengupdate SUCCESS atau FAILED di akhir.

# ─── Cek prasyarat ───────────────────────────────────────────────────────────

if ! command -v pg_dump &>/dev/null; then
  log "ERROR" "pg_dump tidak ditemukan. Install: sudo apt-get install postgresql-client"
  exit 1
fi

if ! command -v gpg &>/dev/null; then
  log "ERROR" "gpg tidak ditemukan. Install: sudo apt-get install gnupg"
  exit 1
fi

# Fallback: jika file passphrase tidak ada tetapi env var SINDAS_GPG_PASSPHRASE diset
if [[ ! -f "$GPG_PASSPHRASE_FILE" && -n "${SINDAS_GPG_PASSPHRASE:-}" ]]; then
  log "INFO" "File GPG passphrase tidak ditemukan. Membuat file temporer dari env SINDAS_GPG_PASSPHRASE..."
  mkdir -p "$(dirname "$GPG_PASSPHRASE_FILE")" 2>/dev/null || GPG_PASSPHRASE_FILE="/tmp/gpg_passphrase"
  echo "$SINDAS_GPG_PASSPHRASE" > "$GPG_PASSPHRASE_FILE"
  chmod 600 "$GPG_PASSPHRASE_FILE"
fi

if [[ ! -f "$GPG_PASSPHRASE_FILE" ]]; then
  log "ERROR" "File passphrase GPG tidak ditemukan di: $GPG_PASSPHRASE_FILE"
  log "ERROR" "Buat dengan: echo 'passphrase-anda' | sudo tee $GPG_PASSPHRASE_FILE atau set env var SINDAS_GPG_PASSPHRASE"
  exit 1
fi

# ─── Mulai backup ─────────────────────────────────────────────────────────────

log "INFO" "========================================================"
log "INFO" "Mulai backup SINDAS — Tipe: $BACKUP_TYPE"
log "INFO" "Database: $DB_NAME @ $DB_HOST:$DB_PORT"
log "INFO" "Output: $ENCRYPTED_FILE"

START_TIME="$(date +%s)"

# pg_dump dengan format custom (-Fc):
log "INFO" "Menjalankan pg_dump..."

PGPASSWORD="${DB_PASS:-}" PGPASSFILE="" \
  pg_dump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --format=custom \
    --compress=9 \
    --file="$DUMP_FILE" 2>>"$LOG_FILE"

DUMP_SIZE="$(du -sh "$DUMP_FILE" | cut -f1)"
log "INFO" "pg_dump selesai. Ukuran dump: $DUMP_SIZE"

# ─── Enkripsi dengan GPG (AES256, simetris) ───────────────────────────────────

log "INFO" "Mengenkripsi file backup dengan GPG AES256..."

gpg \
  --batch \
  --yes \
  --passphrase-file "$GPG_PASSPHRASE_FILE" \
  --symmetric \
  --cipher-algo AES256 \
  --output "$ENCRYPTED_FILE" \
  "$DUMP_FILE"

# Hapus dump mentah setelah enkripsi berhasil
rm -f "$DUMP_FILE"

ENCRYPTED_SIZE="$(du -sh "$ENCRYPTED_FILE" | cut -f1)"
log "INFO" "Enkripsi selesai. Ukuran terenkripsi: $ENCRYPTED_SIZE"

# ─── Verifikasi integritas: pastikan file GPG bisa dibuka ─────────────────────

log "INFO" "Memverifikasi integritas file backup..."

gpg \
  --batch \
  --yes \
  --passphrase-file "$GPG_PASSPHRASE_FILE" \
  --decrypt \
  --output /dev/null \
  "$ENCRYPTED_FILE" 2>>"$LOG_FILE"

log "INFO" "Verifikasi integritas: OK"

# ─── Pembersihan backup lama (retensi otomatis) ───────────────────────────────

if [[ "$BACKUP_TYPE" == "weekly" ]]; then
  RETENTION="$WEEKLY_RETENTION_DAYS"
else
  RETENTION="$ANNUAL_RETENTION_DAYS"
fi

log "INFO" "Membersihkan backup $BACKUP_TYPE yang lebih lama dari $RETENTION hari..."

DELETED_COUNT=0
while IFS= read -r -d '' old_file; do
  rm -f "$old_file"
  log "INFO" "Dihapus (kadaluarsa): $old_file"
  DELETED_COUNT=$((DELETED_COUNT + 1))
done < <(find "$SUBDIR" -name "*.dump.gpg" -mtime +"$RETENTION" -print0)

if [[ "$DELETED_COUNT" -eq 0 ]]; then
  log "INFO" "Tidak ada backup lama yang perlu dihapus."
else
  log "INFO" "Total backup lama dihapus: $DELETED_COUNT file"
fi

# ─── Ringkasan akhir & Update DB SUCCESS ──────────────────────────────────────

END_TIME="$(date +%s)"
DURATION=$(( END_TIME - START_TIME ))

log "INFO" "========================================================"
log "INFO" "BACKUP BERHASIL"
log "INFO" "  Tipe         : $BACKUP_TYPE"
log "INFO" "  Database     : $DB_NAME"
log "INFO" "  File backup  : $ENCRYPTED_FILE"
log "INFO" "  Ukuran       : $ENCRYPTED_SIZE"
log "INFO" "  Durasi       : ${DURATION}s"
log "INFO" "  Timestamp    : $(date '+%Y-%m-%d %H:%M:%S')"
log "INFO" "========================================================"

update_job_status "SUCCESS" "$ENCRYPTED_FILE" "$ENCRYPTED_SIZE"

exit 0