package enum

type EnvEnum string

const (
	LOCAL       EnvEnum = "local"
	DEVELOPMENT EnvEnum = "development"
	PRODUCTION  EnvEnum = "production"
	STAGING     EnvEnum = "staging"
)

func (e EnvEnum) ToString() string {
	switch e {
	case LOCAL:
		return "local"
	case DEVELOPMENT:
		return "development"
	case PRODUCTION:
		return "production"
	case STAGING:
		return "staging"
	}
	return ""
}

func (e EnvEnum) IsValid() bool {
	switch e {
	case LOCAL, DEVELOPMENT, PRODUCTION, STAGING:
		return true
	}
	return false
}
