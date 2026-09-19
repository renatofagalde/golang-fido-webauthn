-- ============================================================================
-- FIDO USER MODULE - schema
-- ============================================================================

-- Shared trigger function: keeps updated_at fresh on every UPDATE.
-- BEFORE UPDATE with RETURN NEW is the correct shape here (we mutate the row
-- being written). RETURN OLD would silently cancel the UPDATE.
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS fido_user (
    id           BIGSERIAL PRIMARY KEY,
    hash         VARCHAR(36) UNIQUE NOT NULL,
    username     VARCHAR(150) UNIQUE NOT NULL,
    display_name VARCHAR(150) NOT NULL,
    is_active    BOOLEAN DEFAULT TRUE,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP,
    created_by   BIGINT,
    updated_by   BIGINT
);

CREATE INDEX idx_fido_user_hash ON fido_user(hash);
CREATE INDEX idx_fido_user_username ON fido_user(username);
CREATE INDEX idx_fido_user_deleted_at ON fido_user(deleted_at);

COMMENT ON TABLE fido_user IS 'WebAuthn users. hash (UUIDv7) doubles as the WebAuthn user handle.';
COMMENT ON COLUMN fido_user.hash IS 'Public identifier and WebAuthn user handle source, carries no PII.';

CREATE TRIGGER update_fido_user_updated_at
    BEFORE UPDATE ON fido_user
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
