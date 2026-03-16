package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	parquet "github.com/parquet-go/parquet-go"
)

// Run reads raw 496-byte IO output frames from ch and writes them to rotating
// parquet files. It does not return until ch is closed.
func runHistorian(ch <-chan [496]byte) {
	for {
		cfg := GetConfig()
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			log.Printf("historian: mkdir %q: %v - retrying in 5s", cfg.OutputDir, err)
			time.Sleep(5 * time.Second)
			continue
		}
		done := writeFile(ch, cfg)
		cleanupDataDir(cfg)
		if done {
			return
		}
	}
}

// writeFile writes rows to one parquet file until the file reaches the size
// limit, a config-change signal arrives, or ch is closed.
// Returns true only when ch is closed (program shutting down).
func writeFile(ch <-chan [496]byte, cfg Config) (done bool) {
	schema, colMap, err := buildSchema(cfg)
	if err != nil {
		log.Printf("historian: build schema: %v - sleeping 5s", err)
		time.Sleep(5 * time.Second)
		return false
	}

	fname := filepath.Join(cfg.OutputDir,
		fmt.Sprintf("data_%s.parquet", time.Now().Format("20060102_150405")))
	f, err := os.Create(fname)
	if err != nil {
		log.Printf("historian: create %q: %v - sleeping 1s", fname, err)
		time.Sleep(time.Second)
		return false
	}
	defer f.Close()

	w := parquet.NewWriter(f, schema)
	defer func() {
		if cerr := w.Close(); cerr != nil {
			log.Printf("historian: close writer: %v", cerr)
		}
	}()

	log.Printf("historian: writing to %s", fname)
	numCols := len(schema.Columns())

	// Set up optional time-based rotation ticker.
	var rotateCh <-chan time.Time
	if cfg.RotateIntervalHours > 0 {
		ticker := time.NewTicker(time.Duration(cfg.RotateIntervalHours * float64(time.Hour)))
		defer ticker.Stop()
		rotateCh = ticker.C
	}

	for {
		select {
		case <-ChangeCh:
			log.Printf("historian: config changed - rotating file")
			return false

		case <-RotateCh:
			log.Printf("historian: manual rotation requested")
			return false

		case <-rotateCh:
			log.Printf("historian: time-based rotation after %.4g hours", cfg.RotateIntervalHours)
			return false

		case raw, ok := <-ch:
			if !ok {
				return true
			}
			row := makeRow(raw[:], cfg, colMap, numCols)
			if _, err := w.WriteRows([]parquet.Row{row}); err != nil {
				log.Printf("historian: write row: %v", err)
			}
			if info, err := f.Stat(); err == nil && info.Size() >= cfg.MaxFileSizeBytes {
				log.Printf("historian: rotating file at %d bytes", info.Size())
				return false
			}
		}
	}
}

func buildSchema(cfg Config) (*parquet.Schema, map[string]int, error) {
	group := parquet.Group{
		"timestamp": parquet.Int(64),
	}
	for _, f := range cfg.Fields {
		node, err := nodeForType(f.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("field %q: %w", f.Name, err)
		}
		group[f.Name] = node
	}
	schema := parquet.NewSchema("io_row", group)

	colMap := make(map[string]int)
	for _, path := range schema.Columns() {
		lc, ok := schema.Lookup(path...)
		if ok {
			colMap[path[len(path)-1]] = lc.ColumnIndex
		}
	}
	return schema, colMap, nil
}

func nodeForType(t string) (parquet.Node, error) {
	switch t {
	case "sint":
		return parquet.Int(8), nil
	case "int":
		return parquet.Int(16), nil
	case "dint":
		return parquet.Int(32), nil
	case "float32", "real":
		return parquet.Leaf(parquet.FloatType), nil
	case "bool":
		return parquet.Leaf(parquet.BooleanType), nil
	default:
		return nil, fmt.Errorf("unknown type %q", t)
	}
}

func makeRow(data []byte, cfg Config, colMap map[string]int, numCols int) parquet.Row {
	row := make(parquet.Row, numCols)

	tsIdx := colMap["timestamp"]
	row[tsIdx] = parquet.Int64Value(time.Now().UnixNano()).Level(0, 0, tsIdx)

	for _, f := range cfg.Fields {
		idx, ok := colMap[f.Name]
		if !ok {
			continue
		}
		row[idx] = encodeValue(data, f).Level(0, 0, idx)
	}
	return row
}

func encodeValue(data []byte, f FieldDef) parquet.Value {
	off := f.Offset
	switch f.Type {
	case "sint":
		if off < len(data) {
			return parquet.Int32Value(int32(int8(data[off])))
		}
	case "int":
		if off+1 < len(data) {
			return parquet.Int32Value(int32(int16(binary.LittleEndian.Uint16(data[off:]))))
		}
	case "dint":
		if off+3 < len(data) {
			return parquet.Int32Value(int32(binary.LittleEndian.Uint32(data[off:])))
		}
	case "float32", "real":
		if off+3 < len(data) {
			return parquet.FloatValue(math.Float32frombits(binary.LittleEndian.Uint32(data[off:])))
		}
	case "bool":
		if off < len(data) {
			return parquet.BooleanValue((data[off]>>f.BitPos)&1 == 1)
		}
	}
	return parquet.NullValue()
}
