package middlewareid

type MiddlewareId string

// 列管middleware function,用於traceMap
const (
	AuthMiddleware        MiddlewareId = "AuthMiddleware"
	TraceIDMiddleware     MiddlewareId = "TraceIDMiddleware"
	ErrorHandleMiddleware MiddlewareId = "ErrorHandleMiddleware"
	IPWhiteListMiddleware MiddlewareId = "IPWhiteListMiddleware"
	LogOriginRequest      MiddlewareId = "logOriginRequest"
)
