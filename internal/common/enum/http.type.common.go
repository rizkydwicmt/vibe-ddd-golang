package enum

type HTTPContentTypeEnum string

const (
	ApplicationJSON  HTTPContentTypeEnum = "application/json"
	ApplicationXform HTTPContentTypeEnum = "application/x-www-form-urlencoded"
	MultipartForm    HTTPContentTypeEnum = "multipart/form-data"
)

func (e HTTPContentTypeEnum) ToString() string {
	switch e {
	case ApplicationJSON:
		return "application/json"
	case ApplicationXform:
		return "application/x-www-form-urlencoded"
	case MultipartForm:
		return "multipart/form-data"
	default:
		return ""
	}
}
