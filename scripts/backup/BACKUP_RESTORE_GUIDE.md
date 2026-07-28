# Panduan Backup & Restore Database SINDAS

> **Untuk Revisi Skripsi:** Sistem backup ini memenuhi prinsip integritas data (availability) dan perlindungan data pribadi spesifik (UU PDP No. 27/2022 Pasal 4 huruf b) mengingat database menyimpan riwayat diagnosis kesehatan mental siswa.

---

## Arsitektur Backup

| Tipe | Jadwal | Retensi | Direktori |
|------|--------|---------|-----------|
| **Mingguan** | Setiap Senin 02.00 WIB | 90 hari (3 bulan) | `/var/backups/sindas/weekly/` |
| **Tahunan** | Setiap 1 Jan 02.00 WIB | 1825 hari (5 tahun) | `/var/backups/sindas/annual/` |

Format file: `backup_sindas_<tipe>_<YYYYMMDD_HHMMSS>.dump.gpg`

Contoh: `backup_sindas_annual_20260101_020005.dump.gpg`

---

## Prasyarat

```bash
# Install di VPS 1 (jalankan sebagai root/sudo)
sudo apt-get update
sudo apt-get install -y postgresql-client gnupg

# Verifikasi
pg_dump --version
gpg --version
```

---

## Setup Sekali (One-Time Setup)

### 1. Upload skrip ke VPS

```bash
# Dari komputer lokal Anda
scp backup_sindas.sh user@IP_VPS_1:/opt/sindas/scripts/
ssh user@IP_VPS_1 "chmod +x /opt/sindas/scripts/backup_sindas.sh"
```

### 2. Buat direktori backup

```bash
sudo mkdir -p /var/backups/sindas/{weekly,annual}
sudo chown postgres:postgres /var/backups/sindas -R
sudo chmod 750 /var/backups/sindas -R

sudo touch /var/log/sindas_backup.log
sudo chown postgres:postgres /var/log/sindas_backup.log
sudo chmod 640 /var/log/sindas_backup.log
```

### 3. Setup passphrase GPG

> ⚠️ **Simpan passphrase ini di tempat yang aman terpisah** (password manager, tidak di VPS). Jika passphrase hilang, backup TIDAK bisa dibuka.

```bash
sudo mkdir -p /etc/sindas
# Ganti 'passphrase-rahasia-anda' dengan passphrase kuat (min 32 karakter)
echo "passphrase-rahasia-anda" | sudo tee /etc/sindas/gpg_passphrase > /dev/null
sudo chmod 600 /etc/sindas/gpg_passphrase
sudo chown postgres:postgres /etc/sindas/gpg_passphrase
```

### 4. Sesuaikan konfigurasi di skrip

Edit `/opt/sindas/scripts/backup_sindas.sh` dan sesuaikan variabel berikut:

```bash
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="sindas_db"    # ← Nama database PostgreSQL Anda yang sebenarnya
DB_USER="postgres"      # ← User PostgreSQL dengan akses SELECT
```

### 5. Daftarkan cron job

```bash
# Buka crontab untuk user postgres
sudo crontab -u postgres -e
```

Tambahkan baris berikut:

```cron
# ─── SINDAS Database Backup ─────────────────────────────────────────────
# Backup mingguan: setiap Senin pukul 02:00 WIB (UTC+7 = 19:00 UTC Minggu)
0 19 * * 0  /opt/sindas/scripts/backup_sindas.sh weekly >> /var/log/sindas_backup.log 2>&1

# Backup tahunan: setiap 1 Januari pukul 02:00 WIB (19:00 UTC 31 Des)
0 19 31 12 *  /opt/sindas/scripts/backup_sindas.sh annual >> /var/log/sindas_backup.log 2>&1
```

> **Catatan timezone:** VPS kemungkinan menggunakan UTC. Jam 02.00 WIB = 19.00 UTC (hari sebelumnya). Cek timezone VPS: `timedatectl`.
> Jika VPS sudah di-set ke WIB (`Asia/Jakarta`), gunakan: `0 2 * * 1` untuk mingguan dan `0 2 1 1 *` untuk tahunan.

### 6. Uji jalankan manual

```bash
# Test backup mingguan
sudo -u postgres /opt/sindas/scripts/backup_sindas.sh weekly

# Lihat log
tail -50 /var/log/sindas_backup.log

# Cek file backup terbuat
ls -lh /var/backups/sindas/weekly/
```

---

## Cara Memonitor Backup

### Cek log terakhir

```bash
# 50 baris terakhir log backup
tail -50 /var/log/sindas_backup.log

# Cek backup yang berhasil/gagal
grep -E "\[(ERROR|INFO)\].*BACKUP" /var/log/sindas_backup.log

# Cek kapan backup terakhir berjalan
grep "BACKUP BERHASIL" /var/log/sindas_backup.log | tail -5
```

### Daftar semua file backup

```bash
# Backup mingguan
ls -lh /var/backups/sindas/weekly/

# Backup tahunan
ls -lh /var/backups/sindas/annual/

# Total ukuran
du -sh /var/backups/sindas/
```

---

## Prosedur Restore

> ⚠️ **PENTING:** Selalu restore ke database SEMENTARA terlebih dahulu untuk memverifikasi isi backup — JANGAN langsung restore ke database produksi tanpa verifikasi.

### Langkah 1: Dekripsi file backup

