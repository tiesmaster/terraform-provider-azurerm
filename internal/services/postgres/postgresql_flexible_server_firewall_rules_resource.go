// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/postgresql/2022-12-01/firewallrules"
	"github.com/hashicorp/go-azure-sdk/resource-manager/postgresql/2023-06-01-preview/servers"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/locks"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/postgres/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

const maxConcurrency = 100

type Rule struct {
	Name           string `tfschema:"name"`
	StartIPAddress string `tfschema:"start_ip_address"`
	EndIPAddress   string `tfschema:"end_ip_address"`
}

type FlexibleServerFirewallRulesModel struct {
	ServerID string `tfschema:"server_id"`
	Rule     []Rule `tfschema:"rule"`
}

var (
	_ sdk.Resource           = FlexibleServerFirewallRulesResource{}
	_ sdk.ResourceWithUpdate = FlexibleServerFirewallRulesResource{}
)

type FlexibleServerFirewallRulesResource struct{}

func (r FlexibleServerFirewallRulesResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"server_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: firewallrules.ValidateFlexibleServerID,
		},

		"rule": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validate.FlexibleServerFirewallRuleName,
					},

					"end_ip_address": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.IsIPAddress,
					},

					"start_ip_address": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.IsIPAddress,
					},
				},
			},
		},
	}
}

func (r FlexibleServerFirewallRulesResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r FlexibleServerFirewallRulesResource) ResourceType() string {
	return "azurerm_postgresql_flexible_server_firewall_rules"
}

func (r FlexibleServerFirewallRulesResource) ModelObject() interface{} {
	return &FlexibleServerFirewallRulesModel{}
}

func (r FlexibleServerFirewallRulesResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return servers.ValidateFlexibleServerID
}

func (r FlexibleServerFirewallRulesResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			rulesClient := metadata.Client.Postgres.FlexibleServerFirewallRuleClient

			model := FlexibleServerFirewallRulesModel{}
			if err := metadata.Decode(&model); err != nil {
				return err
			}

			id, err := firewallrules.ParseFlexibleServerID(model.ServerID)
			if err != nil {
				return err
			}

			locks.ByName(id.FlexibleServerName, postgresqlFlexibleServerResourceName)
			defer locks.UnlockByName(id.FlexibleServerName, postgresqlFlexibleServerResourceName)

			listFirewallRulesResult, err := rulesClient.ListByServerComplete(ctx, *id)
			if err != nil {
				return err
			}
			if len(listFirewallRulesResult.Items) != 0 {
				return tf.ImportAsExistsError(r.ResourceType(), id.ID())
			}

			firewallRules := make(map[string]firewallrules.FirewallRule)
			for _, rule := range model.Rule {
				fwRule := firewallrules.FirewallRule{
					Properties: firewallrules.FirewallRuleProperties{
						EndIPAddress:   rule.EndIPAddress,
						StartIPAddress: rule.StartIPAddress,
					},
				}
				fwRuleId := firewallrules.NewFirewallRuleID(id.SubscriptionId, id.ResourceGroupName, id.FlexibleServerName, rule.Name)
				firewallRules[fwRuleId.ID()] = fwRule
			}

			maxRulesAtOnce := make(chan struct{}, maxConcurrency)
			errs := make(chan error)
			wg := &sync.WaitGroup{}

			for i, f := range firewallRules {
				wg.Add(1)
				fid, _ := firewallrules.ParseFirewallRuleID(i)
				go batchCreateOrUpdate(ctx, rulesClient, *fid, f, wg, maxRulesAtOnce, errs)

			}

			go func() {
				wg.Wait()
				close(errs)
			}()

			for chanErr := range errs {
				if chanErr != nil {
					return chanErr
				}
			}

			wg.Wait()

			metadata.SetID(id)

			return nil
		},
	}
}

func (r FlexibleServerFirewallRulesResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Postgres.FlexibleServerFirewallRuleClient

			id, err := firewallrules.ParseFlexibleServerID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			state := FlexibleServerFirewallRulesModel{
				ServerID: firewallrules.NewFlexibleServerID(id.SubscriptionId, id.ResourceGroupName, id.FlexibleServerName).ID(),
			}

			fwRules, err := client.ListByServerComplete(ctx, *id)
			if err != nil {
				if response.WasNotFound(fwRules.LatestHttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			rules := make([]Rule, 0)
			for _, rule := range fwRules.Items {
				rules = append(rules, Rule{
					Name:           pointer.From(rule.Name),
					StartIPAddress: rule.Properties.StartIPAddress,
					EndIPAddress:   rule.Properties.EndIPAddress,
				})
			}

			state.Rule = rules

			return metadata.Encode(&state)
		},
	}
}

func (r FlexibleServerFirewallRulesResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			subscriptionId := metadata.Client.Account.SubscriptionId
			client := metadata.Client.Postgres.FlexibleServerFirewallRuleClient

			model := FlexibleServerFirewallRulesModel{}
			if err := metadata.Decode(&model); err != nil {
				return err
			}

			id, err := firewallrules.ParseFlexibleServerID(model.ServerID)
			if err != nil {
				return err
			}

			locks.ByName(id.FlexibleServerName, postgresqlFlexibleServerResourceName)
			defer locks.UnlockByName(id.FlexibleServerName, postgresqlFlexibleServerResourceName)

			if metadata.ResourceData.HasChange("firewall_rule") {
				listFirewallRulesResult, err := client.ListByServerComplete(ctx, *id)
				if err != nil {
					return err
				}
				currentFirewallRules := listFirewallRulesResult.Items

				firewallRules := make(map[string]firewallrules.FirewallRule)
				// Build a map of what the firewall rules should look like with the ID as the key
				for _, rule := range model.Rule {
					fwRule := firewallrules.FirewallRule{
						Properties: firewallrules.FirewallRuleProperties{
							EndIPAddress:   rule.EndIPAddress,
							StartIPAddress: rule.StartIPAddress,
						},
					}
					fwRuleId := firewallrules.NewFirewallRuleID(subscriptionId, id.ResourceGroupName, id.FlexibleServerName, rule.Name)
					firewallRules[fwRuleId.ID()] = fwRule
				}

				rulesToDelete := make([]firewallrules.FirewallRuleId, 0)

				// iterate over the received rules for ID matches for rules to remove.
				for _, v := range currentFirewallRules {
					if cId, err := firewallrules.ParseFirewallRuleIDInsensitively(pointer.From(v.Id)); err == nil {
						if _, ok := firewallRules[cId.ID()]; !ok {
							rulesToDelete = append(rulesToDelete, *cId)
						}
					}
				}

				// Delete removed rules first to avoid potential errors from overlapping ranges or renamed rules
				semaphore := make(chan struct{}, maxConcurrency)
				errs := make(chan error)
				wg := &sync.WaitGroup{}
				for _, f := range rulesToDelete {
					wg.Add(1)
					go batchDelete(ctx, client, f, wg, semaphore, errs)
				}

				go func() {
					wg.Wait()
					close(errs)
				}()

				for chanErr := range errs {
					if chanErr != nil {
						return chanErr
					}
				}

				wg.Wait()

				errs = make(chan error)

				// Add / update rules - Rules are governed by their name, so updates and creates do not need to be split here
				for i, f := range firewallRules {
					wg.Add(1)
					fid, _ := firewallrules.ParseFirewallRuleID(i)
					go batchCreateOrUpdate(ctx, client, *fid, f, wg, semaphore, errs)
				}

				go func() {
					wg.Wait()
					close(errs)
				}()

				for chanErr := range errs {
					if chanErr != nil {
						return chanErr
					}
				}

				wg.Wait()
			}

			return nil
		},
	}
}

func (r FlexibleServerFirewallRulesResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Postgres.FlexibleServerFirewallRuleClient

			id, err := firewallrules.ParseFlexibleServerID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			locks.ByName(id.FlexibleServerName, postgresqlFlexibleServerResourceName)
			defer locks.UnlockByName(id.FlexibleServerName, postgresqlFlexibleServerResourceName)

			listFirewallRulesResult, err := client.ListByServerComplete(ctx, *id)
			if err != nil {
				return err
			}

			maxRulesAtOnce := make(chan struct{}, maxConcurrency)
			errs := make(chan error)
			wg := &sync.WaitGroup{}
			for _, v := range listFirewallRulesResult.Items {
				ruleId, err := firewallrules.ParseFirewallRuleID(pointer.From(v.Id))
				if err != nil {
					return fmt.Errorf("deleting Firewall Rules %s: %+v", *id, err)
				}
				wg.Add(1)
				go batchDelete(ctx, client, *ruleId, wg, maxRulesAtOnce, errs)
			}

			go func() {
				wg.Wait()
				close(errs)
			}()

			for chanErr := range errs {
				if chanErr != nil {
					return chanErr
				}
			}

			wg.Wait()

			return nil
		},
	}
}

func batchCreateOrUpdate(ctx context.Context, client *firewallrules.FirewallRulesClient, id firewallrules.FirewallRuleId, rule firewallrules.FirewallRule, wg *sync.WaitGroup, semaphore chan struct{}, errs chan error) {
	defer wg.Done()
	semaphore <- struct{}{}
	errs <- client.CreateOrUpdateThenPoll(ctx, id, rule)
	<-semaphore
}

func batchDelete(ctx context.Context, client *firewallrules.FirewallRulesClient, id firewallrules.FirewallRuleId, wg *sync.WaitGroup, semaphore chan struct{}, errs chan error) {
	defer wg.Done()
	semaphore <- struct{}{}
	errs <- client.DeleteThenPoll(ctx, id)
	<-semaphore
}
