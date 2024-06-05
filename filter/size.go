package filter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

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
