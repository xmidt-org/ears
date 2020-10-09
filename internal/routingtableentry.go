/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type (
	// A RoutingTableEntry represents an entry in the EARS routing table
	RoutingTableEntry struct {
		OrgId        string              `json:"orgId"`        // org ID for quota and rate limiting
		AppId        string              `json:"appId"`        // app ID for quota and rate limiting
		UserId       string              `json:"userId"`       // user ID / author of route
		Source       *IOPlugin           `json:"source`        // pointer to source plugin instance
		Destination  *IOPlugin           `json:"destination"`  // pointer to destination plugin instance
		FilterChain  *FilterChain        `json:"filterChain"`  // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
		DeliveryMode string              `json:"deliveryMode"` // possible values: fire_and_forget, at_least_once, exactly_once
		Debug        bool                `json:"debug"`        // if true generate debug logs and metrics for events taking this route
		Ts           int                 `json:"ts"`           // timestamp when route was created or updated
		tblMgr       RoutingTableManager `json:"-"`            // pointer to routing table manager
	}
)

func (rte *RoutingTableEntry) Hash(ctx context.Context) string {
	str := rte.Source.Hash(ctx) + rte.Destination.Hash(ctx)
	if rte.FilterChain != nil && rte.FilterChain.Filters != nil {
		for _, filter := range rte.FilterChain.Filters {
			str += filter.Hash(ctx)
		}
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

func (rte *RoutingTableEntry) Validate(ctx context.Context) error {
	if rte.Source == nil {
		return new(MissingSourcePluginConfiguraton)
	}
	if rte.Destination == nil {
		return new(MissingDestinationPluginConfiguraton)
	}
	rte.Source.Mode = PluginModeInput
	rte.Destination.Mode = PluginModeOutput
	return nil
}

func (rte *RoutingTableEntry) String() string {
	buf, _ := json.Marshal(rte)
	return string(buf)
}

func (rte *RoutingTableEntry) Initialize(ctx context.Context) error {
	var err error
	//
	// initialize input plugin
	//
	rte.Source, err = GetIOPluginManager(ctx).RegisterRoute(ctx, rte, rte.Source)
	if err != nil {
		return err
	}
	//
	// initialize filter chain
	//
	if rte.FilterChain == nil {
		rte.FilterChain = NewFilterChain(ctx)
	}
	err = rte.FilterChain.Initialize(ctx, rte)
	if err != nil {
		return err
	}
	//
	// initialize output plugin
	//
	rte.Destination, err = GetIOPluginManager(ctx).RegisterRoute(ctx, rte, rte.Destination)
	if err != nil {
		return err
	}
	return nil
}

func (rte *RoutingTableEntry) Withdraw(ctx context.Context) error {
	// withdraw from input plugin
	var err error
	err = GetIOPluginManager(ctx).WithdrawRoute(ctx, rte, rte.Source)
	if err != nil {
		return err
	}
	// withdraw filter chain
	if rte.FilterChain != nil {
		rte.FilterChain.Withdraw(ctx)
	}
	// withdraw from output plugin
	err = GetIOPluginManager(ctx).WithdrawRoute(ctx, rte, rte.Destination)
	if err != nil {
		return err
	}
	return nil
}
