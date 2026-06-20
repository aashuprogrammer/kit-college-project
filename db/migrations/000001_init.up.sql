CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password TEXT NOT NULL
);

CREATE TABLE courses (
  id SERIAL PRIMARY KEY,
  course_code TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  fee_amount INT NOT NULL
);

CREATE TABLE admissions (
  id SERIAL PRIMARY KEY,
  course_id INT NOT NULL REFERENCES courses(id),
  full_name TEXT NOT NULL,
  father_name TEXT NOT NULL,
  mother_name TEXT NOT NULL,
  dob DATE NOT NULL,
  gender TEXT NOT NULL,
  religion TEXT NOT NULL,
  category TEXT NOT NULL,
  sub_category TEXT,
  caste_certificate_number TEXT,
  is_ews BOOLEAN NOT NULL DEFAULT FALSE,
  domicile_certificate_number TEXT,
  domicile_state TEXT NOT NULL,
  mobile TEXT NOT NULL,
  email TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
  id SERIAL PRIMARY KEY,
  admission_id INT NOT NULL REFERENCES admissions(id) ON DELETE CASCADE,
  order_id TEXT UNIQUE NOT NULL,
  amount INT NOT NULL,
  currency TEXT NOT NULL DEFAULT 'INR',
  payment_session_id TEXT NOT NULL,
  cf_payment_id TEXT,
  status TEXT NOT NULL DEFAULT 'PENDING',
  payment_method TEXT,
  transaction_time TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
