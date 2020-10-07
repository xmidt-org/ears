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
	"fmt"
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
	return nil
}

func (rte *RoutingTableEntry) Initialize(ctx context.Context) error {
	//
	// initialize input plugin
	//
	var err error
	rte.Source, err = NewInputPlugin(ctx, rte)
	if err != nil {
		return err
	}
	//
	// initialize filter chain
	//
	err = rte.FilterChain.Initialize(ctx, rte)
	if err != nil {
		return err
	}
	//
	// initialize output plugin
	//
	rte.Destination, err = NewOutputPlugin(ctx, rte)
	if err != nil {
		return err
	}
	return nil
}
