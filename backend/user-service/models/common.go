package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMap JSONB字段通用类型
type JSONMap map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	case map[string]interface{}:
		*j = JSONMap(v)
		return nil
	default:
		*j = JSONMap{}
		return fmt.Errorf("unsupported type for JSONMap: %T", value)
	}
}
