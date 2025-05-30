package networking

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	conns "github.com/terraform-providers/terraform-provider-nutanix/nutanix"
	"github.com/terraform-providers/terraform-provider-nutanix/nutanix/client"
	v3 "github.com/terraform-providers/terraform-provider-nutanix/nutanix/sdks/v3/prism"
	"github.com/terraform-providers/terraform-provider-nutanix/utils"
)

func DataSourceNutanixSubnet() *schema.Resource {
	return &schema.Resource{
		ReadContext:   dataSourceNutanixSubnetRead,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceNutanixDatasourceSubnetResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceDatasourceSubnetStateUpgradeV0,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"subnet_name"},
			},
			"subnet_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"subnet_id"},
			},
			"additional_filter": DataSourceFiltersSchema(),
			"api_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"categories": categoriesSchema(),
			"owner_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"message_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"details": {
							Type:     schema.TypeMap,
							Computed: true,
						},
					},
				},
			},
			"cluster_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vswitch_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_gateway_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnet_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dhcp_server_address": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dhcp_server_address_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ip_config_pool_list_ranges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dhcp_options": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dhcp_domain_name_server_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dhcp_domain_search_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vlan_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"network_function_chain_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpc_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_external": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_nat": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func findSubnetByUUID(conn *v3.Client, uuid string) (*v3.SubnetIntentResponse, error) {
	return conn.V3.GetSubnet(uuid)
}

func findSubnetByName(conn *v3.Client, name string, clientSideFilters []*client.AdditionalFilter) (*v3.SubnetIntentResponse, error) {
	filter := fmt.Sprintf("name==%s", name)
	resp, err := conn.V3.ListAllSubnet(filter, clientSideFilters)
	if err != nil {
		return nil, err
	}

	entities := resp.Entities

	found := make([]*v3.SubnetIntentResponse, 0)
	for _, v := range entities {
		if *v.Spec.Name == name {
			found = append(found, v)
		}
	}

	if len(found) > 1 {
		return nil, fmt.Errorf("your query returned more than one result. Please use subnet_id argument or use additional filters instead")
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("subnet with the given name, not found")
	}

	return found[0], nil
}

func dataSourceNutanixSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection
	conn := meta.(*conns.Client).API

	subnetID, iok := d.GetOk("subnet_id")
	subnetName, nok := d.GetOk("subnet_name")
	var clientSideFilters []*client.AdditionalFilter
	if v, ok := d.GetOk("additional_filter"); ok {
		clientSideFilters = BuildFiltersDataSource(v.(*schema.Set))
	}

	mappings := map[string]string{
		"cluster_name":             "cluster_reference.name",
		"cluster_uuid":             "cluster_reference.uuid",
		"default_gateway_ip":       "ip_config.default_gateway_ip",
		"prefix_length":            "ip_config.prefix_length",
		"subnet_ip":                "ip_config.subnet_ip",
		"dhcp_server_address":      "ip_config.dhcp_server_address.ip",
		"dhcp_server_address_port": "ip_config.dhcp_server_address.port",
		"dhcp_options":             "ip_config.dhcp_options",
		"dhcp_domain_search_list":  "ip_config.dhcp_options.domain_search_list",
	}

	clientSideFilters = ReplaceFilterPrefixes(clientSideFilters, mappings)

	if !iok && !nok {
		return diag.Errorf("please provide one of subnet_id or subnet_name attributes")
	}

	var reqErr error
	var resp *v3.SubnetIntentResponse

	if iok {
		resp, reqErr = findSubnetByUUID(conn, subnetID.(string))
	} else {
		resp, reqErr = findSubnetByName(conn, subnetName.(string), clientSideFilters)
	}

	if reqErr != nil {
		return diag.FromErr(reqErr)
	}

	m, c := setRSEntityMetadata(resp.Metadata)

	if err := d.Set("metadata", m); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("categories", c); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_reference", flattenReferenceValues(resp.Metadata.ProjectReference)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("owner_reference", flattenReferenceValues(resp.Metadata.OwnerReference)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("availability_zone_reference", flattenReferenceValues(resp.Status.AvailabilityZoneReference)); err != nil {
		return diag.FromErr(err)
	}
	if err := flattenClusterReference(resp.Status.ClusterReference, d); err != nil {
		return diag.FromErr(err)
	}

	dgIP := ""
	sIP := ""
	pl := int64(0)
	port := int64(0)
	dhcpSA := make(map[string]interface{})
	dOptions := make(map[string]interface{})
	ipcpl := make([]string, 0)
	dnsList := make([]string, 0)
	dsList := make([]string, 0)

	if resp.Status.Resources.IPConfig != nil {
		dgIP = utils.StringValue(resp.Status.Resources.IPConfig.DefaultGatewayIP)
		pl = utils.Int64Value(resp.Status.Resources.IPConfig.PrefixLength)
		sIP = utils.StringValue(resp.Status.Resources.IPConfig.SubnetIP)

		if resp.Status.Resources.IPConfig.DHCPServerAddress != nil {
			dhcpSA["ip"] = utils.StringValue(resp.Status.Resources.IPConfig.DHCPServerAddress.IP)
			dhcpSA["fqdn"] = utils.StringValue(resp.Status.Resources.IPConfig.DHCPServerAddress.FQDN)
			dhcpSA["ipv6"] = utils.StringValue(resp.Status.Resources.IPConfig.DHCPServerAddress.IPV6)
			port = utils.Int64Value(resp.Status.Resources.IPConfig.DHCPServerAddress.Port)
		}

		if resp.Status.Resources.IPConfig.PoolList != nil {
			pl := resp.Status.Resources.IPConfig.PoolList
			poolList := make([]string, len(pl))
			for k, v := range pl {
				poolList[k] = utils.StringValue(v.Range)
			}
			ipcpl = poolList
		}
		if resp.Status.Resources.IPConfig.DHCPOptions != nil {
			dOptions["boot_file_name"] = utils.StringValue(resp.Status.Resources.IPConfig.DHCPOptions.BootFileName)
			dOptions["domain_name"] = utils.StringValue(resp.Status.Resources.IPConfig.DHCPOptions.DomainName)
			dOptions["tftp_server_name"] = utils.StringValue(resp.Status.Resources.IPConfig.DHCPOptions.TFTPServerName)

			if resp.Status.Resources.IPConfig.DHCPOptions.DomainNameServerList != nil {
				dnsList = utils.StringValueSlice(resp.Status.Resources.IPConfig.DHCPOptions.DomainNameServerList)
			}
			if resp.Status.Resources.IPConfig.DHCPOptions.DomainSearchList != nil {
				dsList = utils.StringValueSlice(resp.Status.Resources.IPConfig.DHCPOptions.DomainSearchList)
			}
		}
	}

	if err := d.Set("dhcp_server_address", dhcpSA); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip_config_pool_list_ranges", ipcpl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dhcp_options", dOptions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dhcp_domain_name_server_list", dnsList); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dhcp_domain_search_list", dsList); err != nil {
		return diag.FromErr(err)
	}

	d.Set("api_version", utils.StringValue(resp.APIVersion))
	d.Set("name", utils.StringValue(resp.Status.Name))
	d.Set("description", utils.StringValue(resp.Status.Description))
	d.Set("state", utils.StringValue(resp.Status.State))
	d.Set("vswitch_name", utils.StringValue(resp.Status.Resources.VswitchName))
	d.Set("subnet_type", utils.StringValue(resp.Status.Resources.SubnetType))
	d.Set("default_gateway_ip", dgIP)
	d.Set("prefix_length", pl)
	d.Set("subnet_ip", sIP)
	d.Set("dhcp_server_address_port", port)
	d.Set("vlan_id", utils.Int64Value(resp.Status.Resources.VlanID))
	d.Set("network_function_chain_reference", flattenReferenceValues(resp.Status.Resources.NetworkFunctionChainReference))
	d.Set("vpc_reference", flattenReferenceValues(resp.Status.Resources.VPCReference))
	d.Set("is_external", resp.Status.Resources.IsExternal)
	d.Set("enable_nat", resp.Status.Resources.EnableNAT)

	d.SetId(utils.StringValue(resp.Metadata.UUID))

	return nil
}

func resourceDatasourceSubnetStateUpgradeV0(ctx context.Context, is map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Entering resourceDatasourceSubnetStateUpgradeV0")
	return resourceNutanixCategoriesMigrateState(is, meta)
}

func resourceNutanixDatasourceSubnetResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"subnet_name"},
			},
			"subnet_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"subnet_id"},
			},
			"api_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"categories": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"owner_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"message_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"details": {
							Type:     schema.TypeMap,
							Computed: true,
						},
					},
				},
			},
			"cluster_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vswitch_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_gateway_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnet_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dhcp_server_address": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dhcp_server_address_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ip_config_pool_list_ranges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dhcp_options": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dhcp_domain_name_server_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dhcp_domain_search_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vlan_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"network_function_chain_reference": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
