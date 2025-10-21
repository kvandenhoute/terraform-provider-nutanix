package vmmv2

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	import3 "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/prism/v4/config"
	import2 "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/ahv/config"
	import1 "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/content"
	conns "github.com/terraform-providers/terraform-provider-nutanix/nutanix"
	"github.com/terraform-providers/terraform-provider-nutanix/utils"
)

func ResourceNutanixOvaVMDeploymentV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceNutanixOvaVMDeploymentCreate,
		ReadContext:   ResourceNutanixOvaVMDeploymentRead,
		UpdateContext: ResourceNutanixOvaVMDeploymentUpdate,
		DeleteContext: ResourceNutanixOvaVMDeploymentDelete,
		Schema: map[string]*schema.Schema{
			"ext_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"override_vm_config": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"num_sockets": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"num_cores_per_socket": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"num_threads_per_core": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"memory_size_bytes": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"power_state": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "ON",
							ValidateFunc: validation.StringInSlice([]string{"ON", "OFF"}, false),
						},
						"nics": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ext_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"backing_info": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"model": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"mac_address": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"is_connected": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"num_queues": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  1,
												},
											},
										},
									},
									"network_info": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"nic_type": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														"SPAN_DESTINATION_NIC",
														"NORMAL_NIC", "DIRECT_NIC", "NETWORK_FUNCTION_NIC",
													}, false),
												},
												"network_function_chain": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ext_id": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"network_function_nic_type": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														"TAP", "EGRESS",
														"INGRESS",
													}, false),
												},
												"subnet": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ext_id": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"vlan_mode": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"TRUNK", "ACCESS"}, false),
												},
												"trunked_vlans": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type: schema.TypeInt,
													},
												},
												"should_allow_unknown_macs": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"ipv4_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"should_assign_ip": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"ip_address": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"value": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																		"prefix_length": {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
															"secondary_ip_address_list": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"value": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																		"prefix_length": {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"ipv4_info": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"learned_ip_addresses": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"value": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																		"prefix_length": {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"cd_roms": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ext_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_address": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bus_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"IDE", "SATA"}, false),
												},
												"index": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"backing_info": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_size_bytes": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"storage_container": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"ext_id": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"storage_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"is_flash_mode_enabled": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"data_source": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"reference": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"image_reference": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"image_ext_id": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																				},
																			},
																		},
																		"vm_disk_reference": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"disk_ext_id": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																					"disk_address": {
																						Type:     schema.TypeList,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"bus_type": {
																									Type:     schema.TypeString,
																									Optional: true,
																									ValidateFunc: validation.StringInSlice([]string{
																										"SCSI", "SPAPR", "PCI",
																										"IDE", "SATA",
																									}, false),
																								},
																								"index": {
																									Type:     schema.TypeInt,
																									Optional: true,
																								},
																							},
																						},
																					},
																					"vm_reference": {
																						Type:     schema.TypeList,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"ext_id": {
																									Type:     schema.TypeString,
																									Optional: true,
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									"iso_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"OTHER", "GUEST_TOOLS", "GUEST_CUSTOMIZATION"}, false),
									},
								},
							},
						},
						"categories": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ext_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"cluster_location_ext_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func ResourceNutanixOvaVMDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.Client).VmmAPI

	extID := d.Get("ext_id").(string)
	log.Printf("🚀🚀🚀 [DEBUG] Starting OVA VM deployment with OVA ext_id: %s 🚀🚀🚀", extID)
	log.Printf("🔥🔥🔥 [INFO] CUSTOM PROVIDER IS BEING USED - OVA DEPLOYMENT STARTING 🔥🔥🔥")
	vmDeploymentSpec := &import1.OvaDeploymentSpec{}
	if clusterLocationExtID := d.Get("cluster_location_ext_id").(string); clusterLocationExtID != "" {
		vmDeploymentSpec.ClusterLocationExtId = &clusterLocationExtID
	}
	if overrideVMConfig, ok := d.GetOk("override_vm_config"); ok {
		overrideVMConfigList := overrideVMConfig.([]interface{})
		if len(overrideVMConfigList) > 0 && overrideVMConfigList[0] != nil {
			ovm := overrideVMConfigList[0].(map[string]interface{})
			overrideSpec := &import2.OvaVmConfigOverrideSpec{}

			if v, ok := ovm["name"].(string); ok && v != "" {
				overrideSpec.Name = &v
			}
			if v, ok := ovm["num_sockets"].(int); ok && v != 0 {
				overrideSpec.NumSockets = &v
			}
			if v, ok := ovm["num_cores_per_socket"].(int); ok && v != 0 {
				overrideSpec.NumCoresPerSocket = &v
			}
			if v, ok := ovm["num_threads_per_core"].(int); ok && v != 0 {
				overrideSpec.NumThreadsPerCore = &v
			}
			if v, ok := ovm["memory_size_bytes"].(int); ok && v != 0 {
				mem := int64(v)
				overrideSpec.MemorySizeBytes = &mem
			}
			// Handle categories
			if cats, ok := ovm["categories"]; ok {
				overrideSpec.Categories = expandCategoryReference(cats.([]interface{}))
			}
			// Handle nics
			if nics, ok := ovm["nics"]; ok {
				overrideSpec.Nics = expandNic(nics.([]interface{}))
			}
			// Handle cd_roms
			if cdroms, ok := ovm["cd_roms"]; ok {
				overrideSpec.CdRoms = expandCdRom(cdroms.([]interface{}))
			}
			vmDeploymentSpec.OverrideVmConfig = overrideSpec
		}
	}

	log.Printf("🔥🔥🔥 [DEBUG] Calling DeployOva API with OVA ext_id: %s 🔥🔥🔥", extID)
	resp, err := conn.OvasAPIInstance.DeployOva(&extID, vmDeploymentSpec)
	if err != nil {
		log.Printf("[ERROR] Failed to deploy OVA: %v", err)
		return diag.FromErr(err)
	}

	TaskRef := resp.Data.GetValue().(import3.TaskReference)
	taskUUID := TaskRef.ExtId
	log.Printf("[DEBUG] OVA deployment task started with UUID: %s", utils.StringValue(taskUUID))

	taskconn := meta.(*conns.Client).PrismAPI
	// Wait for the deployed vm from ova to be available
	stateConf := &resource.StateChangeConf{
		Pending: []string{"QUEUED", "RUNNING", "PENDING"},
		Target:  []string{"SUCCEEDED"},
		Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
		log.Printf("[ERROR] OVA deployment task failed with UUID %s: %v", utils.StringValue(taskUUID), errWaitTask)
		return diag.Errorf("error in Ova VM Deployment (%s): %s", utils.StringValue(taskUUID), errWaitTask)
	}

	log.Printf("[DEBUG] OVA deployment task completed successfully with UUID: %s", utils.StringValue(taskUUID))

	name := ""
	if overrideVMConfig, ok := d.GetOk("override_vm_config"); ok {
		overrideVMConfigList := overrideVMConfig.([]interface{})
		if len(overrideVMConfigList) > 0 && overrideVMConfigList[0] != nil {
			ovm := overrideVMConfigList[0].(map[string]interface{})
			if v, ok := ovm["name"].(string); ok {
				name = v
			}
		}
	}
	filterf := "name eq '" + name + "'"
	filter := utils.StringPtr(filterf)
	offset := new(int)
	limit := new(int)
	*offset = 0
	*limit = 1
	log.Printf("[DEBUG] Listing VMs to find deployed VM with filter: %s", filterf)
	listResp, err := conn.VMAPIInstance.ListVms(offset, limit, filter, nil, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to list VMs: %v", err)
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Processing ListVms response")
	vms, ok := listResp.Data.GetValue().([]import2.Vm)
	if !ok || len(vms) == 0 {
		log.Printf("[ERROR] No VM found with filter: %s", filterf)
		return diag.Errorf("No VM found with filter: %s", filterf)
	}
	log.Printf("[DEBUG] Found deployed VM: count=%d, vm_id=%s", len(vms), *vms[0].ExtId)
	getResp := vms[0]

	d.SetId(*getResp.ExtId)
	log.Printf("✅✅✅ [DEBUG] OVA VM deployment completed successfully: vm_id=%s, name=%s ✅✅✅", *getResp.ExtId, utils.StringValue(getResp.Name))
	log.Printf("🎉🎉🎉 [INFO] CUSTOM PROVIDER DEPLOYMENT COMPLETED SUCCESSFULLY 🎉🎉🎉")

	return nil
}

func ResourceNutanixOvaVMDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.Client).VmmAPI

	resp, err := conn.VMAPIInstance.GetVmById(utils.StringPtr(d.Id()))
	if err != nil {
		return diag.Errorf("error while fetching vm : %v", err)
	}

	getResp := resp.Data.GetValue().(import2.Vm)
	return setOvaVMConfig(d, getResp)
}

func ResourceNutanixOvaVMDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.Client).VmmAPI

	log.Printf("[DEBUG] 🔄 Starting OVA VM update for VM ID: %s", d.Id())

	// Check for hot-plug changes that require VM power off
	hotPlugRequired := false
	var oldConfig, newConfig interface{}

	if d.HasChange("override_vm_config") {
		oldConfig, newConfig = d.GetChange("override_vm_config")
		if oldList, ok := oldConfig.([]interface{}); ok && len(oldList) > 0 {
			if newList, ok := newConfig.([]interface{}); ok && len(newList) > 0 {
				oldMap := oldList[0].(map[string]interface{})
				newMap := newList[0].(map[string]interface{})

				// Check if hot-plug sensitive fields changed
				hotPlugFields := []string{"num_sockets", "num_cores_per_socket", "num_threads_per_core", "memory_size_bytes"}
				for _, field := range hotPlugFields {
					if oldMap[field] != newMap[field] {
						hotPlugRequired = true
						log.Printf("[DEBUG] 🔥 Hot-plug change detected for field: %s", field)
						break
					}
				}
			}
		}
	}

	// Power off VM if hot-plug changes are required
	if hotPlugRequired && !isVMPowerOff(d, conn) {
		log.Printf("[DEBUG] 🔌 VM needs to be powered off for hot-plug changes")
		if err := callForPowerOffVM(ctx, conn, d, meta); err != nil {
			return err
		}
	}

	// Handle basic VM configuration updates (CPU, memory, name)

	// Get current VM state
	updatedVMResp, err := conn.VMAPIInstance.GetVmById(utils.StringPtr(d.Id()))
	if err != nil {
		return diag.Errorf("error reading VM for update: %v", err)
	}

	respVM := updatedVMResp.Data.GetValue().(import2.Vm)
	updateSpec := respVM
	checkForUpdateParams := false

	// Get the override_vm_config data
	if overrideVMConfig, ok := d.GetOk("override_vm_config"); ok {
		overrideVMConfigList := overrideVMConfig.([]interface{})
		if len(overrideVMConfigList) > 0 {
			overrideConfig := overrideVMConfigList[0].(map[string]interface{})

			// Handle individual field changes
			if name, exists := overrideConfig["name"]; exists && name != nil && name.(string) != "" {
				currentName := ""
				if respVM.Name != nil {
					currentName = *respVM.Name
				}
				if name.(string) != currentName {
					updateSpec.Name = utils.StringPtr(name.(string))
					checkForUpdateParams = true
					log.Printf("[DEBUG] 📝 Updating VM name from '%s' to '%s'", currentName, name.(string))
				}
			}

			if numSockets, exists := overrideConfig["num_sockets"]; exists && numSockets != nil && numSockets.(int) > 0 {
				currentSockets := 0
				if respVM.NumSockets != nil {
					currentSockets = *respVM.NumSockets
				}
				if numSockets.(int) != currentSockets {
					updateSpec.NumSockets = utils.IntPtr(numSockets.(int))
					checkForUpdateParams = true
					log.Printf("[DEBUG] 🖥️ Updating VM sockets from %d to %d", currentSockets, numSockets.(int))
				}
			}

			if numCoresPerSocket, exists := overrideConfig["num_cores_per_socket"]; exists && numCoresPerSocket != nil && numCoresPerSocket.(int) > 0 {
				currentCores := 0
				if respVM.NumCoresPerSocket != nil {
					currentCores = *respVM.NumCoresPerSocket
				}
				if numCoresPerSocket.(int) != currentCores {
					updateSpec.NumCoresPerSocket = utils.IntPtr(numCoresPerSocket.(int))
					checkForUpdateParams = true
					log.Printf("[DEBUG] 🖥️ Updating VM cores per socket from %d to %d", currentCores, numCoresPerSocket.(int))
				}
			}

			if numThreadsPerCore, exists := overrideConfig["num_threads_per_core"]; exists && numThreadsPerCore != nil && numThreadsPerCore.(int) > 0 {
				currentThreads := 0
				if respVM.NumThreadsPerCore != nil {
					currentThreads = *respVM.NumThreadsPerCore
				}
				if numThreadsPerCore.(int) != currentThreads {
					updateSpec.NumThreadsPerCore = utils.IntPtr(numThreadsPerCore.(int))
					checkForUpdateParams = true
					log.Printf("[DEBUG] 🖥️ Updating VM threads per core from %d to %d", currentThreads, numThreadsPerCore.(int))
				}
			}

			if memorySizeBytes, exists := overrideConfig["memory_size_bytes"]; exists && memorySizeBytes != nil && memorySizeBytes.(int) > 0 {
				currentMemory := int64(0)
				if respVM.MemorySizeBytes != nil {
					currentMemory = *respVM.MemorySizeBytes
				}
				if int64(memorySizeBytes.(int)) != currentMemory {
					updateSpec.MemorySizeBytes = utils.Int64Ptr(int64(memorySizeBytes.(int)))
					checkForUpdateParams = true
					log.Printf("[DEBUG] 💾 Updating VM memory from %d to %d", currentMemory, memorySizeBytes.(int))
				}
			}
		}
	}

	// Apply basic VM configuration updates if needed
	if checkForUpdateParams {
		log.Printf("[DEBUG] 📤 Applying VM configuration updates")
		// Extract E-Tag Header
		args := make(map[string]interface{})
		args["If-Match"] = getEtagHeader(updatedVMResp, conn)

		updateResp, err := conn.VMAPIInstance.UpdateVmById(utils.StringPtr(d.Id()), &updateSpec, args)
		if err != nil {
			return diag.Errorf("error updating VM: %v", err)
		}

		TaskRef := updateResp.Data.GetValue().(import3.TaskReference)
		taskUUID := TaskRef.ExtId

		taskconn := meta.(*conns.Client).PrismAPI
		// Wait for the VM update to complete
		stateConf := &resource.StateChangeConf{
			Pending: []string{"QUEUED", "RUNNING"},
			Target:  []string{"SUCCEEDED"},
			Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
			Timeout: d.Timeout(schema.TimeoutUpdate),
		}

		if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
			return diag.Errorf("error waiting for VM update task (%s): %s", utils.StringValue(taskUUID), errWaitTask)
		}
		log.Printf("[DEBUG] ✅ VM configuration update completed successfully")
	}

	// Power VM back on if it was powered off for hot-plug changes
	if hotPlugRequired {
		// Check if the desired power state is ON
		if overrideVMConfig, ok := d.GetOk("override_vm_config"); ok {
			overrideVMConfigList := overrideVMConfig.([]interface{})
			if len(overrideVMConfigList) > 0 {
				overrideConfig := overrideVMConfigList[0].(map[string]interface{})
				if powerState, exists := overrideConfig["power_state"]; exists && powerState.(string) == "ON" {
					log.Printf("[DEBUG] 🔌 Powering VM back on after hot-plug changes")
					if err := callForPowerOnVM(ctx, conn, d, meta); err != nil {
						return err
					}
				}
			}
		}
	}

	// Handle power state changes
	if d.HasChange("override_vm_config") {
		var oldPowerState, newPowerState string

		// Extract power state from old config
		if oldList, ok := oldConfig.([]interface{}); ok && len(oldList) > 0 {
			if oldMap, ok := oldList[0].(map[string]interface{}); ok {
				if ps, exists := oldMap["power_state"]; exists && ps != nil {
					oldPowerState = ps.(string)
				}
			}
		}

		// Extract power state from new config
		if newList, ok := newConfig.([]interface{}); ok && len(newList) > 0 {
			if newMap, ok := newList[0].(map[string]interface{}); ok {
				if ps, exists := newMap["power_state"]; exists && ps != nil {
					newPowerState = ps.(string)
				}
			}
		}

		// Handle power state change if it actually changed and we haven't already handled it
		if oldPowerState != newPowerState && newPowerState != "" && !hotPlugRequired {
			log.Printf("[DEBUG] ⚡ Handling power state change from '%s' to '%s'", oldPowerState, newPowerState)

			readResp, err := conn.VMAPIInstance.GetVmById(utils.StringPtr(d.Id()))
			if err != nil {
				return diag.Errorf("error reading VM for power state change: %v", err)
			}

			args := make(map[string]interface{})
			args["If-Match"] = getEtagHeader(readResp, conn)

			var powerResp interface{}
			var taskUUID *string

			switch newPowerState {
			case "ON":
				powerResp, err = conn.VMAPIInstance.PowerOnVm(utils.StringPtr(d.Id()), args)
				if err != nil {
					return diag.Errorf("error powering on VM: %v", err)
				}
				TaskRef := powerResp.(*import3.TaskReference)
				taskUUID = TaskRef.ExtId
			case "OFF":
				powerResp, err = conn.VMAPIInstance.PowerOffVm(utils.StringPtr(d.Id()), args)
				if err != nil {
					return diag.Errorf("error powering off VM: %v", err)
				}
				TaskRef := powerResp.(*import3.TaskReference)
				taskUUID = TaskRef.ExtId
			default:
				return diag.Errorf("unsupported power state: %s", newPowerState)
			}

			if taskUUID != nil {
				taskconn := meta.(*conns.Client).PrismAPI
				stateConf := &resource.StateChangeConf{
					Pending: []string{"QUEUED", "RUNNING"},
					Target:  []string{"SUCCEEDED"},
					Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
					Timeout: d.Timeout(schema.TimeoutUpdate),
				}

				if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
					return diag.Errorf("error waiting for power state change (%s): %s", utils.StringValue(taskUUID), errWaitTask)
				}
			}
			log.Printf("[DEBUG] ✅ Power state changed successfully to: %s", newPowerState)
		}
	}

	log.Printf("[DEBUG] ✅ OVA VM update completed successfully")
	return ResourceNutanixOvaVMDeploymentRead(ctx, d, meta)
}

