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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xmidt-org/ears/internal"

	"github.com/rs/zerolog/log"
)

var (
	ROUTE = `
	{
		"org_id" : "comcast",
		"app_id" : "xfi",
		"user_id" : "boris",
		"src_type" : "debug",
		"src_params" :
		{
			"rounds" : 10,
			"interval_ms" : 1000,
			"payload" : {
				"foo" : "bar"
			}
		},
		"dst_type" : "debug",
		"dst_params" : {},
		"filter_chain" : [
			{
				"type" : "match",
				"params" : {
					"pattern" : {
						"foo" : "bar"
					}
				}
			},
			{
				"type" : "filter",
				"params" : {
					"pattern" : {
						"hello" : "world"
					}
				}
			},
			{
				"type" : "split",
				"params" : {}
			},
			{
				"type" : "transform",
				"params" : {}
			}
		],
		"delivery_mode" : "at_least_once"
	}
	`
)

func main() {
	ctx := context.Background()
	var rtmgr internal.RoutingTableManager
	rtmgr = internal.NewInMemoryRoutingTableManager()
	log.Debug().Msg(fmt.Sprintf("ears has %d routes", rtmgr.GetRouteCount(ctx)))
	var rte internal.RoutingTableEntry
	err := json.Unmarshal([]byte(ROUTE), &rte)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	//buf, _ := json.MarshalIndent(rte, "", "\t")
	//fmt.Printf("%s\n", string(buf))
	err = rtmgr.AddRoute(ctx, &rte)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	log.Debug().Msg(fmt.Sprintf("ears has %d routes", rtmgr.GetRouteCount(ctx)))
	allRoutes, err := rtmgr.GetAllRoutes(ctx)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	if len(allRoutes) > 0 {
		log.Debug().Msg(fmt.Sprintf("first route has hash %s", allRoutes[0].Hash(ctx)))
	}
	time.Sleep(time.Duration(60) * time.Second)
}
