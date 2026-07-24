package pointer

func StringPtr(s string) *string {
	return &s
}

func StringNilPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}
