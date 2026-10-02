package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ewallet-monolith/models"
)

type TransactionController struct {
	DB *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{DB: db}
}

func (tc *TransactionController) TopUp(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nominal top-up tidak valid"})
		return
	}

	err := tc.DB.Transaction(func(dbTx *gorm.DB) error {
		var wallet models.Wallet
		if err := dbTx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		wallet.Balance += req.Amount
		if err := dbTx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			ReceiverID: &userID,
			Amount:     req.Amount,
			Type:       models.TypeTopUp,
			Status:     models.StatusSuccess,
		}

		if err := dbTx.Create(&transaction).Error; err != nil {
			return err
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Top-up berhasil",
			"amount_added":   req.Amount,
			"latest_balance": wallet.Balance,
		})

		return nil
	})

	if err != nil && c.Writer.Status() == http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses top-up"})
	}
}

func (tc *TransactionController) InitiatePayment(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	tx := models.Transaction{
		SenderID:  &userID,
		Amount:    req.Amount,
		Type:      models.TypePayment,
		Status:    models.StatusPending,
		ExpiresAt: &expiresAt,
	}

	if err := tc.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Payment session created",
		"transaction_id": tx.ID,
		"amount":         tx.Amount,
		"status":         tx.Status,
		"expires_at":     tx.ExpiresAt,
	})
}

func (tc *TransactionController) ConfirmPayment(c *gin.Context) {
	txID := c.Param("id")
	userID := c.MustGet("user_id").(uint)

	err := tc.DB.Transaction(func(dbTx *gorm.DB) error {
		var transaction models.Transaction
		if err := dbTx.Set("gorm:query_option", "FOR UPDATE").First(&transaction, txID).Error; err != nil {
			return err
		}

		if *transaction.SenderID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized transaction access"})
			return gorm.ErrInvalidTransaction
		}

		if transaction.Status != models.StatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction is already processed or expired"})
			return gorm.ErrInvalidTransaction
		}

		if time.Now().After(*transaction.ExpiresAt) {
			transaction.Status = models.StatusExpired
			dbTx.Save(&transaction)

			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "Payment session has expired. Balance was not deducted.",
				"status": models.StatusExpired,
			})
			return gorm.ErrInvalidTransaction
		}

		var wallet models.Wallet
		if err := dbTx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		if wallet.Balance < transaction.Amount {
			transaction.Status = models.StatusFailed
			dbTx.Save(&transaction)

			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
			return gorm.ErrInvalidTransaction
		}

		wallet.Balance -= transaction.Amount
		if err := dbTx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction.Status = models.StatusSuccess
		if err := dbTx.Save(&transaction).Error; err != nil {
			return err
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Payment successful",
			"transaction_id": transaction.ID,
			"amount_paid":    transaction.Amount,
			"remaining_bal":  wallet.Balance,
			"status":         transaction.Status,
		})

		return nil
	})

	if err != nil && c.Writer.Status() == http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction process failed"})
	}
}