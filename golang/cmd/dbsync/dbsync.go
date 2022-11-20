package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type DnsServer struct {
	addr string
	desc string
}

type DnsLookupResult struct {
	ret         int
	ip          string
	lookup_time float64
}

func current_time_and_date_str2() string {
	s := time.Now().Format("2006-01-02 15:04:05.000")
	return s
}

func get_uuid() string {
	v, _ := uuid.NewRandom()
	return v.String()
}

func ldb_key(group string, id string) string {
	return fmt.Sprintf("%s|%s", group, id)
}

type ServerRecord struct {
	Addr string `json:"addr"`
	Desc string `json:"desc"`
}
type OutputRecord struct {
	// {
	// 	"server"        : { "addr" : server_addr, "desc" : server_desc },
	// 	"at"            : now_str,
	// 	"rid"           : ruuid_str,
	// 	"counter"       : counter,
	// 	"id"            : uuid_str,
	// 	"domain"        : domain,
	// 	"lookup_time"   : lres.lookup_time,
	// 	"lookup_ip"     : lres.ip,
	// }
	Server     ServerRecord `json:"server"`
	At         string       `json:"at"`
	Rid        string       `json:"rid"`
	Counter    int          `json:"counter"`
	Id         string       `json:"id"`
	Domain     string       `json:"domain"`
	LookupTime float64      `json:"lookup_time"`
	LookupIp   string       `json:"lookup_ip"`
}

const (
	create_table_perf_str string = `
		CREATE TABLE IF NOT EXISTS perf (
			Id          TEXT PRIMARY KEY,
			Rid         TEXT NOT NULL,
			Counter     INTEGER NOT NULL,
			At			TEXT NOT NULL,
			ServerAddr  TEXT NOT NULL,
			ServerDesc  TEXT NOT NULL,
			Domain      TEXT NOT NULL,
			LookupTime  REAL NOT NULL,
			LookupIp    TEXT
		);
	`

	insert_table_perf_str_NAMED string = `
		INSERT OR IGNORE INTO perf (Id, Rid, Counter, At, ServerAddr, ServerDesc, Domain, LookupTime, LookupIp)
		VALUES ($Id, $Rid, $Counter, $At, $ServerAddr, $ServerDesc, $Domain, $LookupTime, $LookupIp);
	`

	insert_table_perf_str string = `
	INSERT OR IGNORE INTO perf (Id, Rid, Counter, At, ServerAddr, ServerDesc, Domain, LookupTime, LookupIp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`
)

func run_sync(ldb_path string, sqldb_path string, verbose bool) {
	fmt.Printf("run_sync, ldbpath: %v, sqldbpath: %v\n", ldb_path, sqldb_path)

	// Input db
	// Set up database connection information and open database
	o := &opt.Options{
		Compression: opt.NoCompression,
	}
	input_db, err := leveldb.OpenFile(ldb_path, o)
	if err != nil {
		log.Fatal(err)
	}
	defer input_db.Close()
	// wo := opt.WriteOptions{}

	// Output db
	output_db, err := sql.Open("sqlite3", sqldb_path)
	fmt.Printf("run_sync, sqldb: %v\n", output_db)
	if err != nil {
		log.Fatal(err)
	}

	_, err = output_db.Exec(create_table_perf_str)
	if err != nil {
		log.Printf("%q: %s\n", err, create_table_perf_str)
		return
	}
	defer output_db.Close()

	// insert into output db
	var recordCounter int = 0
	iter := input_db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := string(iter.Key())
		value := string(iter.Value())
		_ = value

		//fmt.Printf("DEBUG run_sync, iter loop %d, saw key: <%v>\n", recordCounter, key)

		var r OutputRecord
		err := json.Unmarshal([]byte(value), &r)
		if err != nil {
			log.Printf("%q: could not decode json data for key: <%v>\n", err, key)
			// continue;
			break
		}
		//fmt.Printf("DEBUG run_sync, iter loop %d, record: <%v>\n", recordCounter, r)

		// insert into db
		_, err = output_db.Exec(insert_table_perf_str,
			r.Id, r.Rid, r.Counter, r.At, r.Server.Addr, r.Server.Desc, r.Domain, r.LookupTime, r.LookupIp)
		if err != nil {
			log.Printf("%q: could not insert into db for key: <%v>\n", err, key)
			break
		}

		recordCounter++
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// parse command line arguments
	var ldbpath, sqldbpath string
	var verbose bool
	flag.StringVar(&ldbpath, "ldbpath", "dnsperfdb", "Level DB path")
	flag.StringVar(&sqldbpath, "sqldbpath", "dns.db", "SQLite DB path")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.Parse()

	fmt.Printf("ldbpath: %v\n", ldbpath)

	run_sync(ldbpath, sqldbpath, verbose)
}
