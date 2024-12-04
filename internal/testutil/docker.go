package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type TestDB struct {
	ContainerID string
	DB          *sql.DB
}

func SetupTestDB() (*TestDB, error) {
	return setupDB("mysql", "3306", []string{
		"-e", "MYSQL_ROOT_PASSWORD=password",
		"-e", "MYSQL_DATABASE=test_export",
	}, "mysql:8.0",
		"root:password@tcp(localhost:3306)/test_export?charset=utf8mb4",
		"../../testdata/init.sql")
}

func SetupTestPostgresDB() (*TestDB, error) {
	return setupDB("postgres", "5432", []string{
		"-e", "POSTGRES_PASSWORD=password",
		"-e", "POSTGRES_DB=test_export",
	}, "postgres:14",
		"host=localhost port=5432 user=postgres password=password dbname=test_export sslmode=disable",
		"../../testdata/init_postgres.sql")
}

func setupDB(driver, port string, envVars []string, image, dsn, initSQL string) (*TestDB, error) {
	sqlFile, err := filepath.Abs(initSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(sqlFile); err != nil {
		return nil, fmt.Errorf("init.sql not found: %w", err)
	}

	args := []string{"run", "--rm", "-d"}
	args = append(args, envVars...)
	args = append(args, "-v", fmt.Sprintf("%s:/docker-entrypoint-initdb.d/init.sql", sqlFile))
	args = append(args, "-p", fmt.Sprintf("%s:%s", port, port))
	args = append(args, image)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start %s container: %v, output: %s", driver, err, output)
	}

	containerID := string(output)[:12]

	time.Sleep(5 * time.Second)

	var db *sql.DB
	for i := 0; i < 60; i++ {
		db, err = sql.Open(driver, dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
			db.Close()
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		exec.Command("docker", "stop", containerID).Run()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &TestDB{
		ContainerID: containerID,
		DB:          db,
	}, nil
}

func (tdb *TestDB) Cleanup() error {
	if tdb.DB != nil {
		tdb.DB.Close()
	}

	cmd := exec.Command("docker", "stop", tdb.ContainerID)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop container: %v, output: %s", err, output)
	}

	return nil
}
