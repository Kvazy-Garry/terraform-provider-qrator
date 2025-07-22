package resources

import (
	"context"
	"fmt"

	"git.sberdevices.ru/iakarpov/terraform-provider-qrator/provider/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DomainIPSet — ресурс для задания IP-адресов домену
func DomainIPSet() *schema.Resource {
	return &schema.Resource{
		CreateContext: domainIPSetCreate,
		ReadContext:   domainIPSetRead,
		UpdateContext: domainIPSetUpdate,
		DeleteContext: domainIPSetDelete,
		Schema: map[string]*schema.Schema{
			"domain_name": {
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
				Required: true,
			},
		},
	}
}

func domainIPSetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	domainName := d.Get("domain_name").(string)
	upstreams := expandStringList(d.Get("upstreams").([]interface{}))

	err := cli.SendRPCRequest("domain_ip_set", [][]interface{}{upstreams, domainName}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка задания IP для домена: %v", err))
	}

	d.SetId(domainName)
	return diags
}

func domainIPSetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	domainName := d.Id()

	err := cli.SendRPCRequest("domain_name_get", nil, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка проверки существования домена: %v", err))
	}

	return diags
}

func domainIPSetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	domainName := d.Id()
	upstreams := expandStringList(d.Get("upstreams").([]interface{}))

	err := cli.SendRPCRequest("domain_ip_set", [][]interface{}{upstreams, domainName}, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("ошибка обновления IP для домена: %v", err))
	}

	return diags
}

func domainIPSetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*client.QRAPI)
	domainName := d.Id()

	err := cli.SendRPCRequest("domain_delete", domainName, nil)
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
