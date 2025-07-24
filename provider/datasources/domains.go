package datasources

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DomainsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*client.QRAPI)

	// Выполняем запрос к API
	var response interface{}
	err := cli.SendRPCRequest("domains_get", nil, &response)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка запроса доменов: %v", err))
	}

	// Преобразуем ответ в список доменов
	var domains []interface{}
	switch v := response.(type) {
	case []interface{}:
		domains = v
	case map[string]interface{}:
		if result, ok := v["result"].([]interface{}); ok {
			domains = result
		} else {
			return diag.FromErr(fmt.Errorf("неверный формат ответа: ожидался массив доменов"))
		}
	default:
		return diag.FromErr(fmt.Errorf("неподдерживаемый формат ответа API: %T", response))
	}

	// Подготавливаем данные для Terraform
	tfDomains := make([]map[string]interface{}, len(domains))
	for i, domain := range domains {
		domainMap, ok := domain.(map[string]interface{})
		if !ok {
			return diag.FromErr(fmt.Errorf("неверный формат домена #%d", i))
		}

		// Обрабатываем ip
		var ipList []interface{}
		if ipRaw, exists := domainMap["ip"]; exists {
			switch v := ipRaw.(type) {
			case []interface{}:
				ipList = v
			case string:
				ipList = []interface{}{v}
			case nil:
				ipList = []interface{}{}
			default:
				return diag.FromErr(fmt.Errorf("неподдерживаемый тип IP в домене #%d: %T", i, ipRaw))
			}
		}

		// Обрабатываем ip_json
		var ipJsonList []interface{}
		if ipJsonRaw, exists := domainMap["ip_json"]; exists && ipJsonRaw != nil {
			switch v := ipJsonRaw.(type) {
			case []interface{}:
				for _, item := range v {
					if ipJsonMap, ok := item.(map[string]interface{}); ok {
						ipJsonList = append(ipJsonList, utils.NormalizeIpJson(ipJsonMap))
					}
				}
			case map[string]interface{}:
				ipJsonList = append(ipJsonList, utils.NormalizeIpJson(v))
			default:
				log.Printf("[WARN] Неподдерживаемый тип ip_json в домене %d: %T", i, ipJsonRaw)
			}
		}

		tfDomains[i] = map[string]interface{}{
			"id":         domainMap["id"],
			"name":       domainMap["name"],
			"status":     domainMap["status"],
			"ip":         ipList,
			"ip_json":    ipJsonList,
			"qrator_ip":  domainMap["qratorIp"],
			"is_service": domainMap["isService"],
			"ports":      domainMap["ports"],
		}

		log.Printf("[DEBUG] Processed domain %d: %+v", i, tfDomains[i])
	}

	if err := d.Set("domains", tfDomains); err != nil {
		log.Printf("[ERROR] Failed to set domains: %v", err)
		return diag.FromErr(fmt.Errorf("ошибка сохранения доменов: %v", err))
	}

	d.SetId(fmt.Sprintf("%d", time.Now().UnixNano()))
	return nil
}

// func normalizeIpJson(ipJson map[string]interface{}) map[string]interface{} {
// 	normalized := make(map[string]interface{})

// 	for k, v := range ipJson {
// 		switch k {
// 		case "backups", "clusters", "weights":
// 			if b, ok := v.(bool); ok {
// 				normalized[k] = b
// 			} else {
// 				normalized[k] = false
// 			}
// 		case "balancer":
// 			if s, ok := v.(string); ok {
// 				normalized[k] = s
// 			} else {
// 				normalized[k] = ""
// 			}
// 		case "upstreams":
// 			if upstreams, ok := v.([]interface{}); ok {
// 				var normalizedUpstreams []interface{}
// 				for _, u := range upstreams {
// 					if upstream, ok := u.(map[string]interface{}); ok {
// 						normalizedUpstream := make(map[string]interface{})
// 						if ip, ok := upstream["ip"].(string); ok {
// 							normalizedUpstream["ip"] = ip
// 						}
// 						if name, ok := upstream["name"].(string); ok {
// 							normalizedUpstream["name"] = name
// 						}
// 						if typ, ok := upstream["type"].(string); ok {
// 							normalizedUpstream["type"] = typ
// 						}
// 						if weight, ok := upstream["weight"].(float64); ok {
// 							normalizedUpstream["weight"] = int(weight)
// 						}
// 						normalizedUpstreams = append(normalizedUpstreams, normalizedUpstream)
// 					}
// 				}
// 				normalized[k] = normalizedUpstreams
// 			}
// 		default:
// 			normalized[k] = v
// 		}
// 	}

// 	return normalized
// }

func DomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*client.QRAPI)
	var diags diag.Diagnostics

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("неверный ID домена: %v", err))
	}

	// Получаем информацию о домене
	var domain map[string]interface{}
	err = cli.SendRPCRequest("domain_get", []interface{}{domainID}, &domain)
	if err != nil {
		if utils.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("ошибка получения домена: %v", err))
	}

	// Заполняем атрибуты
	d.Set("name", domain["name"])
	d.Set("status", domain["status"])
	d.Set("qrator_ip", domain["qratorIp"])

	if ipList, ok := domain["ip"].([]interface{}); ok {
		d.Set("ip_list", ipList)
	}

	if ipJson, ok := domain["ip_json"].(map[string]interface{}); ok {
		upstreamConfig := []map[string]interface{}{
			{
				"balancer":  ipJson["balancer"],
				"weights":   ipJson["weights"],
				"backups":   ipJson["backups"],
				"upstreams": utils.FlattenUpstreams(ipJson["upstreams"].([]interface{})),
			},
		}
		d.Set("upstream_config", upstreamConfig)
	}

	return diags
}
