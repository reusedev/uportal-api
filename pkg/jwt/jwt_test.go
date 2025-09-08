package jwt

import "testing"

func TestParseToken(t *testing.T) {
	token, _ := GenerateToken(12, false, "113dsjhfjdf", "shuzilm##123", "admin")
	t.Log(token)
}
