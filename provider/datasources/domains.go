package datasources

import (
	"context"
	"fmt"

	"git.sberdevices.ru/iakarpov/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Domains — датасорс для получения списка доменов
func Domains() *schema.Resource {
	return &schema.Resource{
		ReadContext: domainsRead,
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeList,
				Description: "Список всех зарегистрированных доменов.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Имя домена.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Статус домена (например, 'active', 'inactive').",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func domainsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)

	var domains []map[string]interface{}
	err := cli.SendRPCRequest("domains_get", nil, &domains)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка получения списка доменов: %v", err))
	}

	if err := d.Set("domains", domains); err != nil {
		return diag.FromErr(fmt.Errorf("ошибка установки значений: %v", err))
	}

	d.SetId("domains-datasource")

	return diags
}
