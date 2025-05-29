package handler

// Handlers 处理器集合
type Handlers struct {
	AuthHandler    *AuthHandler
	AdminHandler   *AdminHandler
	TokenHandler   *TokenHandler
	TaskHandler    *TaskHandler
	PaymentHandler *PaymentHandler
	InviteHandler  *InviteHandler
}

// NewHandlers 创建处理器集合
func NewHandlers(
	authHandler *AuthHandler,
	adminHandler *AdminHandler,
	tokenHandler *TokenHandler,
	taskHandler *TaskHandler,
	paymentHandler *PaymentHandler,
	inviteHandler *InviteHandler,
) *Handlers {
	return &Handlers{
		AuthHandler:    authHandler,
		AdminHandler:   adminHandler,
		TokenHandler:   tokenHandler,
		TaskHandler:    taskHandler,
		PaymentHandler: paymentHandler,
		InviteHandler:  inviteHandler,
	}
}
