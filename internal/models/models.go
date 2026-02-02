package models

import (
	"time"
)

type Tenant struct {
	ID         uint    `json:"id"`
	TenantName string  `json:"tenant_name"`
	PaymentNo1 string  `json:"payment_no1"` //1 -> 3
	PaymentNo2 string  `json:"payment_no2"`
	Rent       float64 `json:"rent"`
	Balance    float64 `json:"balance"`
	UnitID     uint    `json:"unit_id"`     //FK -> units table
	LandlordID uint    `json:"landlord_id"` //FK -> users table

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// used for sql joins, joining tenant to unit
type TenantWithUnit struct {
	ID         uint    `json:"id"`
	TenantName string  `json:"tenant_name"`
	PaymentNo1 string  `json:"payment_no1"`
	PaymentNo2 string  `json:"payment_no2"`
	Rent       float64 `json:"rent"`
	Balance    float64 `json:"balance"`

	UnitID    uint    `json:"unit_id"`
	UnitName  string  `json:"unit_name"`
	UnitType  string  `json:"unit_type"`
	UnitPrice float64 `json:"unit_price"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Property struct {
	ID           uint      `json:"id"`
	LandlordID   uint      `json:"landlord_id"` //FK userID users table(landlord is a user)
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	PropertyType string    `json:"property_type"`
	Vacancy      bool      `json:"vacancy"`
	TotalRent    int       `json:"total_rent"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// not used at the moment
type Unit struct {
	ID         uint      `json:"id"`
	PropertyID uint      `json:"property_id"` //FK property_id -> properties
	UnitName   string    `json:"unit_name"`
	UnitType   string    `json:"unit_type"`
	UnitPrice  float64   `json:"unit_price"`
	Vacancy    bool      `json:"vacancy"` //in db i've set default vacancy true
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// helper struct for creating unit
type CreateUnitInput struct {
	PropertyID uint    `json:"property_id"`
	UnitName   string  `json:"unit_name" binding:"required"`
	UnitType   string  `json:"unit_type"`
	UnitPrice  float64 `json:"unit_price" binding:"required"`
}

type Notification struct {
	ID       uint   `json:"id"`
	Message  string `json:"message"`
	IsRead   bool   `json:"is_read"`
	TenantID uint   `json:"tenant_id"`

	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID       uint   `json:"id"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`

	CreatedAt time.Time `json:"created_at"`
}

// Payment represents a payment transaction (cash or M-Pesa)
type Payment struct {
	ID         uint      `json:"id"`
	LandlordID uint      `json:"landlord_id"`
	TenantID   *uint     `json:"tenant_id"` // Nullable for unassigned payments
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"` // PENDING, COMPLETED, FAILED
	Method     string    `json:"method"` // CASH, MPESA_TILL, MPESA_PAYBILL
	Receipt    string    `json:"receipt"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LandlordPaymentConfig stores M-Pesa credentials per landlord
type LandlordPaymentConfig struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	LandlordID        uint      `json:"landlord_id" gorm:"unique"` // One config per landlord
	ShortCode         string    `json:"short_code"`                // Paybill or Till Number
	ShortCodeType     string    `json:"short_code_type"`           // "paybill" or "till"
	ConsumerKey       string    `json:"consumer_key"`
	ConsumerSecret    string    `json:"consumer_secret"`
	Environment       string    `json:"environment"`        // "sandbox" or "production"
	ValidationEnabled bool      `json:"validation_enabled"` // If false, we might not register validation URL
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
