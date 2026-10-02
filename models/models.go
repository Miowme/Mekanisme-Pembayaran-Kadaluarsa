package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"unique;not null" json:"username"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Password  string         `json:"-"` 
	Wallet    Wallet         `json:"wallet,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Wallet struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;unique" json:"user_id"`
	Balance   float64        `gorm:"default:0" json:"balance"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type TransactionType string
type TransactionStatus string

const (
	TypeTopUp    TransactionType = "TOPUP"
	TypeTransfer TransactionType = "TRANSFER"
	TypePayment  TransactionType = "PAYMENT"

	StatusPending TransactionStatus = "PENDING"
	StatusSuccess TransactionStatus = "SUCCESS"
	StatusFailed  TransactionStatus = "FAILED"
	StatusExpired TransactionStatus = "EXPIRED"
)

type Transaction struct {
	ID           uint              `gorm:"primaryKey" json:"id"`
	SenderID     *uint             `json:"sender_id,omitempty"`
	ReceiverID   *uint             `json:"receiver_id,omitempty"`
	Amount       float64           `gorm:"not null" json:"amount"`
	Type         TransactionType   `gorm:"not null" json:"type"`
	Status       TransactionStatus `gorm:"not null;default:'PENDING'" json:"status"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"` 
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}