---
subcategory: "Database"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_postgresql_flexible_server_firewall_rules"
description: |-
  Manages a group of PostgreSQL Flexible Server Firewall Rules.
---

# azurerm_postgresql_flexible_server_firewall_rules

Manages PostgreSQL Flexible Server Firewall Rules as a single resource.

~> **NOTE:** This resource is not compatible with the use of the `azurerm_postgresql_flexible_server_firewall_rule` resource. Firewall rules may be managed individually with that resource, or collectively with this one.

## Example Usage

```hcl
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_postgresql_flexible_server" "example" {
  name                   = "example-psqlflexibleserver"
  resource_group_name    = azurerm_resource_group.example.name
  location               = azurerm_resource_group.example.location
  version                = "12"
  administrator_login    = "psqladmin"
  administrator_password = "H@Sh1CoR3!"

  storage_mb = 32768

  sku_name = "GP_Standard_D4s_v3"
}

resource "azurerm_postgresql_flexible_server_firewall_rules" "example" {
  server_id = azurerm_postgresql_flexible_server.example.id

  rule {
    name             = "example1"
    start_ip_address = "10.0.0.1"
    end_ip_address   = "10.0.0.254"
  }

  rule {
    name             = "example2"
    start_ip_address = "10.1.0.1"
    end_ip_address   = "10.1.0.128"
  }
}
```

## Arguments Reference

The following arguments are supported:

* `server_id` - (Required) The ID of the PostgreSQL Flexible Server from which to create these PostgreSQL Flexible Server Firewall Rules. Changing this forces a new resource to be created.

* `rule` - (Required) One or more `rule` blocks as defined below.

---

The `rule` block supports the following:

* `name` - (Required) The name which should be used for this PostgreSQL Flexible Server Firewall Rule.

* `start_ip_address` - (Required) The Start IP Address associated with this PostgreSQL Flexible Server Firewall Rule.

* `end_ip_address` - (Required) The End IP Address associated with this PostgreSQL Flexible Server Firewall Rule.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/language/resources/syntax#operation-timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the PostgreSQL Flexible Server Firewall Rule.
* `read` - (Defaults to 5 minutes) Used when retrieving the PostgreSQL Flexible Server Firewall Rule.
* `update` - (Defaults to 30 minutes) Used when updating the PostgreSQL Flexible Server Firewall Rule.
* `delete` - (Defaults to 30 minutes) Used when deleting the PostgreSQL Flexible Server Firewall Rule.

```shell
terraform import azurerm_postgresql_flexible_server_firewall_rule.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.DBforPostgreSQL/flexibleServers/flexibleServer1
```
