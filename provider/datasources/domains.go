package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DomainsRead экспортируемая функция для чтения доменов
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

		tfDomains[i] = map[string]interface{}{
			"id":         domainMap["id"],
			"name":       domainMap["name"],
			"status":     domainMap["status"],
			"ip":         ipList,
			"ip_json":    domainMap["ip_json"],
			"qrator_ip":  domainMap["qratorIp"],
			"is_service": domainMap["isService"],
			"ports":      domainMap["ports"],
		}
	}

	if err := d.Set("domains", tfDomains); err != nil {
		return diag.FromErr(fmt.Errorf("ошибка сохранения доменов: %v", err))
	}

	d.SetId(fmt.Sprintf("%d", time.Now().UnixNano()))
	return nil
}
