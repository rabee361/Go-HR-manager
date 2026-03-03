package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	devMode = true // Set to true for development
	clients = make(map[chan bool]bool)
	mu      sync.Mutex
)


type Reloader struct {
	watcher *fsnotify.Watcher
}

func (r *Reloader) Close() error {
	return r.watcher.Close()
}

func NewReloader() *Reloader {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	return &Reloader{
		watcher: watcher,
	}
}

type App struct {
	DepartmentRepository  DepartmentRepository
	PositionRepository    PositionRepository
	EmployeeRepository    EmployeeRepository
	ApplicationRepository ApplicationRepository
	LeaveRepository       LeaveRepository
	reloader				  Reloader
	Templates             map[string]*template.Template
}

func (app *App) AddPath(path string) error {
	return app.reloader.watcher.Add(path)
}

var app *App
	
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reloader := NewReloader()
	defer reloader.Close()

	db, err := connectToDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	app = &App{
		DepartmentRepository:  NewDepartmentRepository(db),
		PositionRepository:    NewPositionRepository(db),
		EmployeeRepository:    NewEmployeeRepository(db),
		ApplicationRepository: NewApplicationRepository(db),
		LeaveRepository:       NewLeaveRepository(db),
		reloader:             *reloader,
		Templates:             loadTemplates(),
	}

	if devMode {
		app.AddPath("templates/dashboard")
		app.AddPath("templates/partials")
		app.AddPath("static/css")

		go func() {
			for {
				select {
				case event, ok := <-app.reloader.watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						fmt.Printf("📝 File modified: %s. Notifying clients... 🔊\n", event.Name)
						mu.Lock()
						for client := range clients {
							client <- true
						}
						mu.Unlock()
					}
				case err, ok := <-app.reloader.watcher.Errors:
					if !ok {
						return
					}
					log.Println("Watcher error:", err)
				}
			}
		}()
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/dev-reload", app.handleDevReload)
	http.HandleFunc("/departments", app.handleDepartments)
	http.HandleFunc("/departments/export", app.handleExportDepartments)
	http.HandleFunc("/positions", app.handlePositions)
	http.HandleFunc("/positions/export", app.handleExportPositions)
	http.HandleFunc("/employees", app.handleEmployees)
	http.HandleFunc("/employees/export", app.handleExportEmployees)
	http.HandleFunc("/applications", app.handleApplications)
	http.HandleFunc("/applications/export", app.handleExportApplications)
	http.HandleFunc("/leaves", app.handleLeaves)
	http.HandleFunc("/leaves/export", app.handleExportLeaves)

	fmt.Println("🚀 Server starting on :8080... 🌐")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func loadTemplates() map[string]*template.Template {
	tmpls := make(map[string]*template.Template)
	baseFile := "templates/dashboard/base.html"

	partials, err := filepath.Glob("templates/partials/*.html")
	if err != nil {
		log.Fatalf("Error globbing partials: %v", err)
	}

	files, err := filepath.Glob("templates/dashboard/*.html")
	if err != nil {
		log.Fatalf("Error globbing templates: %v", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		if name == "base.html" {
			continue
		}
		allFiles := append([]string{baseFile, file}, partials...)
		tmpls[name] = template.Must(template.ParseFiles(allFiles...))
	}
	return tmpls
}

func (app *App) handleDevReload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan bool)
	mu.Lock()
	clients[messageChan] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, messageChan)
		mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-messageChan:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if devMode {
		app.Templates = loadTemplates()
	}

	data := map[string]any{
		"ActivePage": "dashboard",
	}

	if err := app.Templates["index.html"].Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleDepartments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	departments, err := app.DepartmentRepository.GetDepartments(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching departments: %v", err)
		http.Error(w, "Failed to fetch departments", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"ActivePage":  "departments",
		"Departments": departments,
	}

	if devMode {
		app.Templates = loadTemplates()
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := app.Templates["departments.html"].ExecuteTemplate(w, "departments_partial", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := app.Templates["departments.html"].Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handlePositions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	positions, err := app.PositionRepository.GetPositions(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching positions: %v", err)
		http.Error(w, "Failed to fetch positions", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"ActivePage": "positions",
		"Positions":  positions,
	}

	if devMode {
		app.Templates = loadTemplates()
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := app.Templates["positions.html"].ExecuteTemplate(w, "positions_partial", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := app.Templates["positions.html"].Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleEmployees(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	employees, err := app.EmployeeRepository.GetEmployees(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching employees: %v", err)
		http.Error(w, "Failed to fetch employees", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"ActivePage": "employees",
		"Employees":  employees,
	}

	if devMode {
		app.Templates = loadTemplates()
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := app.Templates["employees.html"].ExecuteTemplate(w, "employees_partial", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := app.Templates["employees.html"].Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleApplications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	applications, err := app.ApplicationRepository.GetApplications(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching applications: %v", err)
		http.Error(w, "Failed to fetch applications", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"ActivePage":   "applications",
		"Applications": applications,
	}

	if devMode {
		app.Templates = loadTemplates()
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := app.Templates["applications.html"].ExecuteTemplate(w, "applications_partial", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := app.Templates["applications.html"].Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleLeaves(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	leaves, err := app.LeaveRepository.GetLeaves(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching leaves: %v", err)
		http.Error(w, "Failed to fetch leaves", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"ActivePage": "leaves",
		"Leaves":     leaves,
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := app.Templates["leaves.html"].ExecuteTemplate(w, "leaves_partial", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := app.Templates["leaves.html"].Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleExportDepartments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	departments, err := app.DepartmentRepository.GetDepartments(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching departments for export: %v", err)
		http.Error(w, "Failed to fetch departments", http.StatusInternalServerError)
		return
	}

	headers := []string{"ID", "Name", "Description", "Created At"}
	mapper := func(d Department) []string {
		return []string{
			fmt.Sprintf("%d", d.ID),
			d.Name,
			d.Description,
			d.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	SetExcelHeaders(w, "Departments")
	if err := ExportToExcel(w, departments, headers, mapper); err != nil {
		log.Printf("Error exporting departments: %v", err)
	}
}

func (app *App) handleExportPositions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	positions, err := app.PositionRepository.GetPositions(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching positions for export: %v", err)
		http.Error(w, "Failed to fetch positions", http.StatusInternalServerError)
		return
	}

	headers := []string{"ID", "Name", "Description", "Created At"}
	mapper := func(p Position) []string {
		return []string{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			p.Description,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	SetExcelHeaders(w, "Positions")
	if err := ExportToExcel(w, positions, headers, mapper); err != nil {
		log.Printf("Error exporting positions: %v", err)
	}
}

func (app *App) handleExportEmployees(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	employees, err := app.EmployeeRepository.GetEmployees(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching employees for export: %v", err)
		http.Error(w, "Failed to fetch employees", http.StatusInternalServerError)
		return
	}

	headers := []string{"ID", "First Name", "Last Name", "Email", "Job Title", "Hire Date", "Salary", "Status", "Department ID", "Created At"}
	mapper := func(e Employee) []string {
		return []string{
			fmt.Sprintf("%d", e.ID),
			e.FirstName,
			e.LastName,
			e.Email,
			e.JobTitle,
			e.HireDate.Format("2006-01-02"),
			fmt.Sprintf("%.2f", e.Salary),
			e.Status,
			fmt.Sprintf("%d", e.DepartmentID),
			e.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	SetExcelHeaders(w, "Employees")
	if err := ExportToExcel(w, employees, headers, mapper); err != nil {
		log.Printf("Error exporting employees: %v", err)
	}
}

func (app *App) handleExportApplications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	applications, err := app.ApplicationRepository.GetApplications(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching applications for export: %v", err)
		http.Error(w, "Failed to fetch applications", http.StatusInternalServerError)
		return
	}

	headers := []string{"ID", "Name", "Email", "Phone", "Applied For", "Resume URL", "Status", "Created At"}
	mapper := func(a Application) []string {
		return []string{
			fmt.Sprintf("%d", a.ID),
			a.Name,
			a.Email,
			a.Phone,
			a.AppliedFor,
			a.ResumeURL,
			a.Status,
			a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	SetExcelHeaders(w, "Applications")
	if err := ExportToExcel(w, applications, headers, mapper); err != nil {
		log.Printf("Error exporting applications: %v", err)
	}
}

func (app *App) handleExportLeaves(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	leaves, err := app.LeaveRepository.GetLeaves(r.Context(), q)
	if err != nil {
		log.Printf("Error fetching leaves for export: %v", err)
		http.Error(w, "Failed to fetch leaves", http.StatusInternalServerError)
		return
	}

	headers := []string{"ID", "Employee ID", "Leave Type", "Start Date", "End Date", "Status", "Reason", "Created At"}
	mapper := func(l Leave) []string {
		return []string{
			fmt.Sprintf("%d", l.ID),
			fmt.Sprintf("%d", l.EmployeeID),
			l.LeaveType,
			l.StartDate.Format("2006-01-02"),
			l.EndDate.Format("2006-01-02"),
			l.Status,
			l.Reason,
			l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	SetExcelHeaders(w, "Leaves")
	if err := ExportToExcel(w, leaves, headers, mapper); err != nil {
		log.Printf("Error exporting leaves: %v", err)
	}
}
