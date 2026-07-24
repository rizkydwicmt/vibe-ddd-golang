package convert

import (
	"strconv"
	"strings"

	"github.com/samber/lo"
)

func GetMapBoolValue(header map[string]interface{}, key string) *bool {
	val := false
	parsed, err := strconv.ParseBool(strings.ToLower(*GetMapStringValue(header, key)))
	return lo.Ternary(err == nil, &parsed, &val)
}
