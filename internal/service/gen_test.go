package service

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestNewAdminService(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	t.Logf(string(password))
}
