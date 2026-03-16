package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//go:embed templates/index.html
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/index.html"))

type fileEntry struct {
	Name    string
	Size    string
	ModTime string
}

type pageData struct {
	Config  Config
	Files   []fileEntry
	Message string
}

// Serve starts the HTTP server on addr (e.g. ":8080"). Blocks until error.
func serveWeb(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("POST /config", handleSaveConfig)
	log.Printf("web: listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("saved")
	pd := buildPageData(msg)
	if err := tmpl.Execute(w, pd); err != nil {
		log.Printf("web: template error: %v", err)
	}
}

func handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	names := r.Form["fields_name"]
	types := r.Form["fields_type"]
	offsets := r.Form["fields_offset"]
	bitpos := r.Form["fields_bitpos"]

	n := len(names)
	if len(types) < n {
		n = len(types)
	}
	if len(offsets) < n {
		n = len(offsets)
	}

	fields := make([]FieldDef, 0, n)
	for i := 0; i < n; i++ {
		name := names[i]
		if name == "" {
			continue
		}
		off, err := strconv.Atoi(offsets[i])
		if err != nil {
			off = 0
		}
		bp := 0
		if i < len(bitpos) {
			bp, err = strconv.Atoi(bitpos[i])
			if err != nil || bp < 0 || bp > 7 {
				bp = 0
			}
		}
		fields = append(fields, FieldDef{
			Name:   name,
			Type:   types[i],
			Offset: off,
			BitPos: bp,
		})
	}

	maxSize, err := strconv.ParseInt(r.FormValue("max_file_size"), 10, 64)
	if err != nil || maxSize <= 0 {
		maxSize = 100 * 1024 * 1024
	}

	outputDir := r.FormValue("output_dir")
	if outputDir == "" {
		outputDir = "data"
	}

	cfg := Config{
		Fields:           fields,
		MaxFileSizeBytes: maxSize,
		OutputDir:        outputDir,
	}

	if err := SaveConfig(cfg); err != nil {
		http.Error(w, fmt.Sprintf("save error: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?saved=1", http.StatusSeeOther)
}

func buildPageData(savedMsg string) pageData {
	cfg := GetConfig()
	pd := pageData{Config: cfg}
	if savedMsg != "" {
		pd.Message = "Configuration saved."
	}
	entries, err := os.ReadDir(cfg.OutputDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".parquet" {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			pd.Files = append(pd.Files, fileEntry{
				Name:    e.Name(),
				Size:    fmtSize(info.Size()),
				ModTime: info.ModTime().Format(time.DateTime),
			})
		}
	}
	return pd
}

func fmtSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
