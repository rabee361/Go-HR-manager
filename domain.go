package main

import (
	"time"
)

type Department struct {
	ID         int
	name   string
	description   string
	CreatedAt  time.Time
}

type Employee struct {
	ID        int
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Position struct {
	ID        int
	content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}


type DepartmentRepository interface {
	GetDepartments() ([]Department, error)
	GetDepartmentByID(id int) (*Department, error)
	DeleteDepartment(id int) error
	CreateDepartment(department *Department) error
}

type PositionRepository interface {
	GetPositions()
	GetPositionByID()
	DeletePosition()
	CreatePosition()
}

type EmployeeRepository interface {
	GetEmployees()
	GetEmployeeByID()
	DeleteEmployee()
	CreateEmployee()
}