// setOvaVMConfig sets the VM configuration in the Terraform state for OVA VMs
func setOvaVMConfig(d *schema.ResourceData, vm import2.Vm) diag.Diagnostics {
	log.Printf("[DEBUG] 📊 Setting OVA VM state from API response")

	// For OVA VMs, we want to be very conservative about state updates
	// The main issue we're solving is the trunked_vlans drift
	// We should preserve user configuration and only update specific fields that cause drift

	// Get existing override_vm_config to preserve all user configuration
	if v, ok := d.GetOk("override_vm_config"); ok {
		existingList := v.([]interface{})
		if len(existingList) > 0 {
			if existingConfig, ok := existingList[0].(map[string]interface{}); ok {
				// Create a copy of the existing configuration
				overrideConfig := make(map[string]interface{})
				for k, v := range existingConfig {
					overrideConfig[k] = v
				}

				// Update basic VM properties from API response only if they have meaningful values
				if vm.Name != nil {
					overrideConfig["name"] = utils.StringValue(vm.Name)
				}

				// Only update CPU settings if they're provided and non-zero
				// This prevents overwriting user config with API defaults or missing values
				if vm.NumSockets != nil && utils.IntValue(vm.NumSockets) > 0 {
					overrideConfig["num_sockets"] = utils.IntValue(vm.NumSockets)
					log.Printf("[DEBUG] 🖥️ Preserved num_sockets from API: %d", utils.IntValue(vm.NumSockets))
				} else {
					log.Printf("[DEBUG] ⚠️ API did not return num_sockets or returned 0, preserving user config")
				}

				if vm.NumCoresPerSocket != nil && utils.IntValue(vm.NumCoresPerSocket) > 0 {
					overrideConfig["num_cores_per_socket"] = utils.IntValue(vm.NumCoresPerSocket)
					log.Printf("[DEBUG] 🖥️ Preserved num_cores_per_socket from API: %d", utils.IntValue(vm.NumCoresPerSocket))
				} else {
					log.Printf("[DEBUG] ⚠️ API did not return num_cores_per_socket or returned 0, preserving user config")
				}

				if vm.NumThreadsPerCore != nil && utils.IntValue(vm.NumThreadsPerCore) > 0 {
					overrideConfig["num_threads_per_core"] = utils.IntValue(vm.NumThreadsPerCore)
					log.Printf("[DEBUG] 🖥️ Preserved num_threads_per_core from API: %d", utils.IntValue(vm.NumThreadsPerCore))
				} else {
					log.Printf("[DEBUG] ⚠️ API did not return num_threads_per_core or returned 0, preserving user config")
				}

				if vm.MemorySizeBytes != nil && utils.Int64Value(vm.MemorySizeBytes) > 0 {
					overrideConfig["memory_size_bytes"] = int(utils.Int64Value(vm.MemorySizeBytes))
				}
				if vm.PowerState != nil {
					overrideConfig["power_state"] = vm.PowerState.GetName()
					log.Printf("[DEBUG] ⚡ Set power_state: %s", vm.PowerState.GetName())
				}

				// Only update NICs to fix the trunked_vlans drift issue
				if len(vm.Nics) > 0 {
					nicsList := make([]interface{}, 0)

					for _, nic := range vm.Nics {
						nicMap := make(map[string]interface{})

						if nic.ExtId != nil {
							nicMap["ext_id"] = utils.StringValue(nic.ExtId)
						}

						// Preserve existing backing_info if it exists
						if existingNics, ok := existingConfig["nics"].([]interface{}); ok && len(existingNics) > len(nicsList) {
							if existingNic, ok := existingNics[len(nicsList)].(map[string]interface{}); ok {
								if existingBackingInfo, ok := existingNic["backing_info"]; ok {
									nicMap["backing_info"] = existingBackingInfo
								}
							}
						}

						if nic.NetworkInfo != nil {
							networkInfoList := make([]map[string]interface{}, 0)
							networkInfo := make(map[string]interface{})

							if nic.NetworkInfo.NicType != nil {
								networkInfo["nic_type"] = nic.NetworkInfo.NicType.GetName()
							}

							if nic.NetworkInfo.Subnet != nil && nic.NetworkInfo.Subnet.ExtId != nil {
								subnetList := make([]map[string]interface{}, 0)
								subnet := make(map[string]interface{})
								subnet["ext_id"] = utils.StringValue(nic.NetworkInfo.Subnet.ExtId)
								subnetList = append(subnetList, subnet)
								networkInfo["subnet"] = subnetList
							}

							if nic.NetworkInfo.VlanMode != nil {
								networkInfo["vlan_mode"] = nic.NetworkInfo.VlanMode.GetName()
							}

							// Handle trunked_vlans properly - this is the main fix for drift
							if len(nic.NetworkInfo.TrunkedVlans) > 0 {
								networkInfo["trunked_vlans"] = nic.NetworkInfo.TrunkedVlans
								log.Printf("[DEBUG] 🏷️ Setting trunked_vlans: %v", nic.NetworkInfo.TrunkedVlans)
							} else {
								// Set empty array if no trunked VLANs to prevent drift
								networkInfo["trunked_vlans"] = []int{}
								log.Printf("[DEBUG] 🏷️ Setting empty trunked_vlans to prevent drift")
							}

							if nic.NetworkInfo.ShouldAllowUnknownMacs != nil {
								networkInfo["should_allow_unknown_macs"] = utils.BoolValue(nic.NetworkInfo.ShouldAllowUnknownMacs)
							}

							networkInfoList = append(networkInfoList, networkInfo)
							nicMap["network_info"] = networkInfoList
						}

						nicsList = append(nicsList, nicMap)
					}

					// Update only the NICs configuration, preserving everything else
					overrideConfig["nics"] = nicsList
					log.Printf("[DEBUG] ✅ Updated NICs configuration with %d NICs", len(nicsList))
				}

				// Set the complete override_vm_config with preserved user settings
				overrideConfigList := []map[string]interface{}{overrideConfig}
				if err := d.Set("override_vm_config", overrideConfigList); err != nil {
					return diag.FromErr(fmt.Errorf("failed setting override_vm_config: %w", err))
				}
			}
		}
	}

	log.Printf("[DEBUG] ✅ OVA VM state set successfully (minimal update approach)")
	return nil
}

func ResourceNutanixOvaVMDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.Client).VmmAPI

	readResp, err := conn.VMAPIInstance.GetVmById(utils.StringPtr(d.Id()))
	if err != nil {
		return diag.Errorf("error while reading vm : %v", err)
	}
	// Extract E-Tag Header
	args := make(map[string]interface{})
	args["If-Match"] = getEtagHeader(readResp, conn)

	resp, err := conn.VMAPIInstance.DeleteVmById(utils.StringPtr(d.Id()), args)
	if err != nil {
		return diag.Errorf("error while deleting vm : %v", err)
	}
	TaskRef := resp.Data.GetValue().(import3.TaskReference)
	taskUUID := TaskRef.ExtId

	// calling group API to poll for completion of task

	taskconn := meta.(*conns.Client).PrismAPI
	// Wait for the VM to be available
	stateConf := &resource.StateChangeConf{
		Pending: []string{"QUEUED", "RUNNING"},
		Target:  []string{"SUCCEEDED"},
		Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
		return diag.Errorf("error waiting for vm (%s) to delete: %s", utils.StringValue(taskUUID), errWaitTask)
	}
	return nil
}

// Helper function to handle power state changes
func handlePowerStateChange(ctx context.Context, conn *conns.Client, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	targetPowerState := d.Get("power_state").(string)
	log.Printf("[DEBUG] ⚡ Changing power state to: %s", targetPowerState)

	readResp, err := conn.VmmAPI.VMAPIInstance.GetVmById(utils.StringPtr(d.Id()))
	if err != nil {
		return diag.Errorf("error reading VM for power state change: %v", err)
	}

	// Extract E-Tag Header
	args := make(map[string]interface{})
	args["If-Match"] = getEtagHeader(readResp, conn.VmmAPI)

	var powerResp interface{}
	var taskUUID *string

	switch targetPowerState {
	case "ON":
		powerResp, err = conn.VmmAPI.VMAPIInstance.PowerOnVm(utils.StringPtr(d.Id()), args)
		if err != nil {
			return diag.Errorf("error powering on VM: %v", err)
		}
		TaskRef := powerResp.(*import3.TaskReference)
		taskUUID = TaskRef.ExtId
	case "OFF":
		powerResp, err = conn.VmmAPI.VMAPIInstance.PowerOffVm(utils.StringPtr(d.Id()), args)
		if err != nil {
			return diag.Errorf("error powering off VM: %v", err)
		}
		TaskRef := powerResp.(*import3.TaskReference)
		taskUUID = TaskRef.ExtId
	default:
		return diag.Errorf("unsupported power state: %s", targetPowerState)
	}

	if taskUUID != nil {
		taskconn := meta.(*conns.Client).PrismAPI

		// Wait for power state change to complete
		stateConf := &resource.StateChangeConf{
			Pending: []string{"QUEUED", "RUNNING"},
			Target:  []string{"SUCCEEDED"},
			Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
			Timeout: d.Timeout(schema.TimeoutUpdate),
		}

		if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
			return diag.Errorf("error waiting for power state change (%s): %s", utils.StringValue(taskUUID), errWaitTask)
		}
	}

	log.Printf("[DEBUG] ✅ Power state changed successfully to: %s", targetPowerState)
	return nil
}
