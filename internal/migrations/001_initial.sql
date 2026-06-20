CREATE TABLE copied (
    mapping   TEXT    NOT NULL,
    sha256    TEXT    NOT NULL,
    rel_path  TEXT    NOT NULL,
    size      INTEGER NOT NULL,
    mtime     INTEGER NOT NULL,
    copied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (mapping, rel_path)
);

CREATE INDEX idx_copied_hash ON copied (mapping, sha256);
