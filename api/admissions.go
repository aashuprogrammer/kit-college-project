package api

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aashuprogrammer/fee-management-system/db/pgdb"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

type studentRegisterRequest struct {
	CourseID int32  `json:"course_id" form:"course_id" validate:"required"`
	FullName string `json:"full_name" form:"full_name" validate:"required"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Mobile   string `json:"mobile" form:"mobile" validate:"required"`
}

type studentRegisterResponse struct {
	RegistrationNumber string `json:"registration_number"`
	FullName           string `json:"full_name"`
	Email              string `json:"email"`
	Mobile             string `json:"mobile"`
	CourseID           int32  `json:"course_id"`
}

type registrationDetailsResponse struct {
	RegistrationNumber string `json:"registration_number"`
	CourseID           int32  `json:"course_id"`
	CourseName         string `json:"course_name"`
	FullName           string `json:"full_name"`
	Email              string `json:"email"`
	Mobile             string `json:"mobile"`
}

type createAdmissionRequest struct {
	RegistrationNumber        string `json:"registration_number" form:"registration_number" validate:"required"`
	CourseID                  int32  `json:"course_id" form:"course_id" validate:"required"`
	FullName                  string `json:"full_name" form:"full_name" validate:"required"`
	FatherName                string `json:"father_name" form:"father_name" validate:"required"`
	MotherName                string `json:"mother_name" form:"mother_name" validate:"required"`
	DOB                       string `json:"dob" form:"dob" validate:"required"` // Format: YYYY-MM-DD
	Gender                    string `json:"gender" form:"gender" validate:"required"`
	Religion                  string `json:"religion" form:"religion" validate:"required"`
	Category                  string `json:"category" form:"category" validate:"required"`
	SubCategory               string `json:"sub_category" form:"sub_category"`
	CasteCertificateNumber    string `json:"caste_certificate_number" form:"caste_certificate_number"`
	IsEWS                     bool   `json:"is_ews" form:"is_ews"`
	DomicileCertificateNumber string `json:"domicile_certificate_number" form:"domicile_certificate_number"`
	DomicileState             string `json:"domicile_state" form:"domicile_state" validate:"required"`
	Mobile                    string `json:"mobile" form:"mobile" validate:"required"`
	Email                     string `json:"email" form:"email" validate:"required,email"`
	ReturnURL                 string `json:"return_url" form:"return_url" validate:"omitempty"`
}

type createAdmissionResponse struct {
	AdmissionID      int32   `json:"admission_id"`
	OrderID          string  `json:"order_id"`
	PaymentSessionID string  `json:"payment_session_id"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
}

type verifyPaymentRequest struct {
	OrderID string `json:"order_id" validate:"required"`
}

type verifyPaymentResponse struct {
	OrderID         string  `json:"order_id"`
	AdmissionID     int32   `json:"admission_id"`
	PaymentStatus   string  `json:"payment_status"`
	AdmissionStatus string  `json:"admission_status"`
}