```bash
# Ganti nama file sesuai backup yang ingin di-restore
BACKUP_FILE="/var/backups/sindas/weekly/backup_sindas_weekly_20260101_020005.dump.gpg"
DECRYPTED_FILE="/tmp/sindas_restore_$(date +%Y%m%d).dump"

gpg \
  --batch \
  --passphrase-file /etc/sindas/gpg_passphrase \
  --decrypt \
  --output "$DECRYPTED_FILE" \
  "$BACKUP_FILE"

echo "Dekripsi selesai: $DECRYPTED_FILE"
ls -lh "$DECRYPTED_FILE"
```

### Langkah 2: Buat database sementara

```bash
sudo -u postgres psql -c "CREATE DATABASE sindas_restore_test;"
```

### Langkah 3: Restore ke database sementara

```bash
sudo -u postgres pg_restore \
  --host=localhost \
  --port=5432 \
  --username=postgres \
  --dbname=sindas_restore_test \
  --no-password \
  --verbose \
  "$DECRYPTED_FILE" 2>&1 | tee /tmp/restore_log.txt
```

### Langkah 4: Verifikasi isi database hasil restore

```bash
# Masuk ke database sementara
sudo -u postgres psql -d sindas_restore_test

# Di dalam psql, jalankan query verifikasi:
```

```sql
-- Cek tabel utama
\dt

-- Cek jumlah data di tabel kritis
SELECT 'students'       AS tabel, COUNT(*) AS jumlah FROM students
UNION ALL
SELECT 'test_sessions',           COUNT(*) FROM test_sessions
UNION ALL
SELECT 'hasil_diagnoses',         COUNT(*) FROM hasil_diagnoses
UNION ALL
SELECT 'sekolahs',                COUNT(*) FROM sekolahs;

-- Cek data terbaru masih ada
SELECT created_at FROM test_sessions ORDER BY created_at DESC LIMIT 5;

-- Keluar dari psql
\q
```

### Langkah 5: Restore Parsial (opsional, jika hanya butuh satu tabel)

Format custom (`-Fc`) mendukung restore parsial:

```bash
# Hanya restore tabel hasil_diagnoses
sudo -u postgres pg_restore \
  --host=localhost \
  --port=5432 \
  --username=postgres \
  --dbname=sindas_restore_test \
  --table=hasil_diagnoses \
  --no-password \
  "$DECRYPTED_FILE"
```

### Langkah 6: Bersihkan setelah selesai

```bash
# Hapus file dekripsi sementara
rm -f "$DECRYPTED_FILE"

# Hapus database sementara jika sudah tidak diperlukan
sudo -u postgres psql -c "DROP DATABASE sindas_restore_test;"
```

---

## Restore ke Database Produksi (Emergency Only)

> ⚠️ **LAKUKAN HANYA JIKA DATABASE PRODUKSI RUSAK/HILANG.** Proses ini akan menghapus data yang ada di produksi.

```bash
# 1. Stop aplikasi backend terlebih dahulu agar tidak ada write selama restore
sudo systemctl stop sindas-backend  # Sesuaikan nama service

# 2. Drop dan recreate database produksi
sudo -u postgres psql -c "DROP DATABASE sindas_db;"
sudo -u postgres psql -c "CREATE DATABASE sindas_db;"

# 3. Restore
sudo -u postgres pg_restore \
  --host=localhost \
  --port=5432 \
  --username=postgres \
  --dbname=sindas_db \
  --no-password \
  --verbose \
  "$DECRYPTED_FILE"

# 4. Start kembali backend
sudo systemctl start sindas-backend
```

---

## Jadwal Verifikasi Berkala (Rekomendasi)

Backup yang tidak pernah diuji restore-nya **belum tentu bisa digunakan saat dibutuhkan**. Lakukan verifikasi periodik:

| Frekuensi | Aktivitas |
|-----------|-----------|
| **Setiap bulan** | Cek log backup: `grep "BACKUP BERHASIL" /var/log/sindas_backup.log \| tail -10` |
| **Setiap 3 bulan** | Lakukan restore ke database sementara (Langkah 1–6 di atas) dan verifikasi row count |
| **Setiap tahun** | Verifikasi backup tahunan bisa di-restore penuh, update passphrase GPG jika diperlukan |

---

## Keamanan & Kepatuhan UU PDP

| Aspek | Implementasi |
|-------|-------------|
| **Enkripsi data sensitif** | File backup dienkripsi AES256 via GPG sebelum disimpan |
| **Akses terbatas** | File backup hanya bisa dibaca oleh user `postgres` (`chmod 750`) |
| **Passphrase terproteksi** | Disimpan di `/etc/sindas/gpg_passphrase` dengan `chmod 600` |
| **Jejak audit** | Setiap proses backup tercatat di `/var/log/sindas_backup.log` |
| **Retensi terkontrol** | Penghapusan otomatis backup lama sesuai kebijakan retensi |

> Sesuai **UU PDP No. 27/2022 Pasal 35**, pengendali data wajib melindungi data pribadi dari kehilangan. Enkripsi backup dan retensi 5 tahun untuk arsip tahunan memenuhi prinsip ini untuk data kategori khusus (kesehatan mental).

---

## Troubleshooting

### Error: `pg_dump: error: connection to server failed`

```bash
# Cek PostgreSQL berjalan
sudo systemctl status postgresql

# Cek koneksi
sudo -u postgres psql -c "SELECT version();"
```

### Error: `gpg: decryption failed: Bad session key`

Passphrase salah atau file passphrase rusak. Cek isinya:

```bash
sudo cat /etc/sindas/gpg_passphrase | wc -c  # Hitung karakter, jangan tampilkan
```

### Cron tidak berjalan

```bash
# Cek cron service
sudo systemctl status cron

# Cek log cron
sudo grep CRON /var/log/syslog | tail -20

# Pastikan crontab terdaftar
sudo crontab -u postgres -l
```
