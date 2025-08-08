package handler

// ErrorResponse エラーレスポンス構造体
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 成功レスポンス構造体
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
