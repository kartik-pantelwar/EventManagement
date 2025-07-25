package migrate

import (
	sqlparser "authservice/src/pkg/sqlParser"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

type Migrate struct {
	path           string
	db             *sql.DB
	migrationFiles []DirEntryWithPrefix
	txn            *sql.Tx
}

func NewMigrate(db *sql.DB, dirPath string) Migrate {
	return Migrate{
		db:   db,
		path: dirPath,
	}
}

func (m *Migrate) RunMigrations() error {
	rawEntries, err := os.ReadDir(m.path)
	if err != nil {
		log.Println("Failed to read directories from path")
		return err
	}
	usableEntries := m.filterSqlFilesWithNumberPrefix(m.getFilesFromDirEntries(rawEntries))

	m.sortDirEntryBasedOnPrefix(usableEntries)

	err = m.checkForSamePrefix(usableEntries)
	if err != nil {
		log.Println("Found Same Migration Prefix")
		return err
	}

	version, err := m.getVersion()
	if err != nil {
		log.Println("Failed to get Version of Migrations")
		return err
	}

	if version == len(usableEntries) {
		return nil
	}

	if version == -1 {
		m.migrationFiles = usableEntries
	} else {
		m.migrationFiles = usableEntries[version:]
	}

	m.txn, err = m.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer m.txn.Rollback()

	err = m.parseFilesAndMigrateDb()
	if err != nil {
		return err
	}

	//clear file
	err = os.Truncate(m.path+"/migrate.log", 0)
	if err != nil {
		return err
	}

	//writing latest db version to file
	latest := []byte(fmt.Sprintf("%d", len(usableEntries)))
	outFile, err2 := os.OpenFile(m.path+"/migrate.log", os.O_RDWR, 0777)
	if err2 != nil {
		return err
	}
	_, err2 = outFile.Write(latest)
	if err2 != nil {
		return err
	}

	err = outFile.Close()
	if err != nil {
		return err
	}

	err = m.txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m *Migrate) parseFilesAndMigrateDb() error {
	for _, file := range m.migrationFiles {
		filePath := m.path + "/" + file.Dir.Name()
		fmt.Printf("Reading File %s\n", file.Dir.Name())
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Bypassing unterminated Dollar error in Trigger
		content := string(bytes)
		if strings.Contains(content, "CREATE OR REPLACE") || strings.Contains(content, "CREATE FUNCTION") {
			fmt.Println("Bypassing function/trigger,", file.Dir.Name())

			_, err := m.txn.Exec(content)
			if err != nil {
				fmt.Println("Error in bypassing :", err)
				return err
			}
			continue
		}
		commands := sqlparser.ParseSqlFile(content)
		for _, command := range commands {
			_, err = m.txn.Exec(command)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
