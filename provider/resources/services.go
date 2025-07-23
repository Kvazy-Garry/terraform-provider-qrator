package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Service — ресурс для управления сервисами
func Service() *schema.Resource {
	return &schema.Resource{
		CreateContext: serviceCreate,
		ReadContext:   serviceRead,
		UpdateContext: serviceUpdate,
		DeleteContext: serviceDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Название сервиса",
				Required:    true,
			},
			"ports": {
				Type:        schema.TypeMap,
				Description: "Карта портов и протоколов",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func serviceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Get("name").(string)

	// Приводим тип d.Get("ports") к map[string]interface{}
	portsRaw := d.Get("ports").(map[string]interface{})
	ports := flattenPorts(portsRaw)

	// Подготовим порты для отправки в API
	preparedPorts := preparePortMapForAPI(ports)

	// IP-адрес нужно преобразовать в []interface{}
	ipList := []interface{}{"127.0.0.1"}

	// Соберем параметры для вызова SendRPCRequest
	params := [][]interface{}{
		ipList,        // []interface{} {"127.0.0.1"}
		preparedPorts, // []interface{} {карта портов}
		{name},        // []interface{} {строка name}
	}

	err := cli.SendRPCRequest("service_create", params, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка создания сервиса: %v", err))
	}

	d.SetId(name)
	return diags
}

func serviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)

	err := cli.SendRPCRequest("service_name_get", nil, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка проверки существования сервиса: %v", err))
	}

	return diags
}

// Helper function to convert map[int]string into a suitable format for the request
func preparePortMapForAPI(portMap map[int]string) []interface{} {
	portInfo := make([]interface{}, 0, len(portMap))
	for key, value := range portMap {
		portInfo = append(portInfo, map[string]interface{}{
			"port":     key,
			"protocol": value,
		})
	}
	return portInfo
}

func serviceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()

	// Приводим тип d.Get("ports") к map[string]interface{}
	portsRaw := d.Get("ports").(map[string]interface{})
	ports := flattenPorts(portsRaw)

	// Подготовим порты для отправки в API
	preparedPorts := preparePortMapForAPI(ports)

	// Явно преобразуем []string{"127.0.0.1"} в []interface{}
	ipList := []interface{}{"127.0.0.1"}

	err := cli.SendRPCRequest("service_update", [][]interface{}{ipList, preparedPorts, {name}}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка обновления сервиса: %v", err))
	}

	return diags
}

func serviceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()

	err := cli.SendRPCRequest("service_delete", name, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка удаления сервиса: %v", err))
	}

	return diags
}

func flattenPorts(input map[string]interface{}) map[int]string {
	ports := make(map[int]string)
	for k, v := range input {
		portNum, _ := strconv.Atoi(k)
		protocols := v.(string)
		ports[portNum] = protocols
	}
	return ports
}
