package main

import (
	"context"
	"time"
)

type Department struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type Employee struct {
	ID           int
	FirstName    string
	LastName     string
	Email        string
	JobTitle     string
	HireDate     time.Time
	Salary       float64
	Status       string
	DepartmentID int
	CreatedAt    time.Time
}

type Position struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

type Application struct {
	ID         int
	Name       string
	Email      string
	Phone      string
	AppliedFor string
	ResumeURL  string
	Status     string
	CreatedAt  time.Time
}

type Leave struct {
	ID         int
	EmployeeID int
	LeaveType  string
	StartDate  time.Time
	EndDate    time.Time
	Status     string
	Reason     string
	CreatedAt  time.Time
}

type DepartmentRepository interface {
	GetDepartments(ctx context.Context, q string) ([]Department, error)
	GetDepartmentByID(ctx context.Context, id int) (*Department, error)
	DeleteDepartment(ctx context.Context, id int) error
	CreateDepartment(ctx context.Context, department *Department) error
}

type PositionRepository interface {
	GetPositions(ctx context.Context, q string) ([]Position, error)
	GetPositionByID(ctx context.Context, id int) (*Position, error)
	DeletePosition(ctx context.Context, id int) error
	CreatePosition(ctx context.Context, position *Position) error
}

type EmployeeRepository interface {
	GetEmployees(ctx context.Context, q string) ([]Employee, error)
	GetEmployeeByID(ctx context.Context, id int) (*Employee, error)
	DeleteEmployee(ctx context.Context, id int) error
	CreateEmployee(ctx context.Context, employee *Employee) error
}

type ApplicationRepository interface {
	GetApplications(ctx context.Context, q string) ([]Application, error)
	GetApplicationByID(ctx context.Context, id int) (*Application, error)
	DeleteApplication(ctx context.Context, id int) error
	CreateApplication(ctx context.Context, application *Application) error
}

type LeaveRepository interface {
	GetLeaves(ctx context.Context, q string) ([]Leave, error)
	GetLeaveByID(ctx context.Context, id int) (*Leave, error)
	DeleteLeave(ctx context.Context, id int) error
	CreateLeave(ctx context.Context, leave *Leave) error
}
