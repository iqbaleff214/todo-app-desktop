package repository

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const bakFilename = "data.db.bak"

// MaybeBackup copies data.db to data.db.bak when the database has changed
// since the last backup. It is intended to be called once at startup, before
// any DB writes. Failures are logged but never returned — backup errors must
// not block the app from starting.
func MaybeBackup(dataDir string) error {
	dbPath  := filepath.Join(dataDir, dbFilename)
	bakPath := filepath.Join(dataDir, bakFilename)

	// 1. Nothing to back up if data.db does not exist yet.
	dbInfo, err := os.Stat(dbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		err = fmt.Errorf("backup: stat %s: %w", dbFilename, err)
		log.Print(err)
		return err
	}

	// 2. No backup yet — create one unconditionally.
	bakInfo, err := os.Stat(bakPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if copyErr := copyFile(dbPath, bakPath); copyErr != nil {
				copyErr = fmt.Errorf("backup: initial copy: %w", copyErr)
				log.Print(copyErr)
				return copyErr
			}
			return nil
		}
		err = fmt.Errorf("backup: stat %s: %w", bakFilename, err)
		log.Print(err)
		return err
	}

	// 3-4. Overwrite backup only when data.db is strictly newer.
	if dbInfo.ModTime().After(bakInfo.ModTime()) {
		if copyErr := copyFile(dbPath, bakPath); copyErr != nil {
			copyErr = fmt.Errorf("backup: copy: %w", copyErr)
			log.Print(copyErr)
			return copyErr
		}
	}
	// 5. data.db not newer — nothing to do.
	return nil
}

// copyFile copies src to dst using kernel I/O — no exec, no shell, cross-platform.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	// Flush to disk before returning so the backup is durable.
	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dst, err)
	}
	return nil
}
