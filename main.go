package main

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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
	reloader              Reloader
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
		reloader:              *reloader,
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
	http.HandleFunc("/departments/add", app.handleAddDepartments)
	http.HandleFunc("/departments/delete", app.handleDeleteDepartment)
	http.HandleFunc("/departments/update/{id}", app.handleUpdateDepartment)
	http.HandleFunc("/positions", app.handlePositions)
	http.HandleFunc("/positions/export", app.handleExportPositions)
	http.HandleFunc("/positions/add", app.handleAddPositions)
	http.HandleFunc("/positions/delete", app.handleDeletePosition)
	http.HandleFunc("/positions/update/{id}", app.handleUpdatePosition)
	http.HandleFunc("/employees", app.handleEmployees)
	http.HandleFunc("/employees/export", app.handleExportEmployees)
	http.HandleFunc("/employees/add", app.handleAddEmployees)
	http.HandleFunc("/employees/update/{id}", app.handleUpdateEmployee)
	http.HandleFunc("/employees/delete", app.handleDeleteEmployee)
	http.HandleFunc("/applications", app.handleApplications)
	http.HandleFunc("/applications/export", app.handleExportApplications)
	http.HandleFunc("/applications/add", app.handleAddApplications)
	http.HandleFunc("/applications/update/{id}", app.handleUpdateApplication)
	http.HandleFunc("/applications/delete", app.handleDeleteApplication)
	http.HandleFunc("/leaves", app.handleLeaves)
	http.HandleFunc("/leaves/export", app.handleExportLeaves)
	http.HandleFunc("/leaves/add", app.handleAddLeaves)
	http.HandleFunc("/leaves/update/{id}", app.handleUpdateLeave)
	http.HandleFunc("/leaves/delete", app.handleDeleteLeave)

	fmt.Println("🚀 Server starting on :8080... 🌐")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func loadTemplates() map[string]*template.Template {
	tmpls := make(map[string]*template.Template)
	baseFile := "templates/dashboard/base.html"
	files := []string{}

	partials, err := filepath.Glob("templates/partials/*.html")
	if err != nil {
		log.Fatalf("Error globbing partials: %v", err)
	}

	err = filepath.WalkDir("templates/dashboard/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Handle the error here to prevent panic and continue the walk (or stop by returning the error)
			fmt.Printf("preventing panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if d.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})
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

func (app *App) handleAddDepartments(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := app.Templates["add_department.html"].Execute(w, nil); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	name := r.FormValue("name")

	department := Department{Name: name, Description: description}
	err = app.DepartmentRepository.CreateDepartment(r.Context(), &department)
	if err != nil {
		log.Printf("Error adding department : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/departments")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleUpdateDepartment(w http.ResponseWriter, r *http.Request) {
	// 1. Get the ID from the URL
	idStr := r.PathValue("id") // Works in Go 1.22+
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// 2. Handle GET: Show the form with existing data
	if r.Method == http.MethodGet {
		dept, err := app.DepartmentRepository.GetDepartmentByID(r.Context(), id)
		if err != nil || dept == nil {
			http.Error(w, "Department not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Department": dept,
		}

		if err := app.Templates["update_department.html"].Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	// 3. Handle POST/PUT: Save the changes
	if err := r.ParseForm(); err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	department := Department{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	if err := app.DepartmentRepository.UpdateDepartment(r.Context(), &department); err != nil {
		log.Printf("Error updating department: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/departments")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleDeleteDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		fmt.Println("method is not delete")
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "can't parse id", http.StatusBadRequest)
		return
	}
	err = app.DepartmentRepository.DeleteDepartment(r.Context(), id)
	if err != nil {
		http.Error(w, "can't delete department", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/departments")
	w.WriteHeader(http.StatusSeeOther)
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

func (app *App) handleAddPositions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := app.Templates["add_position.html"].Execute(w, nil); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	name := r.FormValue("name")

	position := Position{Name: name, Description: description}
	err = app.PositionRepository.CreatePosition(r.Context(), &position)
	if err != nil {
		log.Printf("Error adding position : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/positions")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleUpdatePosition(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// 2. Handle GET: Show the form with existing data
	if r.Method == http.MethodGet {
		pos, err := app.PositionRepository.GetPositionByID(r.Context(), id)
		if err != nil || pos == nil {
			http.Error(w, "Position not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Position": pos,
		}

		if err := app.Templates["update_position.html"].Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	// 3. Handle POST/PUT: Save the changes
	if err := r.ParseForm(); err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	position := Position{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	if err := app.PositionRepository.UpdatePosition(r.Context(), &position); err != nil {
		log.Printf("Error updating position: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/positions")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleDeletePosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		return
	}

	err := r.ParseForm()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		return
	}
	app.PositionRepository.DeletePosition(r.Context(), id)
	w.Header().Set("HX-Redirect", "/positions")
	w.WriteHeader(http.StatusSeeOther)

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

func (app *App) handleAddEmployees(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := app.Templates["add_employee.html"].Execute(w, nil); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")

	employee := Employee{FirstName: firstName, LastName: lastName}
	err = app.EmployeeRepository.CreateEmployee(r.Context(), &employee)
	if err != nil {
		log.Printf("Error adding employee : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/employees")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleUpdateEmployee(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		emp, err := app.EmployeeRepository.GetEmployeeByID(r.Context(), id)
		if err != nil || emp == nil {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Emp": emp,
		}
		if err := app.Templates["update_employee.html"].Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	salary, _ := strconv.ParseFloat(r.FormValue("salary"), 64)
	deptID, _ := strconv.Atoi(r.FormValue("department_id"))
	hireDate, _ := time.Parse("2006-01-02", r.FormValue("hire_date"))

	employee := Employee{
		ID:           id,
		FirstName:    r.FormValue("first_name"),
		LastName:     r.FormValue("last_name"),
		Email:        r.FormValue("email"),
		JobTitle:     r.FormValue("job_title"),
		Salary:       salary,
		Status:       r.FormValue("status"),
		DepartmentID: deptID,
		HireDate:     hireDate,
	}

	err = app.EmployeeRepository.UpdateEmployee(r.Context(), &employee)
	if err != nil {
		log.Printf("Error updating employee : %v", err)
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/employees")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleDeleteEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		return
	}

	err := r.ParseForm()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		return
	}
	app.EmployeeRepository.DeleteEmployee(r.Context(), id)
	w.Header().Set("HX-Redirect", "/employees")
	w.WriteHeader(http.StatusSeeOther)

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

func (app *App) handleAddApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := app.Templates["add_application.html"].Execute(w, nil); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	appliedFor := r.FormValue("applied_for")
	resumeURL := r.FormValue("resume_url")
	status := r.FormValue("status")

	application := Application{Name: name, Email: email, Phone: phone, AppliedFor: appliedFor, ResumeURL: resumeURL, Status: status}
	err = app.ApplicationRepository.CreateApplication(r.Context(), &application)
	if err != nil {
		log.Printf("Error adding application : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/applications")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleDeleteApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))

	err = app.ApplicationRepository.DeleteApplication(r.Context(), id)
	if err != nil {
		log.Printf("Error deleting application : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/applications")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleUpdateApplication(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		appData, err := app.ApplicationRepository.GetApplicationByID(r.Context(), id)
		if err != nil || appData == nil {
			http.Error(w, "Application not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Application": appData,
		}
		if err := app.Templates["update_application.html"].Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	application := Application{
		ID:         id,
		Name:       r.FormValue("name"),
		Email:      r.FormValue("email"),
		Phone:      r.FormValue("phone"),
		AppliedFor: r.FormValue("applied_for"),
		ResumeURL:  r.FormValue("resume_url"),
		Status:     r.FormValue("status"),
	}

	err = app.ApplicationRepository.UpdateApplication(r.Context(), &application)
	if err != nil {
		log.Printf("Error updating application : %v", err)
		http.Error(w, "Failed to update application", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/applications")
	w.WriteHeader(http.StatusSeeOther)
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

func (app *App) handleAddLeaves(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := app.Templates["add_leave.html"].Execute(w, nil); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	employeeID, err := strconv.Atoi(r.FormValue("employee_id"))
	leaveType := r.FormValue("leave_type")
	startDate, err := time.Parse("2006-01-02", r.FormValue("start_date"))
	endDate, err := time.Parse("2006-01-02", r.FormValue("end_date"))
	status := r.FormValue("status")
	reason := r.FormValue("reason")

	leave := Leave{EmployeeID: employeeID, LeaveType: leaveType, StartDate: startDate, EndDate: endDate, Status: status, Reason: reason}
	err = app.LeaveRepository.CreateLeave(r.Context(), &leave)
	if err != nil {
		log.Printf("Error adding leave : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/leaves")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleDeleteLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))

	err = app.LeaveRepository.DeleteLeave(r.Context(), id)
	if err != nil {
		log.Printf("Error deleting leave : %v", err)
		return
	}
	w.Header().Set("HX-Redirect", "/leaves")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleUpdateLeave(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		leave, err := app.LeaveRepository.GetLeaveByID(r.Context(), id)
		if err != nil || leave == nil {
			http.Error(w, "Leave not found", http.StatusNotFound)
			return
		}

		data := map[string]any{
			"Leave": leave,
		}
		if err := app.Templates["update_leave.html"].Execute(w, data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "can't parse form", http.StatusBadRequest)
		return
	}

	employeeID, _ := strconv.Atoi(r.FormValue("employee_id"))
	startDate, _ := time.Parse("2006-01-02", r.FormValue("start_date"))
	endDate, _ := time.Parse("2006-01-02", r.FormValue("end_date"))

	leave := Leave{
		ID:         id,
		EmployeeID: employeeID,
		LeaveType:  r.FormValue("leave_type"),
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     r.FormValue("status"),
		Reason:     r.FormValue("reason"),
	}

	err = app.LeaveRepository.UpdateLeave(r.Context(), &leave)
	if err != nil {
		log.Printf("Error updating leave : %v", err)
		http.Error(w, "Failed to update leave", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/leaves")
	w.WriteHeader(http.StatusSeeOther)
}
