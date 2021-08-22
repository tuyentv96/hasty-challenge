package migration

import (
	"database/sql"
	"embed"
	"log"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const driver = "postgres"

func Migrate(dataSource string) error {
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		log.Printf("migration: Could not open sql: %s\n", err)
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("migration: failed to close sql db: %s\n", err)
		}
	}()

	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFS,
		Root:       "migrations",
	}

	_, err = migrate.Exec(db, driver, migrationSource, migrate.Up)
	if err != nil {
		log.Printf("migration: Could not migrate up: %s\n", err)
		return err
	}

	return nil
}
