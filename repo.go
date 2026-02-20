package main

import (
	"database/sql"
	"log"
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


func NewDepartmentRepository(db *sql.DB) *SQLDepartmentRepository {
	return &SQLDepartmentRepository{db: db}
}

func NewPositionRepository(db *sql.DB) *SQLPositionRepository {
	return &SQLPositionRepository{db: db}
}

func NewEmployeeRepository(db *sql.DB) *SQLEmployeeRepository {
	return &SQLEmployeeRepository{db: db}
}


func (r *SQLDepartmentRepository) GetDepartments() ([]Department, error) {
	rows, err := r.db.Query("SELECT * FROM departments;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var departments []Department
	
	for rows.Next() {
		var department Department
		rows.Scan(&department.ID, &department.name,&department.description, &department.CreatedAt)
		departments = append(departments, department)
	}
	return departments, nil
}