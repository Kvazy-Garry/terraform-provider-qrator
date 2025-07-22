package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Config — структура конфигурации провайдера
type Config struct {
	Endpoint string
	Token    string
	ClientID int
}

// GetConfig — получаем значения конфигурации из данных ресурса
func GetConfig(d *schema.ResourceData) *Config {
	return &Config{
		Endpoint: d.Get("endpoint").(string),
		Token:    d.Get("token").(string),
		ClientID: d.Get("client_id").(int),
	}
}
