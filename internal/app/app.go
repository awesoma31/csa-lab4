package app

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/awesoma31/csa-lab4/config"
	"github.com/awesoma31/csa-lab4/internal/middleware"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/machine/logger"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
	"github.com/sanity-io/litter"
	"gopkg.in/yaml.v2"
)

const (
	templateDir = "web/templates"
	staticDir   = "web/static"
)

var tpl = template.Must(template.ParseGlob(filepath.Join(templateDir, "*.html")))

func Run(cfg *config.Config) {
	router := http.NewServeMux()
	router.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(staticDir))))

	loadRoutes(router)

	stack := middleware.CreateStack(
		middleware.Logging,
	)

	s := http.Server{
		Addr:    cfg.Port,
		Handler: stack(router),
	}

	log.Println("Starting server on ", cfg.Port)
	log.Fatal(s.ListenAndServe())
}

func loadRoutes(router *http.ServeMux) {
	router.HandleFunc("/", handleHome)

	router.HandleFunc("POST /api/simulate", makeHTTPHandler(handleSimulate))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "indexTempl", nil)
	if err != nil {
		http.Error(w, "error parsing index template html", http.StatusInternalServerError)
	}
}

func handleSimulate(w http.ResponseWriter, r *http.Request) error {
	//FIX: duplicate tick logs in cpu

	if err := r.ParseForm(); err != nil {
		return apiError{Err: "Couldn't parse request form for simulation", Status: http.StatusBadRequest}
	}

	src := r.Form.Get("src")
	if strings.TrimSpace(src) == "" {
		renderResult(w, SimulateResponse{Output: "Nothing to simulate!"})
		return nil
	}
	ast, pErr := parser.Parse(string(src))
	if len(pErr) != 0 {
		return apiError{Err: "Parsing Error", Status: http.StatusBadRequest, Errors: pErr}
	}

	cg := codegen.NewCodeGenerator()
	memI, memD, dbgAsm, cgErr := cg.Generate(ast)
	if len(cgErr) != 0 {
		return apiError{Err: "code generation error", Status: http.StatusBadRequest, Errors: cgErr}
	}

	if strings.TrimSpace(r.Form.Get("cfg")) == "" {
		renderResult(w, SimulateResponse{Output: "Simulation config is empty!"})
		return nil
	}
	cfgBytes := []byte(r.Form.Get("cfg"))
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

	cpuCFG.MemD = memD
	cpuCFG.MemI = memI
	cpuCFG.IOC = ioc
	cpuCFG.Logger = lg

	cpu := machine.New(cpuCFG)
	output := cpu.Run()

	resp := SimulateResponse{
		Output:   output,
		Ast:      litter.Sdump(ast),
		MemI:     formatMemI(memI),
		MemD:     formatMemD(memD),
		DebugAsm: formatDebugAsm(dbgAsm),
	}

	renderResult(w, resp)

	return nil
}

func renderResult(w http.ResponseWriter, resp SimulateResponse) {
	w.Header().Set("Content-Type", "text/html")
	_ = tpl.ExecuteTemplate(w, "resultTemplate", resp) // ⬅️ имя из {{define "result"}}
}
