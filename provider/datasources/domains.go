package datasources

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
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
				ipJsonList = v
			case map[string]interface{}:
				ipJsonList = []interface{}{v} // Обернуть map в массив
			default:
				log.Printf("[WARN] Неподдерживаемый тип ip_json в домене %d: %T", i, ipJsonRaw)
				ipJsonList = []interface{}{}
			}
		}

		tfDomains[i] = map[string]interface{}{
			"id":         domainMap["id"],
			"name":       domainMap["name"],
			"status":     domainMap["status"],
			"ip":         ipList,
			"ip_json":    ipJsonList, // Используем обработанное значение
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
