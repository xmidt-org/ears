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

	"github.com/xmidt-org/ears/internal/pkg/route"

	"github.com/rs/zerolog/log"
)

var (
	ROUTE_1 = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 3,
				"intervalMS" : 250,
				"payload" : {
					"foo" : "bar"
				}
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"filterChain" : {
			"filters": 
			[
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
			]
		},
		"deliveryMode" : "at_least_once"
	}
	`
	ROUTE_2 = `
	{
		"orgId" : "comcast",
		"appId" : "xfi",
		"userId" : "boris",
		"source" : {
			"type" : "debug",
			"params" :
			{
				"rounds" : 3,
				"intervalMS" : 250,
				"payload" : {
					"foo" : "bar"
				}
			}
		},
		"destination" : {
			"type" : "debug",
			"params" : {}
		},
		"deliveryMode" : "at_least_once"
	}
	`
)

func main() {
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	var rtmgr route.RoutingTableManager
	// init in memory routing table manager
	rtmgr = route.NewInMemoryRoutingTableManager()
	log.Ctx(ctx).Debug().Msg(fmt.Sprintf("ears has %d routes", rtmgr.GetRouteCount(ctx)))
	var rc route.RouteConfig
	err := json.Unmarshal([]byte(ROUTE_1), &rc)
	if err != nil {
		log.Ctx(ctx).Error().Msg(err.Error())
		return
	}
	buf, err := json.MarshalIndent(rc, "", "\t")
	if err != nil {
		log.Ctx(ctx).Error().Msg(err.Error())
		return
	}
	fmt.Printf("%s\n", string(buf))
	// add a route
	err = rtmgr.AddRoute(ctx, route.NewRouteFromRouteConfig(&rc))
	if err != nil {
		log.Ctx(ctx).Error().Msg(err.Error())
		return
	}
	log.Ctx(ctx).Debug().Msg(fmt.Sprintf("ears has %d routes", rtmgr.GetRouteCount(ctx)))
	// check route
	allRoutes, err := rtmgr.GetAllRoutes(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Msg(err.Error())
		return
	}
	if len(allRoutes) > 0 {
		log.Ctx(ctx).Debug().Msg(fmt.Sprintf("first route has hash %s", allRoutes[0].Hash(ctx)))
	}
	time.Sleep(time.Duration(3) * time.Second)
}