func (server *Server) createAdmission(c *fiber.Ctx) error {
	var req createAdmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	validationErrors := server.validate(req)
	if validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	// 1. Verify that the registration number exists
	registration, err := server.store.GetRegistrationByNumber(c.Context(), req.RegistrationNumber)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return BadRequestError("invalid registration number: registration not found")
		}
		return InternalServerError("failed to look up registration details: " + err.Error())
	}

	// Verify that the course exists
	course, err := server.store.GetCourse(c.Context(), registration.CourseID)
	if err != nil {
		return InternalServerError("failed to look up course: " + err.Error())
	}

	// 2. Parse Date of Birth
	dobTime, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return BadRequestError("invalid dob format, must be YYYY-MM-DD")
	}

	// 3. Extract and upload files
	aadharHeader, err := c.FormFile("aadhar_card")
	if err != nil {
		return BadRequestError("aadhar_card file is required")
	}
	fatherAadharHeader, err := c.FormFile("father_aadhar_card")
	if err != nil {
		return BadRequestError("father_aadhar_card file is required")
	}
	tenthMarksheetHeader, err := c.FormFile("tenth_marksheet")
	if err != nil {
		return BadRequestError("tenth_marksheet file is required")
	}
	twelfthMarksheetHeader, err := c.FormFile("twelfth_marksheet")
	if err != nil {
		return BadRequestError("twelfth_marksheet file is required")
	}

	uploadDocument := func(fileHeader *multipart.FileHeader, docType string) (string, error) {
		file, err := fileHeader.Open()
		if err != nil {
			return "", err
		}
		defer file.Close()

		ext := filepath.Ext(fileHeader.Filename)
		if ext == "" {
			ext = ".bin"
		}

		key := fmt.Sprintf("admissions/%s/%s%s", req.RegistrationNumber, docType, ext)
		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		return server.r2Client.UploadFile(c.Context(), key, file, contentType)
	}

	aadharURL, err := uploadDocument(aadharHeader, "aadhar_card")
	if err != nil {
		return InternalServerError("failed to upload Aadhar Card: " + err.Error())
	}
	fatherAadharURL, err := uploadDocument(fatherAadharHeader, "father_aadhar_card")
	if err != nil {
		return InternalServerError("failed to upload Father's Aadhar Card: " + err.Error())
	}
	tenthMarksheetURL, err := uploadDocument(tenthMarksheetHeader, "tenth_marksheet")
	if err != nil {
		return InternalServerError("failed to upload 10th Marksheet: " + err.Error())
	}
	twelfthMarksheetURL, err := uploadDocument(twelfthMarksheetHeader, "twelfth_marksheet")
	if err != nil {
		return InternalServerError("failed to upload 12th Marksheet: " + err.Error())
	}

	// 4. Create a pending admission record
	admission, err := server.store.CreateAdmission(c.Context(), pgdb.CreateAdmissionParams{
		RegistrationNumber:        req.RegistrationNumber,
		CourseID:                  registration.CourseID,
		FullName:                  registration.FullName,
		FatherName:                req.FatherName,
		MotherName:                req.MotherName,
		Dob:                       pgtype.Date{Time: dobTime, Valid: true},
		Gender:                    req.Gender,
		Religion:                  req.Religion,
		Category:                  req.Category,
		SubCategory:               pgtype.Text{String: req.SubCategory, Valid: req.SubCategory != ""},
		CasteCertificateNumber:    pgtype.Text{String: req.CasteCertificateNumber, Valid: req.CasteCertificateNumber != ""},
		IsEws:                     req.IsEWS,
		DomicileCertificateNumber: pgtype.Text{String: req.DomicileCertificateNumber, Valid: req.DomicileCertificateNumber != ""},
		DomicileState:             req.DomicileState,
		Mobile:                    registration.Mobile,
		Email:                     registration.Email,
		AadharCardUrl:             aadharURL,
		FatherAadharCardUrl:       fatherAadharURL,
		TenthMarksheetUrl:         tenthMarksheetURL,
		TwelfthMarksheetUrl:       twelfthMarksheetURL,
		Status:                    "PENDING",
	})
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.UniqueViolation {
			return BadRequestError("admission details have already been submitted for this registration number")
		}
		return InternalServerError("failed to store admission details: " + err.Error())
	}

	// 5. Generate unique order ID
	orderID := fmt.Sprintf("adm_%d_%d", admission.ID, time.Now().Unix())
	courseFee := float64(course.FeeAmount)

	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = "https://example.com/payment-status"
	}

	// 6. Initiate Order with Cashfree
	cfOrder, err := server.cfClient.CreateOrder(
		c.Context(),
		orderID,
		courseFee,
		fmt.Sprintf("stud_%d", admission.ID),
		registration.FullName,
		registration.Email,
		registration.Mobile,
		returnURL,
	)
	if err != nil {
		// Clean up by marking admission as FAILED or deleting it (let's mark as FAILED)
		_, _ = server.store.UpdateAdmissionStatus(c.Context(), pgdb.UpdateAdmissionStatusParams{
			ID:     admission.ID,
			Status: "FAILED",
		})
		return InternalServerError("failed to initialize payment gateway order: " + err.Error())
	}

	// 7. Record the pending payment details
	_, err = server.store.CreatePayment(c.Context(), pgdb.CreatePaymentParams{
		AdmissionID:      admission.ID,
		OrderID:          orderID,
		Amount:           course.FeeAmount,
		Currency:         "INR",
		PaymentSessionID: cfOrder.PaymentSessionID,
		Status:           "PENDING",
	})
	if err != nil {
		return InternalServerError("failed to store payment information: " + err.Error())
	}

	return c.JSON(createAdmissionResponse{
		AdmissionID:      admission.ID,
		OrderID:          orderID,
		PaymentSessionID: cfOrder.PaymentSessionID,
		Amount:           courseFee,
		Currency:         "INR",
		Status:           "PENDING",
	})
}

