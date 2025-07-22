package resources

import (
	"context"
	"fmt"

	"git.sberdevices.ru/iakarpov/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Domain — ресурс для управления доменами
func Domain() *schema.Resource {
	return &schema.Resource{
		CreateContext: domainCreate,
		ReadContext:   domainRead,
		UpdateContext: domainUpdate,
		DeleteContext: domainDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Имя домена",
				Required:    true,
			},
			"upstreams": {
				Type:        schema.TypeList,
				Description: "Список серверов (IP)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func domainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Get("name").(string)
	upstreams := expandStringList(d.Get("upstreams").([]interface{}))

	err := cli.SendRPCRequest("domain_create", [][]interface{}{upstreams, name}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка создания домена: %v", err))
	}

	d.SetId(name)
	return diags
}

func domainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()

	err := cli.SendRPCRequest("domain_name_get", nil, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка проверки существования домена: %v", err))
	}

	return diags
}

func domainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()
	upstreams := expandStringList(d.Get("upstreams").([]interface{}))

	err := cli.SendRPCRequest("domain_ip_set", [][]interface{}{upstreams, name}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка обновления домена: %v", err))
	}

	return diags
}

func domainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	name := d.Id()

	err := cli.SendRPCRequest("domain_delete", name, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка удаления домена: %v", err))
	}

	return diags
}

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if !ok {
			continue
		}
		vs = append(vs, val)
	}
	return vs
}
