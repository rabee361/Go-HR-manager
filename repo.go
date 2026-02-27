package main

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLDepartmentRepository struct {
	db *sql.DB
}

type SQLPositionRepository struct {
	db *sql.DB
}

type SQLEmployeeRepository struct {
	db *sql.DB
}

type SQLLeaveRepository struct {
	db *sql.DB
}

type SQLApplicationRepository struct {
	db *sql.DB
}

func NewDepartmentRepository(db *sql.DB) *SQLDepartmentRepository {
	return &SQLDepartmentRepository{db: db}
}

func NewPositionRepository(db *sql.DB) *SQLPositionRepository {
	return &SQLPositionRepository{db: db}
}

func NewEmployeeRepository(db *sql.DB) *SQLEmployeeRepository {
	return &SQLEmployeeRepository{db: db}
}

func NewLeaveRepository(db *sql.DB) *SQLLeaveRepository {
	return &SQLLeaveRepository{db: db}
}

func NewApplicationRepository(db *sql.DB) *SQLApplicationRepository {
	return &SQLApplicationRepository{db: db}
}

func (r *SQLDepartmentRepository) GetDepartments(ctx context.Context, q string) ([]Department, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM departments WHERE name LIKE ?;", "%"+q+"%;")
	if err != nil {
		return nil, fmt.Errorf("querying departments: %w", err)
	}
	defer rows.Close()
	var departments []Department

	for rows.Next() {
		var department Department
		if err := rows.Scan(&department.ID, &department.Name, &department.Description, &department.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning department: %w", err)
		}
		departments = append(departments, department)
	}
	return departments, nil
}

func (r *SQLDepartmentRepository) GetDepartmentByID(ctx context.Context, id int) (*Department, error) {
	var department Department
	err := r.db.QueryRowContext(ctx, "SELECT id, name, description, created_at FROM departments WHERE id = ?;", id).Scan(&department.ID, &department.Name, &department.Description, &department.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying department by id: %w", err)
	}
	return &department, nil
}

func (r *SQLDepartmentRepository) CreateDepartment(ctx context.Context, department *Department) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO departments (name, description) VALUES (?, ?);", department.Name, department.Description)
	if err != nil {
		return fmt.Errorf("creating department: %w", err)
	}
	return nil
}

func (r *SQLDepartmentRepository) UpdateDepartment(ctx context.Context, department *Department) error {
	_, err := r.db.ExecContext(ctx, "UPDATE departments SET name = ?, description = ? WHERE id = ?;", department.Name, department.Description, department.ID)
	if err != nil {
		return fmt.Errorf("updating department: %w", err)
	}
	return nil
}

func (r *SQLDepartmentRepository) DeleteDepartment(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM departments WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("deleting department: %w", err)
	}
	return nil
}

func (r *SQLPositionRepository) GetPositions(ctx context.Context, q string) ([]Position, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM positions WHERE name LIKE ?;", "%"+q+"%")
	if err != nil {
		return nil, fmt.Errorf("querying positions: %w", err)
	}
	defer rows.Close()
	var positions []Position

	for rows.Next() {
		var position Position
		if err := rows.Scan(&position.ID, &position.Name, &position.Description, &position.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning position: %w", err)
		}
		positions = append(positions, position)
	}
	return positions, nil
}

func (r *SQLPositionRepository) GetPositionByID(ctx context.Context, id int) (*Position, error) {
	var position Position
	err := r.db.QueryRowContext(ctx, "SELECT id, name, description, created_at FROM positions WHERE id = ?;", id).Scan(&position.ID, &position.Name, &position.Description, &position.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying position by id: %w", err)
	}
	return &position, nil
}

func (r *SQLPositionRepository) CreatePosition(ctx context.Context, position *Position) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO positions (name, description) VALUES (?, ?);", position.Name, position.Description)
	if err != nil {
		return fmt.Errorf("creating position: %w", err)
	}
	return nil
}

func (r *SQLPositionRepository) UpdatePosition(ctx context.Context, position *Position) error {
	_, err := r.db.ExecContext(ctx, "UPDATE positions SET name = ?, description = ? WHERE id = ?;", position.Name, position.Description, position.ID)
	if err != nil {
		return fmt.Errorf("updating position: %w", err)
	}
	return nil
}

func (r *SQLPositionRepository) DeletePosition(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM positions WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("deleting position: %w", err)
	}
	return nil
}

func (r *SQLEmployeeRepository) GetEmployees(ctx context.Context, q string) ([]Employee, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, first_name, last_name, email, job_title, hire_date, salary, status, department_id, created_at FROM employees WHERE first_name LIKE ? OR last_name LIKE ? OR email LIKE ?;", "%"+q+"%", "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, fmt.Errorf("querying employees: %w", err)
	}
	defer rows.Close()
	var employees []Employee

	for rows.Next() {
		var employee Employee
		if err := rows.Scan(&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email, &employee.JobTitle, &employee.HireDate, &employee.Salary, &employee.Status, &employee.DepartmentID, &employee.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning employee: %w", err)
		}
		employees = append(employees, employee)
	}
	return employees, nil
}

