package enum

type LogLevelEnum string

const (
	LogLevelDebug   LogLevelEnum = "debug"
	LogLevelInfo    LogLevelEnum = "info"
	LogLevelWarn    LogLevelEnum = "warn"
	LogLevelWarning LogLevelEnum = "warning"
	LogLevelError   LogLevelEnum = "error"
)

func (e LogLevelEnum) ToString() string {
	switch e {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelWarning, LogLevelError:
		return string(e)
	default:
		return ""
	}
}

func (e LogLevelEnum) IsValid() bool {
	switch e {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelWarning, LogLevelError:
		return true
	default:
		return false
	}
}

type LogFormatEnum string

const (
	LogFormatJSON    LogFormatEnum = "json"
	LogFormatConsole LogFormatEnum = "console"
)

func (e LogFormatEnum) ToString() string {
	switch e {
	case LogFormatJSON, LogFormatConsole:
		return string(e)
	default:
		return ""
	}
}

func (e LogFormatEnum) IsValid() bool {
	switch e {
	case LogFormatJSON, LogFormatConsole:
		return true
	default:
		return false
	}
}
