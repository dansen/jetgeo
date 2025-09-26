package geo

import "fmt"

// Level 行政层级
// 顺序必须保持 Province < City < District 语义 (用于比较)
type Level int

const (
	LevelProvince Level = iota
	LevelCity
	LevelDistrict
)

func (l Level) String() string {
	switch l {
	case LevelProvince:
		return "province"
	case LevelCity:
		return "city"
	case LevelDistrict:
		return "district"
	default:
		return fmt.Sprintf("Level(%d)", int(l))
	}
}

// ParseLevel 将字符串解析为 Level
func ParseLevel(s string) (Level, error) {
	switch s {
	case "province":
		return LevelProvince, nil
	case "city":
		return LevelCity, nil
	case "district":
		return LevelDistrict, nil
	case "":
		return LevelDistrict, nil // 默认最细
	default:
		return 0, fmt.Errorf("unknown level: %s", s)
	}
}

// LessOrEqual Java 中的 lessThen: this.ordinal >= other.ordinal
// 这里定义: a.LessOrEqual(b) => a 层级 <= b 层级深度 (province < city < district)
func (l Level) LessOrEqual(other Level) bool { return l <= other }
