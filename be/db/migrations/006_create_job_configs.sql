CREATE TABLE IF NOT EXISTS job_configs (
    job_name          VARCHAR(100) PRIMARY KEY,
    enabled           BOOLEAN     NOT NULL DEFAULT TRUE,
    interval_seconds  INTEGER     NOT NULL,
    retention_seconds INTEGER     NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_job_configs_interval_positive CHECK (interval_seconds > 0),
    CONSTRAINT chk_job_configs_retention_non_negative CHECK (retention_seconds >= 0)
);

INSERT INTO job_configs (job_name, enabled, interval_seconds, retention_seconds)
VALUES ('auth_session_cleanup', TRUE, 86400, 604800)
ON CONFLICT (job_name) DO NOTHING;

CREATE OR REPLACE FUNCTION notify_job_configs_changed()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM pg_notify('job_configs_changed', OLD.job_name);
        RETURN OLD;
    END IF;

    PERFORM pg_notify('job_configs_changed', NEW.job_name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_job_configs_changed ON job_configs;

CREATE TRIGGER trg_job_configs_changed
AFTER INSERT OR UPDATE OR DELETE ON job_configs
FOR EACH ROW
EXECUTE FUNCTION notify_job_configs_changed();
