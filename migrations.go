package aura

import (
	"github.com/nrwiersma/aura/pkg/migrate"
)

// Schema is a database schema.
type Schema struct{}

func (s Schema) migrations() []migrate.Migration {
	return []migrate.Migration{
		{
			ID: 1,
			Up: migrate.Queries(
				`CREATE TABLE IF NOT EXISTS apps (
    id varchar(27) NOT NULL primary key,
    name varchar(50) NOT NULL,
    created_at datetime NOT NULL,
    deleted_at datetime
);`,
				`CREATE TABLE IF NOT EXISTS releases (
    id varchar(27) NOT NULL primary key,
    app_id varchar(27) NOT NULL references apps(id) ON DELETE CASCADE,
    image text NOT NULL,
    version int NOT NULL,
    procfile bytea NOT NULL,
    created_at datetime NOT NULL
);`,
			),
			Down: migrate.Queries(
				`DROP TABLE releases;`,
				`DROP TABLE apps;`,
			),
		},
	}
}
