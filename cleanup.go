package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
)

// cleanupDataDir removes the oldest parquet files in cfg.OutputDir when either
// the total size exceeds cfg.MaxTotalSizeBytes or the count exceeds cfg.MaxFileCount.
// Limits set to 0 are skipped.
func cleanupDataDir(cfg Config) {
	if cfg.MaxTotalSizeBytes <= 0 && cfg.MaxFileCount <= 0 {
		return
	}

	entries, err := os.ReadDir(cfg.OutputDir)
	if err != nil {
		log.Printf("cleanup: readdir %q: %v", cfg.OutputDir, err)
		return
	}

	type fileInfo struct {
		path    string
		size    int64
		modTime int64 // unix nano for sorting
	}

	var files []fileInfo
	var totalSize int64
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".parquet" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path:    filepath.Join(cfg.OutputDir, e.Name()),
			size:    info.Size(),
			modTime: info.ModTime().UnixNano(),
		})
		totalSize += info.Size()
	}

	// Sort oldest first so we delete from the front.
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime < files[j].modTime
	})

	remove := func(f fileInfo) {
		if err := os.Remove(f.path); err != nil {
			log.Printf("cleanup: remove %q: %v", f.path, err)
			return
		}
		log.Printf("cleanup: removed %s (%s)", f.path, fmtSize(f.size))
		totalSize -= f.size
	}

	// Enforce max file count.
	for cfg.MaxFileCount > 0 && len(files) > cfg.MaxFileCount {
		remove(files[0])
		files = files[1:]
	}

	// Enforce max total size.
	for cfg.MaxTotalSizeBytes > 0 && totalSize > cfg.MaxTotalSizeBytes && len(files) > 0 {
		remove(files[0])
		files = files[1:]
	}
}
