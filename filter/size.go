package filter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseSize takes a string representation of a size (e.g. "1G", "10M") and returns
// the size in bytes as an int64. If the input string is not a valid size
// representation, an error is returned.
func ParseSize(sizeStr string) (int64, error) {
	units := map[string]int64{"B": 1, "K": 1 << 10, "M": 1 << 20, "G": 1 << 30, "T": 1 << 40}
	sizeStr = strings.ToUpper(sizeStr)
	unit := sizeStr[len(sizeStr)-1:]
	size, err := strconv.ParseFloat(sizeStr[:len(sizeStr)-1], 64)
	if err != nil {
		return 0, err
	}
	return int64(size * float64(units[unit])), nil
}

// FormatSize takes an int64 representation of a size in bytes and returns a string
// representation of the size with a unit (e.g. "1G", "10M"). The size is rounded to
// the nearest whole number if it is an integer, otherwise it is rounded to one
// decimal place.
func FormatSize(size int64) string {
	units := []string{"", "K", "M", "G", "T", "P"}
	unitIndex := int(math.Log(float64(size)) / math.Log(1024))
	value := float64(size) / math.Pow(1024, float64(unitIndex))
	if unitIndex >= 0 && unitIndex < len(units) {
		if value == math.Floor(value) {
			return fmt.Sprintf("%d%s", int64(value), units[unitIndex])
		}
		return fmt.Sprintf("%.1f%s", value, units[unitIndex])
	} else {
		return fmt.Sprintf("%d", size)
	}
}