func (server *Server) register(c *fiber.Ctx) error {
	var req studentRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	validationErrors := server.validate(req)
	if validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	// 1. Verify course exists
	course, err := server.store.GetCourse(c.Context(), req.CourseID)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return BadRequestError("invalid course id: course not found")
		}
		return InternalServerError("failed to lookup course: " + err.Error())
	}

	// 2. Check if email is already registered
	_, err = server.store.GetRegistrationByEmail(c.Context(), req.Email)
	if err == nil {
		return BadRequestError("email address is already registered")
	} else if pgdb.ErrorCode(err) != pgdb.ErrorNoRow {
		return InternalServerError("failed to check registration: " + err.Error())
	}

	// 3. Generate Unique Registration Number
	// Format: [2-digit Year][College Code][Branch Code][Unique Student Number]
	// Example: 25CSE001
	year := time.Now().Format("06") // "26" for 2026
	collegeCode := server.config.CollegeCode
	branchCode := course.CourseCode // E.g. "CSE"

	prefix := fmt.Sprintf("%s%s%s", year, collegeCode, branchCode)
	likePattern := prefix + "%"

	var regNumber string
	// Retry loop for unique constraint safety
	for i := 0; i < 5; i++ {
		count, err := server.store.GetRegistrationCountForCourseAndYear(c.Context(), pgdb.GetRegistrationCountForCourseAndYearParams{
			CourseID:           req.CourseID,
			RegistrationNumber: likePattern,
		})
		if err != nil {
			return InternalServerError("failed to generate registration number: " + err.Error())
		}

		nextNum := count + 1
		regNumber = fmt.Sprintf("%s%03d", prefix, nextNum)

		// Create registration
		registration, err := server.store.CreateRegistration(c.Context(), pgdb.CreateRegistrationParams{
			RegistrationNumber: regNumber,
			CourseID:           req.CourseID,
			FullName:           req.FullName,
			Email:              req.Email,
			Mobile:             req.Mobile,
		})
		if err == nil {
			return c.JSON(studentRegisterResponse{
				RegistrationNumber: registration.RegistrationNumber,
				FullName:           registration.FullName,
				Email:              registration.Email,
				Mobile:             registration.Mobile,
				CourseID:           registration.CourseID,
			})
		}

		if pgdb.ErrorCode(err) != pgdb.UniqueViolation {
			return InternalServerError("failed to save registration details: " + err.Error())
		}
	}

	return InternalServerError("failed to generate unique registration number after multiple attempts")
}

