CREATE TABLE courses (
  code VARCHAR(50) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  fee_amount NUMERIC(10, 2) NOT NULL
);

CREATE TABLE admissions (
  id SERIAL PRIMARY KEY,
  course_code VARCHAR(50) NOT NULL REFERENCES courses(code),
  full_name VARCHAR(255) NOT NULL,
  father_name VARCHAR(255) NOT NULL,
  mother_name VARCHAR(255) NOT NULL,
  dob DATE NOT NULL,
  gender VARCHAR(50) NOT NULL,
  religion VARCHAR(100) NOT NULL,
  category VARCHAR(100) NOT NULL,
  sub_category VARCHAR(100),
  caste_certificate_number VARCHAR(100),
  is_ews BOOLEAN NOT NULL DEFAULT FALSE,
  domicile_certificate_number VARCHAR(100),
  domicile_state VARCHAR(100) NOT NULL,
  mobile VARCHAR(20) NOT NULL,
  email VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, CONFIRMED, FAILED
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
  id SERIAL PRIMARY KEY,
  admission_id INT NOT NULL REFERENCES admissions(id) ON DELETE CASCADE,
  order_id VARCHAR(100) UNIQUE NOT NULL,
  amount NUMERIC(10, 2) NOT NULL,
  currency VARCHAR(10) NOT NULL DEFAULT 'INR',
  payment_session_id VARCHAR(255) NOT NULL,
  cf_payment_id VARCHAR(100),
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, SUCCESS, FAILED, ACTIVE
  payment_method VARCHAR(100),
  transaction_time TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed data for courses
INSERT INTO courses (code, name, fee_amount) VALUES
('CSE', 'Computer Science & Engineering', 1500.00),
('ECE', 'Electronics & Communication Engineering', 1300.00),
('ME', 'Mechanical Engineering', 1200.00),
('EE', 'Electrical Engineering', 1200.00),
('CE', 'Civil Engineering', 1100.00)
ON CONFLICT (code) DO NOTHING;
