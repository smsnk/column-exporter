package exporter

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

type Exporter struct {
	db        *sql.DB
	batchSize int
	output    string
	prefix    string
	extension string
	nameCol   string
}

func New(db *sql.DB, batchSize int, output, prefix, extension, nameCol string) *Exporter {
	return &Exporter{
		db:        db,
		batchSize: batchSize,
		output:    output,
		prefix:    prefix,
		extension: extension,
		nameCol:   nameCol,
	}
}

func (e *Exporter) Export(table, column string) error {
	pkColumn, err := e.getPrimaryKeyColumn(table)
	if err != nil {
		return fmt.Errorf("failed to get primary key: %w", err)
	}

	nameColumn := pkColumn
	if e.nameCol != "" {
		nameColumn = e.nameCol
	}

	query := fmt.Sprintf("SELECT %s, %s FROM %s", nameColumn, column, table)

	rows, err := e.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}
	defer rows.Close()

	usedNames := make(map[string]struct{})

	for rows.Next() {
		var (
			nameValue  sql.NullString
			binaryData []byte
		)

		if err := rows.Scan(&nameValue, &binaryData); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		filename := fmt.Sprintf("%s%s", nameValue.String, e.extension)
		if _, exists := usedNames[filename]; exists {
			return fmt.Errorf("duplicate filename detected: %s", filename)
		}
		usedNames[filename] = struct{}{}

		if err := os.WriteFile(
			filepath.Join(e.output, filename),
			binaryData,
			0644,
		); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error during row iteration: %w", err)
	}

	return nil
}

func (e *Exporter) getPrimaryKeyColumn(table string) (string, error) {
	var pkColumn string
	var query string

	// ドライバー名の取得
	driverName := ""
	if _, ok := e.db.Driver().(*mysql.MySQLDriver); ok {
		driverName = "mysql"
	} else if _, ok := e.db.Driver().(*pq.Driver); ok {
		driverName = "postgres"
	}

	switch driverName {
	case "mysql":
		query = `
			SELECT COLUMN_NAME
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND CONSTRAINT_NAME = 'PRIMARY'
			LIMIT 1
		`
	case "postgres":
		query = `
			SELECT a.attname
			FROM pg_index i
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
			WHERE i.indrelid = $1::regclass
			AND i.indisprimary
		`
	default:
		return "", fmt.Errorf("unsupported database driver")
	}

	var err error
	if driverName == "postgres" {
		err = e.db.QueryRow(query, table).Scan(&pkColumn)
	} else {
		err = e.db.QueryRow(query, table).Scan(&pkColumn)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get primary key column: %w", err)
	}

	return pkColumn, nil
}