func (server *Server) getRegistration(c *fiber.Ctx) error {
	regNum := c.Params("reg_num")
	if regNum == "" {
		return BadRequestError("registration number is required")
	}

	registration, err := server.store.GetRegistrationByNumber(c.Context(), regNum)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("registration number not found")
		}
		return InternalServerError("failed to look up registration: " + err.Error())
	}

	course, err := server.store.GetCourse(c.Context(), registration.CourseID)
	if err != nil {
		return InternalServerError("failed to lookup registration course: " + err.Error())
	}

	return c.JSON(registrationDetailsResponse{
		RegistrationNumber: registration.RegistrationNumber,
		CourseID:           registration.CourseID,
		CourseName:         course.Name,
		FullName:           registration.FullName,
		Email:              registration.Email,
		Mobile:             registration.Mobile,
	})
}

func (server *Server) verifyPayment(c *fiber.Ctx) error {
	var req verifyPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	validationErrors := server.validate(req)
	if validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	// 1. Fetch payment details from database
	payment, err := server.store.GetPaymentByOrderID(c.Context(), req.OrderID)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("payment order not found")
		}
		return InternalServerError("failed to fetch payment: " + err.Error())
	}

	// 2. Query Cashfree for the current order status
	cfOrder, err := server.cfClient.GetOrder(c.Context(), req.OrderID)
	if err != nil {
		return InternalServerError("failed to verify payment status: " + err.Error())
	}

	admissionStatus := "PENDING"
	paymentStatus := payment.Status

	// 3. If successfully paid, confirm admission
	if cfOrder.OrderStatus == "PAID" {
		err = server.store.ExecTx(c.Context(), func(q *pgdb.Queries) error {
			// Update payment record
			_, err = q.UpdatePaymentStatus(c.Context(), pgdb.UpdatePaymentStatusParams{
				ID:              payment.ID,
				Status:          "SUCCESS",
				CfPaymentID:     pgtype.Text{String: "VERIFIED", Valid: true},
				PaymentMethod:   pgtype.Text{String: "UNKNOWN", Valid: true},
				TransactionTime: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			})
			if err != nil {
				return err
			}

			// Update admission record to CONFIRMED
			_, err = q.UpdateAdmissionStatus(c.Context(), pgdb.UpdateAdmissionStatusParams{
				ID:     payment.AdmissionID,
				Status: "CONFIRMED",
			})
			return err
		})

		if err != nil {
			return InternalServerError("failed to update payment state: " + err.Error())
		}
		paymentStatus = "SUCCESS"
		admissionStatus = "CONFIRMED"
	} else if cfOrder.OrderStatus == "EXPIRED" || cfOrder.OrderStatus == "FAILED" {
		err = server.store.ExecTx(c.Context(), func(q *pgdb.Queries) error {
			_, err = q.UpdatePaymentStatus(c.Context(), pgdb.UpdatePaymentStatusParams{
				ID:              payment.ID,
				Status:          "FAILED",
				CfPaymentID:     pgtype.Text{String: "FAILED", Valid: true},
				PaymentMethod:   pgtype.Text{String: "UNKNOWN", Valid: true},
				TransactionTime: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			})
			if err != nil {
				return err
			}

			_, err = q.UpdateAdmissionStatus(c.Context(), pgdb.UpdateAdmissionStatusParams{
				ID:     payment.AdmissionID,
				Status: "FAILED",
			})
			return err
		})
		if err != nil {
			return InternalServerError("failed to update payment state: " + err.Error())
		}
		paymentStatus = "FAILED"
		admissionStatus = "FAILED"
	}

	return c.JSON(verifyPaymentResponse{
		OrderID:         req.OrderID,
		AdmissionID:     payment.AdmissionID,
		PaymentStatus:   paymentStatus,
		AdmissionStatus: admissionStatus,
	})
}

// Cashfree Webhook Payloads
type webhookOrder struct {
	OrderID       string  `json:"order_id"`
	OrderAmount   float64 `json:"order_amount"`
	OrderCurrency string  `json:"order_currency"`
}

