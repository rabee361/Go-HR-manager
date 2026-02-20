-- Database Schema for HR App

-- 1. Departments table
CREATE TABLE IF NOT EXISTS departments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. Employees table
CREATE TABLE IF NOT EXISTS employees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    department_id INTEGER,
    job_title TEXT,
    hire_date DATE,
    salary REAL,
    status TEXT DEFAULT 'active', -- e.g., active, inactive, suspended
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (department_id) REFERENCES departments(id)
);

-- 3. Applications table (Job applications)
CREATE TABLE IF NOT EXISTS applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    applicant_name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    applied_for TEXT, -- Position name
    resume_url TEXT,
    status TEXT DEFAULT 'pending', -- e.g., pending, interviewing, accepted, rejected
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 4. Leaves table (Vacation/Sick leave requests)
CREATE TABLE IF NOT EXISTS leaves (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    employee_id INTEGER NOT NULL,
    leave_type TEXT NOT NULL, -- e.g., vacation, sick, personal
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT DEFAULT 'pending', -- e.g., pending, approved, rejected
    reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (employee_id) REFERENCES employees(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_leaves_employee_id ON leaves(employee_id);

-- Dummy Data for testing
INSERT INTO departments (name, description) VALUES 
('Engineering', 'Software and infrastructure development'),
('Human Resources', 'Recruitment and employee relations'),
('Marketing', 'Brand awareness and sales support');

INSERT INTO employees (first_name, last_name, email, department_id, job_title, hire_date, salary) VALUES 
('Rabee', 'Engineer', 'rabee@example.com', 1, 'Senior Developer', '2023-01-15', 85000),
('John', 'Doe', 'john@example.com', 1, 'Junior Developer', '2023-06-01', 50000),
('Jane', 'Smith', 'jane@example.com', 2, 'HR Manager', '2022-11-20', 70000);

INSERT INTO leaves (employee_id, leave_type, start_date, end_date, status, reason) VALUES 
(1, 'vacation', '2024-03-01', '2024-03-07', 'approved', 'Annual leave'),
(3, 'sick', '2024-02-15', '2024-02-16', 'approved', 'Flu');

INSERT INTO applications (applicant_name, email, applied_for, status) VALUES 
('Alice Walker', 'alice@example.com', 'Backend Engineer', 'interviewing'),
('Bob Ross', 'bob@example.com', 'UI Designer', 'pending');
