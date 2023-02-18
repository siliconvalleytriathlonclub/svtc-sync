PRAGMA foreign_keys = ON;

CREATE TABLE member (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    clubexpress_id INTEGER,
    num INTEGER NOT NULL,
    active INTEGER NOT NULL,
    firstname TEXT NOT NULL,
    middle TEXT,
    lastname TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    address TEXT,
    addr_ext TEXT,
    city TEXT,
    state TEXT,
    zip TEXT,
    mobile TEXT,
    status TEXT NOT NULL,
    joined TEXT NOT NULL,
    expired TEXT NOT NULL,   
    UNIQUE(num, firstname, lastname, email)
);

CREATE INDEX idx_member ON member(id);

CREATE TABLE id_map (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    clubexpress_id INTEGER,
    strava_id INTEGER,
    slack_id TEXT
    FOREIGN KEY(clubexpress_id) REFERENCES member(clubexpress_id),
);

