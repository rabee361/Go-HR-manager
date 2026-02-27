-- Dummy Data for testing
INSERT INTO departments (name, description) VALUES 
('Engineering', 'Software and infrastructure development'),
('Human Resources', 'Recruitment and employee relations'),
('Marketing', 'Brand awareness and sales support');

REPLACE INTO employees (first_name, last_name, email, department_id, job_title, hire_date, salary) VALUES 
('Rabee', 'Engineer', 'rabee@example.com', 1, 'Senior Developer', '2023-01-15', 85000),
('John', 'Doe', 'john@example.com', 1, 'Junior Developer', '2023-06-01', 50000),
('Jane', 'Smith', 'jane@example.com', 2, 'HR Manager', '2022-11-20', 70000);

REPLACE INTO leaves (employee_id, leave_type, start_date, end_date, status, reason) VALUES 
(1, 'vacation', '2024-03-01', '2024-03-07', 'approved', 'Annual leave'),
(3, 'sick', '2024-02-15', '2024-02-16', 'approved', 'Flu');

REPLACE INTO applications (applicant_name, email, applied_for, status) VALUES 
('Alice Walker', 'alice@example.com', 'Backend Engineer', 'interviewing'),
('Bob Ross', 'bob@example.com', 'UI Designer', 'pending');

REPLACE INTO positions (name, description) VALUES 
('Software Engineer', 'Develop and maintain software applications'),
('Project Manager', 'Plan and execute projects'),
('HR Specialist', 'Manage HR processes');
