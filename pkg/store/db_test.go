package store_test

import (
	"io"
	"testing"

	"github.com/hamba/logger/v2"
	"github.com/nrwiersma/aura/pkg/store"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
)

func testDB(t *testing.T) *store.Database {
	t.Helper()

	log := logger.New(io.Discard, logger.LogfmtFormat(), logger.Debug)

	db, err := store.NewDatabase(sqlite.Open("file::memory:"), log)
	require.NoError(t, err)

	err = db.IsHealthy()
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	return db
}
