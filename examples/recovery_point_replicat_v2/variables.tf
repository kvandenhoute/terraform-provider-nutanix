#define the type of variables to be used in terraform file
variable "nutanix_username" {
  type = string
}
variable "nutanix_password" {
  type = string
}
variable "nutanix_endpoint" {
  type = string
}
variable "nutanix_port" {
  type = string
}

variable "remote_cluster_uuid" {
  type = string
}
variable "remote_pc_uuid" {
  type = string
}
