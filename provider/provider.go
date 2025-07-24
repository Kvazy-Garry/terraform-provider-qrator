package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/client"
	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/datasources"
	"github.com/Kvazy-Garry/terraform-provider-qrator/provider/resources"
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
		ResourcesMap: map[string]*schema.Resource{
			"qrator_domain": resourceQratorDomain(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"qrator_domains": datasourceQratorDomains(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func resourceQratorDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resources.DomainCreate,
		ReadContext:   datasources.DomainRead,
		UpdateContext: resources.DomainUpdate,
		DeleteContext: resources.DomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"upstream_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"balancer": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "roundrobin",
						},
						"weights": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"backups": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"upstreams": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"ip": {
										Type:     schema.TypeString,
										Required: true,
									},
									"weight": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qrator_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"backups": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"balancer": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"clusters": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"upstreams": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"ip": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"weight": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"weights": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
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
