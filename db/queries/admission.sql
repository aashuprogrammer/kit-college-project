-- name: GetCourse :one
SELECT * FROM courses
WHERE id = $1 LIMIT 1;


-- name: ListCourses :many
SELECT * FROM courses
ORDER BY name;

-- name: CreateAdmission :one
INSERT INTO admissions (
  registration_number,
  course_id,
  full_name,
  father_name,
  mother_name,
  dob,
  gender,
  religion,
  category,
  sub_category,
  caste_certificate_number,
  is_ews,
  domicile_certificate_number,
  domicile_state,
  mobile,
  email,
  aadhar_card_url,
  father_aadhar_card_url,
  tenth_marksheet_url,
  twelfth_marksheet_url,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
)
RETURNING *;

-- name: CreateRegistration :one
INSERT INTO registrations (
  registration_number,
  course_id,
  full_name,
  email,
  mobile
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRegistrationByNumber :one
SELECT * FROM registrations
WHERE registration_number = $1 LIMIT 1;

-- name: GetRegistrationByEmail :one
SELECT * FROM registrations
WHERE email = $1 LIMIT 1;

-- name: GetRegistrationCountForCourseAndYear :one
SELECT COUNT(*) FROM registrations
WHERE course_id = $1 AND registration_number LIKE $2;

-- name: GetAdmission :one
SELECT * FROM admissions
WHERE id = $1 LIMIT 1;

-- name: UpdateAdmissionStatus :one
UPDATE admissions
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreatePayment :one
INSERT INTO payments (
  admission_id,
  order_id,
  amount,
  currency,
  payment_session_id,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetPaymentByOrderID :one
SELECT * FROM payments
WHERE order_id = $1 LIMIT 1;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    cf_payment_id = $3,
    payment_method = $4,
    transaction_time = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListAdmissions :many
SELECT * FROM admissions
ORDER BY created_at DESC;

