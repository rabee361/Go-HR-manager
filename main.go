package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type App struct {
	DepartmentRepository  DepartmentRepository
	PositionRepository    PositionRepository
	EmployeeRepository    EmployeeRepository
	ApplicationRepository ApplicationRepository
	LeaveRepository       LeaveRepository
	Templates             map[string]*template.Template
}

func main() {
	// 1. Create a base context with a timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Initialize DB
	db, err := connectToDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Initialize App with Dependencies
	app := &App{
		DepartmentRepository:  NewDepartmentRepository(db),
		PositionRepository:    NewPositionRepository(db),
		EmployeeRepository:    NewEmployeeRepository(db),
		ApplicationRepository: NewApplicationRepository(db),
		LeaveRepository:       NewLeaveRepository(db),
		Templates:             loadTemplates(),
	}

	// 4. Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 5. Routes (using app methods as handlers)
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/departments", app.handleDepartments)
	http.HandleFunc("/positions", app.handlePositions)
	http.HandleFunc("/employees", app.handleEmployees)
	http.HandleFunc("/applications", app.handleApplications)
	http.HandleFunc("/leaves", app.handleLeaves)

	fmt.Println("🚀 Server starting on :8080... 🌐")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func loadTemplates() map[string]*template.Template {
	tmpls := make(map[string]*template.Template)
	baseFile := "templates/dashboard/base.html"

	// Glob all partials
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
		// Combine base, partials, and the page file
		allFiles := append([]string{baseFile, file}, partials...)
		tmpls[name] = template.Must(template.ParseFiles(allFiles...))
	}
	return tmpls
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
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
