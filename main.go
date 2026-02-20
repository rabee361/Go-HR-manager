package main

import (
	// "container/list"
	"fmt"
	"html/template"
	"log"
	"net/http"
	// "strconv"
)

type App struct {
	DepartmentRepository *SQLDepartmentRepository
	PositionRepository   *SQLPositionRepository
	EmployeeRepository   *SQLEmployeeRepository
	Templates            *template.Template
}

func main() {
	// 1. Initialize DB
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 2. Initialize App with Dependencies
	app := &App{
		DepartmentRepository: NewDepartmentRepository(db),
		PositionRepository:   NewPositionRepository(db),
		EmployeeRepository:   NewEmployeeRepository(db),
		Templates:            template.Must(template.ParseGlob("templates/dashboard/*.html")),
	}

	// 3. Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 4. Routes (using app methods as handlers)
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/departments", app.handleDepartments)
	http.HandleFunc("/positions", app.handlePositions)
	http.HandleFunc("/employees", app.handleEmployees)
	http.HandleFunc("/applications", app.handleApplications)

	fmt.Println("🚀 Server starting on :8080... 🌐")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{
		"ActivePage": "dashboard",
	}

	if err := app.Templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleDepartments(w http.ResponseWriter, r *http.Request) {
	departments, err := app.DepartmentRepository.GetDepartments()
	if err != nil {
		http.Error(w, "Failed to fetch departments", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"ActivePage":  "departments",
		"Departments": departments,
	}

	if err := app.Templates.ExecuteTemplate(w, "departments.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handlePositions(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"ActivePage": "positions",
	}
	if err := app.Templates.ExecuteTemplate(w, "positions.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleEmployees(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"ActivePage": "employees",
	}
	if err := app.Templates.ExecuteTemplate(w, "employees.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) handleApplications(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"ActivePage": "applications",
	}
	if err := app.Templates.ExecuteTemplate(w, "applications.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
