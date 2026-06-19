-- name: GetCourse :one
SELECT * FROM courses
WHERE code = $1 LIMIT 1;

-- name: ListCourses :many
SELECT * FROM courses
ORDER BY name;

-- name: CreateAdmission :one
INSERT INTO admissions (
  course_code,
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
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

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
