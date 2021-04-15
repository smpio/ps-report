package datastore

import (
	"database/sql"

	// We support only postgres
	_ "github.com/lib/pq"

	"github.com/smpio/ps-report/process"
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
			nspid bigint,
			vm_peak bigint,
			vm_size bigint,
			vm_lck bigint,
			vm_pin bigint,
			vm_hwm bigint,
			vm_rss bigint,
			rss_anon bigint,
			rss_file bigint,
			rss_shmem bigint,
			cmd text,
			seq_id integer
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS records_ts ON records (ts)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS records_hostname_pid_seq_id ON records (hostname, pid, seq_id)`)
	if err != nil {
		return nil, err
	}

	return &Connection{
		db: db,
	}, nil
}

// Write writes to datastore
func (c *Connection) Write(hostname string, p *process.Process) error {
	_, err := c.db.Exec(`INSERT INTO records(hostname, pid, cgroup, nspid, vm_peak, vm_size, vm_lck, vm_pin, vm_hwm, vm_rss, rss_anon,
							rss_file, rss_shmem, cmd, seq_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		hostname, p.Pid, p.Cgroup, p.NSpid, p.VmPeak, p.VmSize, p.VmLck, p.VmPin, p.VmHWM, p.VmRSS, p.RssAnon,
		p.RssFile, p.RssShmem, p.Cmd, p.SeqID)
	return err
}

// Close closes datastore connection and frees all resources
func (c *Connection) Close() error {
	return c.db.Close()
}
