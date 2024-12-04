package main

import (
	"fmt"
	"os"

	"github.com/smsnk/column-exporter/internal/db"
	"github.com/smsnk/column-exporter/internal/exporter"
	"github.com/spf13/cobra"
)

type exportOptions struct {
	driver    string
	host      string
	port      int
	user      string
	password  string
	database  string
	table     string
	column    string
	output    string
	batchSize int
	prefix    string
	extension string
	nameCol   string
	schema    string
}

func main() {
	opts := &exportOptions{}

	var rootCmd = &cobra.Command{
		Use:   "column-exporter",
		Short: "Export binary column values from RDBMS to files",
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export binary column values to files",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(opts.output, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(opts)
		},
	}

	exportCmd.Flags().StringVar(&opts.driver, "driver", "mysql", "Database driver (mysql/postgres)")
	exportCmd.Flags().StringVar(&opts.host, "host", "localhost", "Database host")
	exportCmd.Flags().IntVar(&opts.port, "port", 3306, "Database port")
	exportCmd.Flags().StringVar(&opts.user, "user", "root", "Database user")
	exportCmd.Flags().StringVar(&opts.password, "password", "", "Database password")
	exportCmd.Flags().StringVar(&opts.database, "database", "", "Database name")
	exportCmd.Flags().StringVar(&opts.table, "table", "", "Table name")
	exportCmd.Flags().StringVar(&opts.column, "column", "", "Binary column to export (BLOB or BINARY types only)")
	exportCmd.Flags().StringVar(&opts.output, "output", "./output", "Output directory")
	exportCmd.Flags().IntVar(&opts.batchSize, "batch-size", 1000, "Number of rows to fetch per batch")
	exportCmd.Flags().StringVar(&opts.prefix, "prefix", "", "Prefix for output files (default: table_column)")
	exportCmd.Flags().StringVar(&opts.extension, "ext", ".bin", "File extension (default: .bin)")
	exportCmd.Flags().StringVar(&opts.nameCol, "name-column", "", "Column to use for output filenames (default: primary key)")
	exportCmd.Flags().StringVar(&opts.schema, "schema", "public", "Database schema (default: public, used in PostgreSQL)")

	exportCmd.MarkFlagRequired("database")
	exportCmd.MarkFlagRequired("table")
	exportCmd.MarkFlagRequired("column")

	rootCmd.AddCommand(exportCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runExport(opts *exportOptions) error {
	if opts.prefix == "" {
		opts.prefix = fmt.Sprintf("%s_%s", opts.table, opts.column)
	}

	dbConfig := db.Config{
		Driver:   opts.driver,
		Host:     opts.host,
		Port:     opts.port,
		User:     opts.user,
		Password: opts.password,
		Database: opts.database,
	}

	conn, err := db.Connect(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	exp := exporter.New(conn, opts.batchSize, opts.output, opts.prefix, opts.extension, opts.nameCol)
	if err := exp.Export(opts.table, opts.column); err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	}

	return nil
}
