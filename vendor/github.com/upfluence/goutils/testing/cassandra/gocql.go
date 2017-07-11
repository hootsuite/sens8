package cassandra

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocql/gocql"
	_ "github.com/mattes/migrate/driver/cassandra"
	"github.com/mattes/migrate/migrate"
)

const (
	defaultCassandraIP = "127.0.0.1"
	defaultKeyspace    = "test"
	defaultTimeout     = time.Minute
	cassandraPort      = 9042
	protocolVersion    = 3
)

var keyspaceMutex = &sync.Mutex{}

func BuildKeySpace(migrationsPath *string, tables []string) (*gocql.Session, *sync.Mutex) {
	var (
		cassandraIP = os.Getenv("CASSANDRA_IP")
		keySpace    = os.Getenv("CASSANDRA_KEY_SPACE")
	)

	if cassandraIP == "" {
		cassandraIP = defaultCassandraIP
	}

	if keySpace == "" {
		keySpace = defaultKeyspace
	}

	cluster := gocql.NewCluster(cassandraIP)
	cluster.Consistency = gocql.All
	cluster.ProtoVersion = protocolVersion
	cluster.Timeout = defaultTimeout

	session, err := cluster.CreateSession()

	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.Query(
		`CREATE KEYSPACE ` + keySpace + ` WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`,
	).Exec()

	if migrationsPath != nil {
		errs, ok := migrate.UpSync(
			fmt.Sprintf(
				"cassandra://%s:%d/%s?protocol=%d",
				cassandraIP,
				cassandraPort,
				keySpace,
				protocolVersion,
			),
			*migrationsPath,
		)

		if !ok {
			strErrs := []string{}
			for _, migrationError := range errs {
				strErrs = append(strErrs, migrationError.Error())
			}

			panic(strings.Join(strErrs, ","))
		}
	}

	keyspaceMutex.Lock()

	cluster.Keyspace = keySpace
	if session, err := cluster.CreateSession(); err != nil {
		panic(err)
	} else {
		for _, table := range tables {
			if err := session.Query(
				fmt.Sprintf("TRUNCATE %s", table),
			).Exec(); err != nil {
				panic(err)
			}
		}

		return session, keyspaceMutex
	}
}
