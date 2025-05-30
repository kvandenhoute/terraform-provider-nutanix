package nke

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	conns "github.com/terraform-providers/terraform-provider-nutanix/nutanix"
	"github.com/terraform-providers/terraform-provider-nutanix/nutanix/sdks/v3/karbon"
	"github.com/terraform-providers/terraform-provider-nutanix/utils"
)

func DataSourceNutanixKarbonPrivateRegistry() *schema.Resource {
	return &schema.Resource{
		ReadContext:   dataSourceNutanixKarbonPrivateRegistryRead,
		SchemaVersion: 1,
		Schema:        KarbonPrivateRegistryDataSourceMap(),
	}
}

func dataSourceNutanixKarbonPrivateRegistryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection
	conn := meta.(*conns.Client).KarbonAPI
	setTimeout(meta)
	// Make request to the API
	karbonPrivateRegistryID, iok := d.GetOk("private_registry_id")
	karbonPrivateRegistryName, nok := d.GetOk("private_registry_name")
	if !iok && !nok {
		return diag.Errorf("please provide one of private_registry_id or private_registry_name attributes")
	}
	var err error
	var resp *karbon.PrivateRegistryResponse

	if iok {
		resp, err = findPrivateRegistryByUUID(conn, karbonPrivateRegistryID.(string))
	} else {
		resp, err = findPrivateRegistryByName(conn, karbonPrivateRegistryName.(string))
	}

	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	uuid := utils.StringValue(resp.UUID)
	if err := d.Set("name", utils.StringValue(resp.Name)); err != nil {
		return diag.Errorf("error occurred while setting name: %s", err)
	}
	if err := d.Set("endpoint", utils.StringValue(resp.Endpoint)); err != nil {
		return diag.Errorf("error occurred while setting endpoint: %s", err)
	}
	if err := d.Set("uuid", uuid); err != nil {
		return diag.Errorf("error occurred while setting endpoint: %s", err)
	}
	d.SetId(uuid)

	return nil
}

func findPrivateRegistryByName(conn *karbon.Client, name string) (*karbon.PrivateRegistryResponse, error) {
	return conn.PrivateRegistry.GetKarbonPrivateRegistry(name)
}

func findPrivateRegistryByUUID(conn *karbon.Client, uuid string) (*karbon.PrivateRegistryResponse, error) {
	resp, err := conn.PrivateRegistry.ListKarbonPrivateRegistries()
	if err != nil {
		return nil, err
	}

	found := make([]*karbon.PrivateRegistryResponse, 0)
	for _, v := range *resp {
		reg := v
		if *v.UUID == uuid {
			found = append(found, &reg)
		}
	}

	if len(found) > 1 {
		return nil, fmt.Errorf("your query returned more than one result. Please use private_registry_name argument instead")
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("private registry with the given uuid not found")
	}

	return found[0], nil
}

func KarbonPrivateRegistryDataSourceMap() map[string]*schema.Schema {
	kcsm := KarbonPrivateRegistryElementDataSourceMap()
	kcsm["private_registry_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"private_registry_name"},
	}
	kcsm["private_registry_name"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"private_registry_id"},
	}
	return kcsm
}

func KarbonPrivateRegistryElementDataSourceMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"uuid": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
