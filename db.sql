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
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    applied_for TEXT, -- Position name
    resume_url TEXT,
    status TEXT DEFAULT 'pending', -- e.g., pending, interviewing, accepted, rejected
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

-- 5. Positions table
CREATE TABLE IF NOT EXISTS positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_leaves_employee_id ON leaves(employee_id);

-- -- Dummy Data for testing
-- INSERT OR IGNORE INTO departments (name, description) VALUES 
-- ('Engineering', 'Software and infrastructure development'),
-- ('Human Resources', 'Recruitment and employee relations'),
-- ('Marketing', 'Brand awareness and sales support');

-- REPLACE INTO employees (first_name, last_name, email, department_id, job_title, hire_date, salary) VALUES 
-- ('Rabee', 'Engineer', 'rabee@example.com', 1, 'Senior Developer', '2023-01-15', 85000),
-- ('John', 'Doe', 'john@example.com', 1, 'Junior Developer', '2023-06-01', 50000),
-- ('Jane', 'Smith', 'jane@example.com', 2, 'HR Manager', '2022-11-20', 70000);

-- REPLACE INTO leaves (employee_id, leave_type, start_date, end_date, status, reason) VALUES 
-- (1, 'vacation', '2024-03-01', '2024-03-07', 'approved', 'Annual leave'),
-- (3, 'sick', '2024-02-15', '2024-02-16', 'approved', 'Flu');

-- INSERT OR IGNORE INTO applications (name, email, phone, resume_url, applied_for, status) VALUES 
-- ('Alice Walker', 'alice@example.com', '1234567890', 'resume.pdf', 'Backend Engineer', 'interviewing'),
-- ('Bob Ross', 'bob@example.com', '1234567890', 'resume.pdf', 'UI Designer', 'pending');

-- INSERT OR IGNORE INTO positions (name, description) VALUES 
-- ('Software Engineer', 'Develop and maintain software applications'),
-- ('Project Manager', 'Plan and execute projects'),
-- ('HR Specialist', 'Manage HR processes');
