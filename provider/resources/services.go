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
	ports := flattenPorts(d.Get("ports"))

	err := cli.SendRPCRequest("service_create", [][]interface{}{[]string{"127.0.0.1"}, ports, name}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка создания сервиса: %v", err))
	}

	d.SetId(name)
	return diags
}

func serviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()

	err := cli.SendRPCRequest("service_name_get", nil, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка проверки существования сервиса: %v", err))
	}

	return diags
}

func serviceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()
	ports := flattenPorts(d.Get("ports"))

	err := cli.SendRPCRequest("service_update", [][]interface{}{[]string{"127.0.0.1"}, ports, name}, nil)
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
