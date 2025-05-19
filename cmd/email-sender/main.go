package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

type EmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())

	// SMTP configuration (using environment variables for security)
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")// Gmail email address
	smtpPassword := os.Getenv("SMTP_PASSWORD")// Gmail App Password


	// Email sending endpoint
	r.POST("/api/email/send", func(c *gin.Context) {
		var req EmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Configure the email message
		m := gomail.NewMessage()
		m.SetHeader("From", smtpUsername)
		m.SetHeader("To", req.To)
		m.SetHeader("Subject", req.Subject)
		m.SetBody("text/plain", req.Body)

		// Configure the SMTP dialer
		d := gomail.NewDialer(smtpHost, 587, smtpUsername, smtpPassword)

		// Send the email
		if err := d.DialAndSend(m); err != nil {
			log.Printf("Failed to send email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
	})

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	port := os.Getenv("EMAIL_SERVICE_PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Email Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start Email Service: %v", err)
	}
}