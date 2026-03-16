package web

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

	"basicHistorian/internal/config"
)

//go:embed ../../templates/index.html
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/index.html"))

type fileEntry struct {
	Name    string
	Size    string
	ModTime string
}

type pageData struct {
	Config  config.Config
	Files   []fileEntry
	Message string
}

// Serve starts the HTTP server on addr (e.g. ":8080"). Blocks until error.
func Serve(addr string) error {
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

	n := len(names)
	if len(types) < n {
		n = len(types)
	}
	if len(offsets) < n {
		n = len(offsets)
	}

	fields := make([]config.FieldDef, 0, n)
	for i := 0; i < n; i++ {
		name := names[i]
		if name == "" {
			continue
		}
		off, err := strconv.Atoi(offsets[i])
		if err != nil {
			off = 0
		}
		fields = append(fields, config.FieldDef{
			Name:   name,
			Type:   types[i],
			Offset: off,
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

	cfg := config.Config{
		Fields:           fields,
		MaxFileSizeBytes: maxSize,
		OutputDir:        outputDir,
	}

	if err := config.Save(cfg); err != nil {
		http.Error(w, fmt.Sprintf("save error: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?saved=1", http.StatusSeeOther)
}

func buildPageData(savedMsg string) pageData {
	cfg := config.Get()
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







































































































































































































































































</html>`</body></script>}  tbody.appendChild(tr);    '<td><button type="button" class="btn" onclick="this.closest(\'tr\').remove()">Remove</button></td>';    '<td><input type="number" name="fields_offset" value="0" min="0" max="499"></td>' +    '</select></td>' +      '<option value="bool">bool  (1 byte)</option>' +      '<option value="real">real  (4 bytes, alias float32)</option>' +      '<option value="float32">float32 (4 bytes)</option>' +      '<option value="dint" selected>dint  (4 bytes, signed)</option>' +      '<option value="int">int   (2 bytes, signed)</option>' +      '<option value="sint">sint  (1 byte, signed)</option>' +    '<td><select name="fields_type">' +    '<td><input type="text" name="fields_name" value=""></td>' +  tr.innerHTML =  const tr = document.createElement('tr');  const tbody = document.getElementById('fields-body');function addFieldRow() {<script>{{end}}<p>No parquet files found in <code>{{.Config.OutputDir}}</code>.</p>{{else}}</table>  </tbody>    {{end}}    <tr><td>{{.Name}}</td><td>{{.Size}}</td><td>{{.ModTime}}</td></tr>    {{range .Files}}  <tbody>  <thead><tr><th>File</th><th>Size</th><th>Last Modified</th></tr></thead><table class="files">{{if .Files}}<h2>Data Files</h2></form>  <button type="submit" class="save-btn">Save Configuration</button>  <br>  <button type="button" class="btn" onclick="addFieldRow()">+ Add Field</button>  <br>  </table>    </tbody>      {{end}}      </tr>        <td><button type="button" class="btn" onclick="this.closest('tr').remove()">Remove</button></td>        <td><input type="number" name="fields_offset" value="{{.Offset}}" min="0" max="499"></td>        </td>          </select>            <option value="bool"  {{if eq .Type "bool"  }}selected{{end}}>bool  (1 byte)</option>            <option value="real"  {{if eq .Type "real"  }}selected{{end}}>real  (4 bytes, alias float32)</option>            <option value="float32" {{if eq .Type "float32"}}selected{{end}}>float32 (4 bytes)</option>            <option value="dint"  {{if eq .Type "dint"  }}selected{{end}}>dint  (4 bytes, signed)</option>            <option value="int"   {{if eq .Type "int"   }}selected{{end}}>int   (2 bytes, signed)</option>            <option value="sint"  {{if eq .Type "sint"  }}selected{{end}}>sint  (1 byte, signed)</option>            {{range $_, $t := $.Config.Fields}}{{end}}            {{$cur := .Type}}          <select name="fields_type">        <td>        <td><input type="text" name="fields_name" value="{{.Name}}"></td>      <tr>      {{range .Config.Fields}}    <tbody id="fields-body">    </thead>      <tr><th>Name</th><th>Type</th><th>Offset (bytes)</th><th></th></tr>    <thead>  <table class="fields">  <p>Map named fields onto the 500-byte output buffer. Offsets are zero-based byte positions.</p>  <h2>Field Definitions</h2>  </p>           value="{{.Config.MaxFileSizeBytes}}" min="1">    <input type="number" id="max_file_size" name="max_file_size"    <label for="max_file_size">Max File Size (bytes):</label>  <p>  </p>    <input type="text" id="output_dir" name="output_dir" value="{{.Config.OutputDir}}">    <label for="output_dir">Output Directory:</label>  <p>  <h2>Historian Settings</h2>  </div>    <strong>Output (from PLC):</strong> 500 bytes    <strong>Input (to PLC):</strong> 4 bytes &nbsp;&nbsp;&nbsp;  <div class="info-box">  <h2>IO Connection</h2><form method="POST" action="/config">{{if .Message}}<p class="msg">&#10003; {{.Message}}</p>{{end}}<h1>BasicHistorian2</h1><body></head>  </style>    .msg  { color: #2a7a2a; font-weight: bold; margin-top: 10px; }    .save-btn { padding: 8px 28px; font-size: 1.05em; margin-top: 16px; cursor: pointer; }    .btn  { cursor: pointer; padding: 4px 10px; }    table.files  th { background: #f0f0f0; text-align: left; }    table.files  th, table.files  td { border: 1px solid #ccc; padding: 6px 10px; }    table.files  { border-collapse: collapse; width: 100%; margin-top: 8px; }    table.fields th { background: #f0f0f0; text-align: left; }    table.fields th, table.fields td { border: 1px solid #ccc; padding: 6px 10px; }    table.fields { border-collapse: collapse; width: 100%; margin-top: 8px; }    .info-box { background: #f4f4f4; border: 1px solid #ddd; padding: 10px 14px; border-radius: 4px; margin: 8px 0; }    select { padding: 4px 6px; }    input[type=text], input[type=number] { padding: 4px 6px; width: 260px; box-sizing: border-box; }    label { display: inline-block; min-width: 200px; }    p { margin: 8px 0; }    h2 { border-bottom: 1px solid #ccc; padding-bottom: 4px; margin-top: 28px; }    h1 { border-bottom: 2px solid #333; padding-bottom: 8px; }    body { font-family: sans-serif; max-width: 960px; margin: 40px auto; padding: 0 20px; }  <style>  <title>BasicHistorian2</title>  <meta charset="UTF-8"><head><html lang="en">const htmlTemplate = `<!DOCTYPE html>}	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])	}		exp++		div *= unit	for n := b / unit; n >= unit; n /= unit {	div, exp := int64(unit), 0	}		return fmt.Sprintf("%d B", b)	if b < unit {	const unit = 1024func fmtSize(b int64) string {}	return pd	}		}			})				ModTime: info.ModTime().Format(time.DateTime),				Size:    fmtSize(info.Size()),				Name:    e.Name(),			pd.Files = append(pd.Files, fileEntry{			}				continue			if err != nil {			info, err := e.Info()			}				continue			if e.IsDir() || filepath.Ext(e.Name()) != ".parquet" {		for _, e := range entries {	if err == nil {	entries, err := os.ReadDir(cfg.OutputDir)	}		pd.Message = "Configuration saved."	if savedMsg != "" {	pd := pageData{Config: cfg}	cfg := config.Get()func buildPageData(savedMsg string) pageData {}	http.Redirect(w, r, "/?saved=1", http.StatusSeeOther)	}		return		http.Error(w, fmt.Sprintf("save error: %v", err), http.StatusInternalServerError)	if err := config.Save(cfg); err != nil {	}		OutputDir:        outputDir,		MaxFileSizeBytes: maxSize,		Fields:           fields,	cfg := config.Config{	}		outputDir = "data"	if outputDir == "" {	outputDir := r.FormValue("output_dir")	}		maxSize = 100 * 1024 * 1024	if err != nil || maxSize <= 0 {	maxSize, err := strconv.ParseInt(r.FormValue("max_file_size"), 10, 64)	}		})			Offset: off,			Type:   types[i],			Name:   name,		fields = append(fields, config.FieldDef{		}			off = 0		if err != nil {		off, err := strconv.Atoi(offsets[i])		}			continue		if name == "" {		name := names[i]	for i := 0; i < n; i++ {	fields := make([]config.FieldDef, 0, n)	}		n = len(offsets)	if len(offsets) < n {	}		n = len(types)	if len(types) < n {	n := len(names)	offsets := r.Form["fields_offset"]	types := r.Form["fields_type"]	names := r.Form["fields_name"]	}		return		http.Error(w, "bad form", http.StatusBadRequest)	if err := r.ParseForm(); err != nil {func handleSaveConfig(w http.ResponseWriter, r *http.Request) {}	}		log.Printf("web: template error: %v", err)	if err := tmpl.Execute(w, pd); err != nil {	pd := buildPageData(msg)	msg := r.URL.Query().Get("saved")func handleIndex(w http.ResponseWriter, r *http.Request) {}	return http.ListenAndServe(addr, mux)	log.Printf("web: listening on %s", addr)	mux.HandleFunc("POST /config", handleSaveConfig)	mux.HandleFunc("GET /", handleIndex)	mux := http.NewServeMux()func Serve(addr string) error {// Serve starts the HTTP server on addr (e.g. ":8080"). Blocks until error.var tmpl = template.Must(template.New("main").Parse(htmlTemplate))}	Message string	Files   []fileEntry	Config  config.Configtype pageData struct {}	ModTime string	Size    string	Name    stringtype fileEntry struct {var supportedTypes = []string{"sint", "int", "dint", "float32", "real", "bool"})	"basicHistorian/internal/config"	"time"	"strconv"	"path/filepath"	"os"	"net/http"	"log"	"html/template"	"fmt"import (package web