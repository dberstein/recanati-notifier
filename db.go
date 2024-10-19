package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // Import driver (blank import for registration)
)

func ensureSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
-- Create tables and indexes
CREATE TABLE IF NOT EXISTS users (
	id        INTEGER PRIMARY KEY,
	name      TEXT,
	frequency INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS mediums (
	uid    INTEGER NOT NULL,
	type   TEXT NOT NULL,
	target TEXT
);

CREATE INDEX IF NOT EXISTS idx_medium_uid ON mediums (
	uid
);

CREATE TABLE IF NOT EXISTS notifications (
	id      INTEGER PRIMARY KEY,
	ts      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	type    INTEGER NOT NULL DEFAULT 0,
	subject TEXT,
	body    TEXT
);

CREATE INDEX IF NOT EXISTS idx_notification_type ON notifications (
    type
);

CREATE TABLE IF NOT EXISTS delivery (
	id      INTEGER PRIMARY KEY,
	ts      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	nid     INTEGER NOT NULL,
	uid     INTEGER NOT NULL,
	type    TEXT,
	target  TEXT,
	status  INTEGER NOT NULL DEFAULT 0,
	attempt INTEGER NOT NULL DEFAULT 0
);

-- Create trigger that upon new notification inserts deliveries per user/medium
CREATE TRIGGER IF NOT EXISTS insert_notification AFTER
INSERT ON notifications
BEGIN
	INSERT INTO delivery (nid, UID, TYPE, target)
	SELECT new.id,
		u.id,
		m.type,
		m.target
	FROM mediums m
    INNER JOIN users u ON u.id = m.uid;
END;

-- Create view that relates notifications with deliveries
CREATE VIEW IF NOT EXISTS deliveries AS
SELECT d.id,
       n.type AS ntype,
       d.type,
       d.attempt,
       d.target,
       n.subject,
       n.body,
	   n.id AS nid,
       d.uid,
       d.status
FROM delivery d
INNER JOIN notifications n ON n.id = d.nid;

-- Insert users
INSERT INTO users (id, name) VALUES (1, "first") ON CONFLICT DO NOTHING;
INSERT INTO users (id, name) VALUES (2, "second") ON CONFLICT DO NOTHING;
INSERT INTO users (id, name) VALUES (3, "third") ON CONFLICT DO NOTHING;

-- Insert users' mediums
INSERT INTO mediums (uid, type, target) VALUES (1, "email", "example@example.com");
INSERT INTO mediums (uid, type, target) VALUES (2, "email", "another@example.com");
INSERT INTO mediums (uid, type, target) VALUES (2, "sms", "0123456789");
INSERT INTO mediums (uid, type, target) VALUES (3, "sms", "9876543210");
	`); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func NewDb(dsn string) *sql.DB {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	err = ensureSchema(db)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
