package app

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/awesoma31/csa-lab4/config"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/machine/logger"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
	"gopkg.in/yaml.v2"
)

const (
	templateDir = "web/templates"
	staticDir   = "web/static"
)

var tpl = template.Must(template.ParseGlob(filepath.Join(templateDir, "*.html")))

func makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(apiError); ok {
				_ = writeJson(w, e.Status, e)
				return
			}
			_ = writeJson(w, http.StatusInternalServerError, apiError{Err: "internal server", Status: http.StatusInternalServerError})
		}
	}
}

func Run(cfg *config.Config) {
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(staticDir))))

	http.HandleFunc("/", handleHome)

	http.HandleFunc("/api/simulate", makeHTTPHandler(handleSimulate))

	log.Println("Starting server on ", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "error parsing index template html", http.StatusInternalServerError)
	}
}

func handleSimulate(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return apiError{Err: "only POST is allowed", Status: http.StatusMethodNotAllowed}
	}

	var req SimulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apiError{Err: "Couldn't decode json data for simulation", Status: http.StatusBadRequest}
	}

	src := req.Code
	ast, pErr := parser.Parse(string(src))
	if len(pErr) != 0 {
		return apiError{Err: "Parsing Error", Status: http.StatusBadRequest, Errors: pErr}
	}

	cg := codegen.NewCodeGenerator()
	imem, dmem, dbgAsm, cgErr := cg.Generate(ast)
	if len(cgErr) != 0 {
		return apiError{Err: "code generation error", Status: http.StatusBadRequest, Errors: cgErr}
	}

	//TODO: to bin, run sim

	cfgBytes := []byte(req.Config)
	var cpuCFG *machine.CpuConfig
	if err := yaml.Unmarshal(cfgBytes, &cpuCFG); err != nil {
		return apiError{
			Err:    "invalid yaml config",
			Status: http.StatusBadRequest,
			Errors: []string{err.Error()},
		}
	}

	ioc := io.NewIOController(cpuCFG.Schedule)
	//TODO: just get the logs
	lg := logger.New(true, cpuCFG.LogFilePath)

	cpuCFG.MemD = dmem
	cpuCFG.MemI = imem
	cpuCFG.IOC = ioc
	cpuCFG.Logger = lg

	cpu := machine.New(cpuCFG)
	output := cpu.Run()

	//TODO: memD into readable view
	resp := SimulateResponse{
		Message:    "success",
		StatusCode: http.StatusOK,
		Output:     output,
		MemI:       imem,
		MemD:       dmem,
		DebugAsm:   dbgAsm,
	}

	return writeJson(w, http.StatusOK, resp)
}
