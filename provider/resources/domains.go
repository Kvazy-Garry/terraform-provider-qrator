package resources

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*client.QRAPI)
	// var diags diag.Diagnostics

	name := d.Get("name").(string)
	ipList := expandStringList(d.Get("ip_list").([]interface{}))
	upstreamConfig := d.Get("upstream_config").([]interface{})

	// Формируем параметры для API
	var params []interface{}
	if len(upstreamConfig) > 0 {
		config := upstreamConfig[0].(map[string]interface{})
		params = []interface{}{
			map[string]interface{}{
				"balancer":  config["balancer"],
				"weights":   config["weights"],
				"backups":   config["backups"],
				"upstreams": expandUpstreams(config["upstreams"].([]interface{})),
			},
			name,
		}
	} else {
		params = []interface{}{ipList, name}
	}

	// Создаем домен
	var domainID int
	err := cli.SendRPCRequest("domain_create", params, &domainID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка создания домена: %v", err))
	}

	d.SetId(strconv.Itoa(domainID))
	log.Printf("[INFO] Создан домен %s (ID: %d)", name, domainID)

	// Читаем созданный домен для заполнения computed полей
	return DomainRead(ctx, d, m)
}

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
		if isNotFoundError(err) {
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
				"upstreams": flattenUpstreams(ipJson["upstreams"].([]interface{})),
			},
		}
		d.Set("upstream_config", upstreamConfig)
	}

	return diags
}

func DomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*client.QRAPI)

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("неверный ID домена: %v", err))
	}

	// Обновляем базовые параметры
	if d.HasChange("ip_list") || d.HasChange("upstream_config") {
		ipList := expandStringList(d.Get("ip_list").([]interface{}))
		upstreamConfig := d.Get("upstream_config").([]interface{})

		var params []interface{}
		if len(upstreamConfig) > 0 {
			config := upstreamConfig[0].(map[string]interface{})
			params = []interface{}{
				domainID,
				map[string]interface{}{
					"balancer":  config["balancer"],
					"weights":   config["weights"],
					"backups":   config["backups"],
					"upstreams": expandUpstreams(config["upstreams"].([]interface{})),
				},
			}
		} else {
			params = []interface{}{domainID, ipList}
		}

		err := cli.SendRPCRequest("domain_update", params, nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("ошибка обновления домена: %v", err))
		}
		log.Printf("[INFO] Домен ID %d обновлен", domainID)
	}

	return DomainRead(ctx, d, m)
}

func DomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*client.QRAPI)
	var diags diag.Diagnostics

	domainID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("неверный ID домена: %v", err))
	}

	err = cli.SendRPCRequest("domain_delete", []interface{}{domainID}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка удаления домена: %v", err))
	}

	d.SetId("")
	log.Printf("[INFO] Домен ID %d удален", domainID)
	return diags
}

// Вспомогательные функции
func expandUpstreams(upstreams []interface{}) []map[string]interface{} {
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

func flattenUpstreams(upstreams []interface{}) []map[string]interface{} {
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

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, v.(string))
	}
	return vs
}

func isNotFoundError(err error) bool {
	return err.Error() == "domain not found"
}
