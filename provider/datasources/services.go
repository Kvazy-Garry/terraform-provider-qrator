package datasources

import (
	"context"
	"fmt"

	"git.sberdevices.ru/iakarpov/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Services — датасорс для получения списка сервисов
func Services() *schema.Resource {
	return &schema.Resource{
		ReadContext: servicesRead,
		Schema: map[string]*schema.Schema{
			"services": {
				Type:        schema.TypeList,
				Description: "Список всех зарегистрированных сервисов.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Имя сервиса.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Статус сервиса (например, 'running', 'stopped').",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func servicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)

	var services []map[string]interface{}
	err := cli.SendRPCRequest("services_get", nil, &services)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка получения списка сервисов: %v", err))
	}

	if err := d.Set("services", services); err != nil {
		return diag.FromErr(fmt.Errorf("ошибка установки значений: %v", err))
	}

	d.SetId("services-datasource")

	return diags
}
