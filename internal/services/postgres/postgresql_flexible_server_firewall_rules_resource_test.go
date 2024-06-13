// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-sdk/resource-manager/postgresql/2022-12-01/firewallrules"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type PostgresqlFlexibleServerFirewallRulesResource struct{}

func TestAccPostgresqlFlexibleServerFirewallRules_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_postgresql_flexible_server_firewall_rules", "test")
	r := PostgresqlFlexibleServerFirewallRulesResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPostgresqlFlexibleServerFirewallRules_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_postgresql_flexible_server_firewall_rules", "test")
	r := PostgresqlFlexibleServerFirewallRulesResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (PostgresqlFlexibleServerFirewallRulesResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := firewallrules.ParseFlexibleServerID(state.ID)
	if err != nil {
		return nil, err
	}

	rules, err := clients.Postgres.FlexibleServerFirewallRuleClient.ListByServerComplete(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}

	return pointer.To(rules.Items != nil), nil
}

func (PostgresqlFlexibleServerFirewallRulesResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_postgresql_flexible_server_firewall_rules" "test" {
  server_id = azurerm_postgresql_flexible_server.test.id
  rule {
    name             = "acctest-FSFR-%[2]d"
    start_ip_address = "120.0.0.0"
    end_ip_address   = "120.0.0.0"
  }
  rule {
    name             = "acctest-FSFR2-%[2]d"
    start_ip_address = "121.0.0.0"
    end_ip_address   = "121.0.0.0"
  }
  rule {
    name             = "acctest-FSFR3-%[2]d"
    start_ip_address = "122.0.0.0"
    end_ip_address   = "122.0.0.0"
  }
  rule {
    name             = "acctest-FSFR4-%[2]d"
    start_ip_address = "123.0.0.0"
    end_ip_address   = "123.0.0.0"
  }
}
`, PostgresqlFlexibleServerResource{}.basic(data), data.RandomInteger)
}

func (r PostgresqlFlexibleServerFirewallRulesResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_postgresql_flexible_server_firewall_rules" "test" {
  server_id = azurerm_postgresql_flexible_server.test.id
  rule {
    name             = "acctest-FSFR-%[2]d"
    start_ip_address = "124.0.0.0"
    end_ip_address   = "124.0.0.254"
  }
  rule {
    name             = "acctest-FSFR2-%[2]d"
    start_ip_address = "125.0.0.0"
    end_ip_address   = "125.0.0.254"
  }
  rule {
    name             = "acctest-FSFR5-%[2]d"
    start_ip_address = "11.0.0.0"
    end_ip_address   = "11.1.0.254"
  }
}
`, PostgresqlFlexibleServerResource{}.basic(data), data.RandomInteger)
}
