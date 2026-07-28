CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    last_active TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE metadata (
    id SERIAL PRIMARY KEY,
    parent_id INT REFERENCES metadata(id) ON DELETE CASCADE,
    owner_id INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    object_key TEXT,  -- RustFS object key; NULL for directories
    file_type TEXT NOT NULL,
    is_file BOOLEAN DEFAULT FALSE,
    name TEXT NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_metadata_owner_id ON metadata(owner_id);
CREATE INDEX idx_metadata_parent_id ON metadata(parent_id);
CREATE INDEX idx_metadata_active ON metadata(id) WHERE deleted_at IS NULL;

-- Uniqueness within a folder
CREATE UNIQUE INDEX uq_metadata_name_in_folder
    ON metadata(parent_id, name)
    WHERE parent_id IS NOT NULL AND deleted_at IS NULL;
    
-- Uniqueness at home/root level
CREATE UNIQUE INDEX uq_metadata_name_at_root
    ON metadata(owner_id, name)
    WHERE parent_id IS NULL AND deleted_at IS NULL;

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    metadata_id INT NOT NULL REFERENCES metadata(id) ON DELETE CASCADE,
    account_id INT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    access_level INT NOT NULL,
    deleted_at TIMESTAMPTZ
);
