package errors

type HTTPError struct {
	Code       int    // HTTP статус код, который будет отправлен клиенту
	Message    string // Сообщение, которое будет отправлено клиенту в теле ответа
	InnerError error  // Оригинальная ошибка, для внутреннего логирования и отладки (не для клиента)
}

// конструктор, дабы удобно возвращать ошибку
func NewHTTPError(code int, message string, inner error) *HTTPError {
	return &HTTPError{
		Code:       code,
		Message:    message,
		InnerError: inner,
	}
}

func (e *HTTPError) Error() string {
	return e.Message
}