type webhookPayment struct {
	CFPaymentID     interface{} `json:"cf_payment_id"`
	PaymentStatus   string      `json:"payment_status"`
	PaymentAmount   float64     `json:"payment_amount"`
	PaymentCurrency string      `json:"payment_currency"`
	PaymentTime     string      `json:"payment_time"`
	PaymentMethod   interface{} `json:"payment_method"`
}

type webhookData struct {
	Order   webhookOrder   `json:"order"`
	Payment webhookPayment `json:"payment"`
}

type cashfreeWebhookPayload struct {
	Data      webhookData `json:"data"`
	EventTime string      `json:"event_time"`
	Type      string      `json:"type"`
}

func (server *Server) handleWebhook(c *fiber.Ctx) error {
	signature := c.Get("x-webhook-signature")
	timestamp := c.Get("x-webhook-timestamp")
	rawBody := c.Body()

	// 1. Verify incoming signature
	if signature == "" || timestamp == "" || !server.cfClient.VerifyWebhookSignature(signature, timestamp, string(rawBody)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid webhook signature"})
	}

	// 2. Parse the payload
	var payload cashfreeWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to parse webhook payload"})
	}

	// 3. Retrieve database payment record
	payment, err := server.store.GetPaymentByOrderID(c.Context(), payload.Data.Order.OrderID)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query order"})
	}

	// Prevent redundant processing if order already verified as success
	if payment.Status == "SUCCESS" {
		return c.SendStatus(fiber.StatusOK)
	}

	cfPaymentIDStr := fmt.Sprintf("%v", payload.Data.Payment.CFPaymentID)
	txTime, err := time.Parse(time.RFC3339, payload.Data.Payment.PaymentTime)
	if err != nil {
		txTime = time.Now()
	}

	methodBytes, _ := json.Marshal(payload.Data.Payment.PaymentMethod)
	methodStr := string(methodBytes)

	// 4. Update status depending on the payload status
	if payload.Data.Payment.PaymentStatus == "SUCCESS" {
		err = server.store.ExecTx(c.Context(), func(q *pgdb.Queries) error {
			_, err = q.UpdatePaymentStatus(c.Context(), pgdb.UpdatePaymentStatusParams{
				ID:              payment.ID,
				Status:          "SUCCESS",
				CfPaymentID:     pgtype.Text{String: cfPaymentIDStr, Valid: cfPaymentIDStr != ""},
				PaymentMethod:   pgtype.Text{String: methodStr, Valid: methodStr != ""},
				TransactionTime: pgtype.Timestamptz{Time: txTime, Valid: true},
			})
			if err != nil {
				return err
			}

			_, err = q.UpdateAdmissionStatus(c.Context(), pgdb.UpdateAdmissionStatusParams{
				ID:     payment.AdmissionID,
				Status: "CONFIRMED",
			})
			return err
		})
	} else if payload.Data.Payment.PaymentStatus == "FAILED" || payload.Data.Payment.PaymentStatus == "CANCELLED" {
		err = server.store.ExecTx(c.Context(), func(q *pgdb.Queries) error {
			_, err = q.UpdatePaymentStatus(c.Context(), pgdb.UpdatePaymentStatusParams{
				ID:              payment.ID,
				Status:          "FAILED",
				CfPaymentID:     pgtype.Text{String: cfPaymentIDStr, Valid: cfPaymentIDStr != ""},
				PaymentMethod:   pgtype.Text{String: methodStr, Valid: methodStr != ""},
				TransactionTime: pgtype.Timestamptz{Time: txTime, Valid: true},
			})
			if err != nil {
				return err
			}

			_, err = q.UpdateAdmissionStatus(c.Context(), pgdb.UpdateAdmissionStatusParams{
				ID:     payment.AdmissionID,
				Status: "FAILED",
			})
			return err
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to record transaction state"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (server *Server) listAdmissions(c *fiber.Ctx) error {
	admissions, err := server.store.ListAdmissions(c.Context())
	if err != nil {
		return InternalServerError("failed to retrieve admissions: " + err.Error())
	}
	return c.JSON(admissions)
}

