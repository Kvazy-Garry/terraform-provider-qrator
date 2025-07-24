package utils

import "encoding/json"

// ToJSON — преобразование значения в строку JSON
func ToJSON(value interface{}) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON — обратное преобразование строки JSON в интерфейс
func FromJSON(str string, result interface{}) error {
	return json.Unmarshal([]byte(str), result)
}

func NormalizeIpJson(ipJson map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	for k, v := range ipJson {
		switch k {
		case "backups", "clusters", "weights":
			if b, ok := v.(bool); ok {
				normalized[k] = b
			} else {
				normalized[k] = false
			}
		case "balancer":
			if s, ok := v.(string); ok {
				normalized[k] = s
			} else {
				normalized[k] = ""
			}
		case "upstreams":
			if upstreams, ok := v.([]interface{}); ok {
				var normalizedUpstreams []interface{}
				for _, u := range upstreams {
					if upstream, ok := u.(map[string]interface{}); ok {
						normalizedUpstream := make(map[string]interface{})
						if ip, ok := upstream["ip"].(string); ok {
							normalizedUpstream["ip"] = ip
						}
						if name, ok := upstream["name"].(string); ok {
							normalizedUpstream["name"] = name
						}
						if typ, ok := upstream["type"].(string); ok {
							normalizedUpstream["type"] = typ
						}
						if weight, ok := upstream["weight"].(float64); ok {
							normalizedUpstream["weight"] = int(weight)
						}
						normalizedUpstreams = append(normalizedUpstreams, normalizedUpstream)
					}
				}
				normalized[k] = normalizedUpstreams
			}
		default:
			normalized[k] = v
		}
	}

	return normalized
}

// Вспомогательные функции
func ExpandUpstreams(upstreams []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(upstreams))
	for i, u := range upstreams {
		upstream := u.(map[string]interface{})
		result[i] = map[string]interface{}{
			"type":   upstream["type"],
			"ip":     upstream["ip"],
			"weight": upstream["weight"],
			"name":   upstream["name"],
		}
	}
	return result
}

func FlattenUpstreams(upstreams []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(upstreams))
	for i, u := range upstreams {
		upstream := u.(map[string]interface{})
		result[i] = map[string]interface{}{
			"type":   upstream["type"],
			"ip":     upstream["ip"],
			"weight": int(upstream["weight"].(float64)),
			"name":   upstream["name"],
		}
	}
	return result
}

func ExpandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, v.(string))
	}
	return vs
}

func IsNotFoundError(err error) bool {
	return err.Error() == "domain not found"
}
