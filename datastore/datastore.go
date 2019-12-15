package datastore

import (
	"database/sql"

	// We support only postgres
	_ "github.com/lib/pq"
)

// Connection is datastore connection
type Connection struct {
	db *sql.DB
}

// New returns new datastore connection
func New(dbURL string) (*Connection, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS records (
			ts timestamp NOT NULL DEFAULT current_timestamp,
			hostname text NOT NULL,
			pid bigint NOT NULL,
			cgroup text NOT NULL,
			nspid bigint
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS records_ts ON records (ts)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS records_hostname_pid ON records (hostname, pid)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`CREATE OR REPLACE FUNCTION add_record2(_hostname text, _pid bigint, _cgroup text, _nspid bigint) RETURNS VOID AS $$
		DECLARE
			last_cgroup text;
		BEGIN
			BEGIN
				SELECT cgroup
					INTO STRICT last_cgroup
					FROM records
					WHERE hostname = _hostname AND pid = _pid
					ORDER BY ts DESC
					LIMIT 1;
			EXCEPTION WHEN NO_DATA_FOUND THEN
				last_cgroup := '';
			END;

			IF last_cgroup <> _cgroup THEN
				INSERT INTO records(hostname, pid, cgroup, nspid) VALUES(_hostname, _pid, _cgroup, _nspid);
			END IF;
		END;
		$$ LANGUAGE plpgsql`)
	if err != nil {
		return nil, err
	}

	return &Connection{
		db: db,
	}, nil
}

// Write writes to datastore
func (c *Connection) Write(hostname string, pid uint64, cgroup string, nspid uint64) error {
	_, err := c.db.Exec(`SELECT add_record2($1, $2, $3, $4)`, hostname, pid, cgroup, nspid)
	return err
}

// Close closes datastore connection and frees all resources
func (c *Connection) Close() error {
	return c.db.Close()
}
