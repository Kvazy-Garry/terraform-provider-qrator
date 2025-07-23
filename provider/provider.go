package provider

import (
	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/datasources"
	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider — главная точка входа для Terraform-провайдера
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Description: "Адрес API Qrator",
				Required:    true,
			},
			"token": {
				Type:        schema.TypeString,
				Description: "Токен авторизации",
				Required:    true,
			},
			"client_id": {
				Type:        schema.TypeInt,
				Description: "Идентификатор клиента",
				Required:    true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"qrator_domains":  datasources.Domains(),
			"qrator_services": datasources.Services(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"qrator_domain":        resources.Domain(),
			"qrator_service":       resources.Service(),
			"qrator_domain_ip_set": resources.DomainIPSet(),
		},
		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			config := GetConfig(d)
			return client.NewQRClient(config.Endpoint, config.Token, config.ClientID), nil
		},
	}
}
