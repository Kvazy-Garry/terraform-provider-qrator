package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/datasources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("QRATOR_TOKEN", nil),
			},
			"client_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://api.qrator.net",
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"qrator_domains": datasourceQratorDomains(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func datasourceQratorDomains() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasources.DomainsRead,
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ip_json": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
							},
						},
						"qrator_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_service": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ports": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	clientID := d.Get("client_id").(int)
	endpoint := d.Get("endpoint").(string)

	if token == "" {
		return nil, diag.FromErr(fmt.Errorf("токен не может быть пустым"))
	}
	if clientID <= 0 {
		return nil, diag.FromErr(fmt.Errorf("client_id должен быть положительным числом"))
	}

	client := client.NewQRClient(endpoint, token, clientID)
	log.Printf("[DEBUG] Создан клиент Qrator API: endpoint=%s, clientID=%d", endpoint, clientID)

	return client, nil
}
