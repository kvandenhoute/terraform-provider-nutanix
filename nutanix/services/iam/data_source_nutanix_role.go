package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	conns "github.com/terraform-providers/terraform-provider-nutanix/nutanix"
	v3 "github.com/terraform-providers/terraform-provider-nutanix/nutanix/sdks/v3/prism"
	"github.com/terraform-providers/terraform-provider-nutanix/utils"
)

func DataSourceNutanixRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNutanixRoleRead,
		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"role_name"},
			},
			"role_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"role_id"},
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
			"permission_reference_list": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNutanixRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection
	conn := meta.(*conns.Client).API

	accessID, iok := d.GetOk("role_id")
	roleName, rnOk := d.GetOk("role_name")

	if !iok && !rnOk {
		return diag.Errorf("please provide `role_id` or `role_name`")
	}

	var err error
	var resp *v3.Role

	if iok {
		resp, err = conn.V3.GetRole(accessID.(string))
	}
	if rnOk {
		resp, err = findRoleByName(conn, roleName.(string))
	}

	if err != nil {
		return diag.FromErr(err)
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
	if err := d.Set("api_version", resp.APIVersion); err != nil {
		return diag.FromErr(err)
	}

	if status := resp.Status; status != nil {
		if err := d.Set("name", utils.StringValue(resp.Status.Name)); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("description", utils.StringValue(resp.Status.Description)); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("state", utils.StringValue(resp.Status.State)); err != nil {
			return diag.FromErr(err)
		}

		if res := status.Resources; res != nil {
			if err := d.Set("permission_reference_list", flattenArrayReferenceValues(status.Resources.PermissionReferenceList)); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	d.SetId(utils.StringValue(resp.Metadata.UUID))

	return nil
}

func findRoleByName(conn *v3.Client, name string) (*v3.Role, error) {
	filter := fmt.Sprintf("name==%s", name)
	resp, err := conn.V3.ListAllRole(filter)
	if err != nil {
		return nil, err
	}

	entities := resp.Entities

	found := make([]*v3.Role, 0)
	for _, v := range entities {
		if *v.Spec.Name == name {
			found = append(found, v)
		}
	}

	if len(found) > 1 {
		return nil, fmt.Errorf("your query returned more than one result. Please use role_id argument instead")
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("role with the given name, not found")
	}

	return found[0], nil
}