func (r *SQLEmployeeRepository) GetEmployeeByID(ctx context.Context, id int) (*Employee, error) {
	var employee Employee
	err := r.db.QueryRowContext(ctx, "SELECT id, first_name, last_name, email, job_title, hire_date, salary, status, department_id, created_at FROM employees WHERE id = ?;", id).Scan(&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email, &employee.JobTitle, &employee.HireDate, &employee.Salary, &employee.Status, &employee.DepartmentID, &employee.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying employee by id: %w", err)
	}
	return &employee, nil
}

func (r *SQLEmployeeRepository) CreateEmployee(ctx context.Context, employee *Employee) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO employees (first_name, last_name, email, job_title, hire_date, salary, status, department_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?);", employee.FirstName, employee.LastName, employee.Email, employee.JobTitle, employee.HireDate, employee.Salary, employee.Status, employee.DepartmentID)
	if err != nil {
		return fmt.Errorf("creating employee: %w", err)
	}
	return nil
}

func (r *SQLEmployeeRepository) DeleteEmployee(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM employees WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("deleting employee: %w", err)
	}
	return nil
}

func (r *SQLApplicationRepository) GetApplications(ctx context.Context, q string) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, email, phone, applied_for, resume_url, status, created_at FROM applications WHERE name LIKE ? OR email LIKE ?;", "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, fmt.Errorf("querying applications: %w", err)
	}
	defer rows.Close()
	var applications []Application

	for rows.Next() {
		var app Application
		if err := rows.Scan(&app.ID, &app.Name, &app.Email, &app.Phone, &app.AppliedFor, &app.ResumeURL, &app.Status, &app.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning application: %w", err)
		}
		applications = append(applications, app)
	}
	return applications, nil
}

func (r *SQLApplicationRepository) GetApplicationByID(ctx context.Context, id int) (*Application, error) {
	var app Application
	err := r.db.QueryRowContext(ctx, "SELECT id, name, email, phone, applied_for, resume_url, status, created_at FROM applications WHERE id = ?;", id).Scan(&app.ID, &app.Name, &app.Email, &app.Phone, &app.AppliedFor, &app.ResumeURL, &app.Status, &app.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying application by id: %w", err)
	}
	return &app, nil
}

func (r *SQLApplicationRepository) CreateApplication(ctx context.Context, app *Application) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO applications (name, email, phone, applied_for, resume_url, status) VALUES (?, ?, ?, ?, ?, ?);", app.Name, app.Email, app.Phone, app.AppliedFor, app.ResumeURL, app.Status)
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}
	return nil
}

func (r *SQLApplicationRepository) DeleteApplication(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM applications WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return nil
}

func (r *SQLLeaveRepository) GetLeaves(ctx context.Context, q string) ([]Leave, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, employee_id, leave_type, start_date, end_date, status, reason, created_at FROM leaves WHERE leave_type LIKE ? OR status LIKE ? OR reason LIKE ?;", "%"+q+"%", "%"+q+"%", "%"+q+"%")
	if err != nil {
		return nil, fmt.Errorf("querying leaves: %w", err)
	}
	defer rows.Close()
	var leaves []Leave

	for rows.Next() {
		var l Leave
		if err := rows.Scan(&l.ID, &l.EmployeeID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.Status, &l.Reason, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning leave: %w", err)
		}
		leaves = append(leaves, l)
	}
	return leaves, nil
}

func (r *SQLLeaveRepository) GetLeaveByID(ctx context.Context, id int) (*Leave, error) {
	var l Leave
	err := r.db.QueryRowContext(ctx, "SELECT id, employee_id, leave_type, start_date, end_date, status, reason, created_at FROM leaves WHERE id = ?;", id).Scan(&l.ID, &l.EmployeeID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.Status, &l.Reason, &l.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying leave by id: %w", err)
	}
	return &l, nil
}

func (r *SQLLeaveRepository) CreateLeave(ctx context.Context, l *Leave) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO leaves (employee_id, leave_type, start_date, end_date, status, reason) VALUES (?, ?, ?, ?, ?, ?);", l.EmployeeID, l.LeaveType, l.StartDate, l.EndDate, l.Status, l.Reason)
	if err != nil {
		return fmt.Errorf("creating leave: %w", err)
	}
	return nil
}

func (r *SQLLeaveRepository) DeleteLeave(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM leaves WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("deleting leave: %w", err)
	}
	return nil
}
