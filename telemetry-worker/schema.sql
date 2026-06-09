CREATE TABLE IF NOT EXISTS installs (
  install_id_hash TEXT PRIMARY KEY,
  version TEXT NOT NULL,
  commit_sha TEXT,
  server_os TEXT NOT NULL,
  server_arch TEXT NOT NULL,
  clients_total INTEGER NOT NULL,
  clients_active_7d INTEGER NOT NULL,
  gpus_total INTEGER NOT NULL,
  gpus_active_7d INTEGER NOT NULL,
  first_seen_epoch INTEGER NOT NULL,
  last_seen_epoch INTEGER NOT NULL,
  reported_at_epoch INTEGER NOT NULL,
  report_count INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_installs_last_seen_epoch ON installs(last_seen_epoch);
