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

package route

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type (
	// A Route represents an entry in the EARS routing table
	Route struct {
		OrgId        string              `json:"orgId"`        // org ID for quota and rate limiting
		AppId        string              `json:"appId"`        // app ID for quota and rate limiting
		UserId       string              `json:"userId"`       // user ID / author of route
		Name         string              `json:"name"`         // optional unique name for route
		Source       Pluginer            `json:"source`        // pointer to source plugin instance
		Destination  Pluginer            `json:"destination"`  // pointer to destination plugin instance
		FilterChain  *FilterChain        `json:"filterChain"`  // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
		DeliveryMode string              `json:"deliveryMode"` // possible values: fire_and_forget, at_least_once, exactly_once
		Debug        bool                `json:"debug"`        // if true generate debug logs and metrics for events taking this route
		Ts           int                 `json:"ts"`           // timestamp when route was created or updated
		tblMgr       RoutingTableManager `json:"-"`            // pointer to routing table manager
	}
	RouteConfig struct {
		OrgId        string       `json:"orgId"`        // org ID for quota and rate limiting
		AppId        string       `json:"appId"`        // app ID for quota and rate limiting
		UserId       string       `json:"userId"`       // user ID / author of route
		Name         string       `json:"name"`         // optional unique name for route
		Source       *Plugin      `json:"source`        // pointer to source plugin instance
		Destination  *Plugin      `json:"destination"`  // pointer to destination plugin instance
		FilterChain  *FilterChain `json:"filterChain"`  // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
		DeliveryMode string       `json:"deliveryMode"` // possible values: fire_and_forget, at_least_once, exactly_once
		Debug        bool         `json:"debug"`        // if true generate debug logs and metrics for events taking this route
		Ts           int          `json:"ts"`           // timestamp when route was created or updated
	}
)

func NewRouteFromRouteConfig(rc *RouteConfig) *Route {
	re := Route{
		OrgId:        rc.OrgId,
		AppId:        rc.AppId,
		UserId:       rc.UserId,
		Name:         rc.Name,
		Source:       rc.Source,
		Destination:  rc.Destination,
		FilterChain:  rc.FilterChain,
		DeliveryMode: rc.DeliveryMode,
		Debug:        rc.Debug,
		Ts:           rc.Ts,
	}
	return &re
}

func (rte *Route) Hash(ctx context.Context) string {
	str := rte.Source.Hash(ctx) + rte.Destination.Hash(ctx)
	if rte.FilterChain != nil && rte.FilterChain.Filters != nil {
		for _, filter := range rte.FilterChain.Filters {
			str += filter.Hash(ctx)
		}
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return hash
}

func (rte *Route) Validate(ctx context.Context) error {
	if rte.Source == nil {
		return &MissingPluginConfiguratonError{"", "", PluginModeInput}
	}
	if rte.Destination == nil {
		return &MissingPluginConfiguratonError{"", "", PluginModeOutput}
	}
	rte.Source.GetConfig().Mode = PluginModeInput
	rte.Destination.GetConfig().Mode = PluginModeOutput
	return nil
}

func (rte *Route) String() string {
	buf, _ := json.Marshal(rte)
	return string(buf)
}

func (rte *Route) Initialize(ctx context.Context) error {
	var err error
	//
	// initialize source plugin
	//
	rte.Source, err = GetIOPluginManager(ctx).RegisterPlugin(ctx, rte, rte.Source)
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
	// initialize destination plugin
	//
	rte.Destination, err = GetIOPluginManager(ctx).RegisterPlugin(ctx, rte, rte.Destination)
	if err != nil {
		return err
	}
	return nil
}

func (rte *Route) Withdraw(ctx context.Context) error {
	// withdraw from destination plugin
	var err error
	err = GetIOPluginManager(ctx).WithdrawPlugin(ctx, rte, rte.Destination)
	if err != nil {
		return err
	}
	// withdraw filter chain
	if rte.FilterChain != nil {
		rte.FilterChain.Withdraw(ctx)
	}
	// withdraw from source plugin
	err = GetIOPluginManager(ctx).WithdrawPlugin(ctx, rte, rte.Source)
	if err != nil {
		return err
	}
	return nil
}
