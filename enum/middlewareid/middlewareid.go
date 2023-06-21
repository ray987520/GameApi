package middlewareid

type MiddlewareId string

// 列管middleware function,用於traceMap,middleware以mw_開頭
const (
	AuthMiddleware        MiddlewareId = "mw_AuthMiddleware"
	TraceIDMiddleware     MiddlewareId = "mw_TraceIDMiddleware"
	ErrorHandleMiddleware MiddlewareId = "mw_ErrorHandleMiddleware"
	IPWhiteListMiddleware MiddlewareId = "mw_IPWhiteListMiddleware"
	LogOriginRequest      MiddlewareId = "mw_logOriginRequest"
	TotalTimeMiddleware   MiddlewareId = "mw_TotalTimeMiddleware"
)
