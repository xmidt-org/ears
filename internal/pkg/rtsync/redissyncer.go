// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rtsync

import (
	"context"
	"errors"
	"github.com/go-redis/redis"
	"os"
	"strings"
	"time"
)

type (
	RedisTableSyncer struct {
		client        *redis.Client
		redisEndpoint string
		subscribers   map[string]int64
	}
)

//TODO: add tests
//TODO: remove ping logic
//TODO: integrate with table manager
//TODO: load entire table on launch
//TODO: logging
//TODO: integrate with uberfx
//TODO: extract configs (redis endpoint,
//TODO: start loading from dynamodb
//TODO: review notification payload (json)
//TODO: introduce global table hash
//TODO: consider app id and org id

func NewRedisTableSyncer(ctx context.Context, env string, endpoint string) TableSyncer {
	rep := "gears-redis-qa-001.6bteey.0001.usw2.cache.amazonaws.com:6379"
	if env != "" {
		rep = strings.Replace(rep, "-qa-", "-"+env+"-", -1)
	}
	if endpoint != "" {
		rep = endpoint
	}
	s := new(RedisTableSyncer)
	//ctx.Log.Info("op", "NewRedisCompLibSyncer", "action", "start_service", "endpoint", rep)
	s.redisEndpoint = rep
	s.client = redis.NewClient(&redis.Options{
		Addr:     rep,
		Password: "",
		DB:       0,
	})
	s.subscribers = make(map[string]int64)
	return s
}

// PublishSyncRequest asks others to sync their routing tables

func (s *RedisTableSyncer) PublishSyncRequest(ctx context.Context, routeId string, add bool) error {
	// listen for ACKs first ...
	go func() {
		numSubscribers := s.GetInstanceCount(ctx)
		received := make(map[string]bool, 0)
		lrc := redis.NewClient(&redis.Options{
			Addr:     s.redisEndpoint,
			Password: "",
			DB:       0,
		})
		defer lrc.Close()
		pubsub := lrc.Subscribe(EARS_REDIS_ACK_CHANNEL)
		defer pubsub.Close()
		// 30 sec timeout on collecting acks
		done := make(chan bool, 1)
		go func() {
			for {
				msg, err := pubsub.ReceiveMessage()
				if err != nil {
					//ctx.Log.Error("op", "PublishSyncRequest", "action", "receive_ack_message", "channel", FSC_COMPLIB_ACK_CHANNEL, "msg", msg.Payload, "error", err.Error())
					break
				} else {
					//ctx.Log.Info("op", "PublishSyncRequest", "action", "receive_ack_message", "channel", FSC_COMPLIB_ACK_CHANNEL, "msg", msg.Payload)
				}
				received[msg.Payload] = true
				// wait until we received an ack from each subscriber (except the one originating the request)
				if len(received) >= numSubscribers-1 {
					break
				}
			}
			done <- true
		}()
		select {
		case <-done:
			//ctx.Log.Info("op", "PublishSyncRequest", "action", "done_collecting_acks", "num_received", len(received), "num_expected", (numSubscribers - 1))
		case <-time.After(30 * time.Second):
			//ctx.Log.Error("op", "PublishSyncRequest", "action", "done_collecting_acks", "num_received", len(received), "num_expected", (numSubscribers - 1))
		}
		// inform flow editor
		//debugSessionMgr.set(ctx, COMPONENT_LIBRARY_TS, time.Now().Unix())
		s.PublishMutationMessage(ctx, routeId, add)
	}()
	// ... then request all flow apis to sync
	hostname, _ := os.Hostname()
	msg := ""
	if add {
		msg = EARS_REDIS_ADD_ROUTE_CMD
	} else {
		msg = EARS_REDIS_REMOVE_ROUTE_CMD
	}
	msg += "," + routeId + "," + hostname
	err := s.client.Publish(EARS_REDIS_SYNC_CHANNEL, msg).Err()
	if err != nil {
		//ctx.Log.Error("op", "PublishSyncRequest", "action", "publish_sync_request_message", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg, "err", err.Error())
		return err
	} else {
		//ctx.Log.Info("op", "PublishSyncRequest", "action", "publish_sync_request_message", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg)
	}
	return nil
}

// ListenForSyncRequests listens for sync request

func (s *RedisTableSyncer) ListenForSyncRequests(ctx context.Context) {
	go func() {
		lrc := redis.NewClient(&redis.Options{
			Addr:     s.redisEndpoint,
			Password: "",
			DB:       0,
		})
		defer lrc.Close()
		pubsub := lrc.Subscribe(EARS_REDIS_SYNC_CHANNEL)
		defer pubsub.Close()
		//ctx.Log.Info("op", "ListenForSyncRequests", "action", "start")
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				//ctx.Log.Error("op", "ListenForSyncRequests", "action", "receive_sync_message", "channel", FSC_COMPLIB_SYNC_CHANNEL, "error", err.Error())
				time.Sleep(EARS_REDIS_RETRY_INTERVAL_SECONDS)
			} else {
				//ctx.Log.Info("op", "ListenForSyncRequests", "action", "receive_sync_message", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg.Payload)
				// parse message
				elems := strings.Split(msg.Payload, ",")
				if len(elems) != 3 {
					//ctx.Log.Error("op", "ListenForSyncRequests", "action", "receive_sync_message", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg.Payload, "error", "bad message structure")
				}
				// sync only whats needed
				hostname, _ := os.Hostname()
				if elems[2] != hostname {
					if elems[0] == EARS_REDIS_ADD_ROUTE_CMD {
						//ctx.Log.Info("op", "ListenForSyncRequests", "action", "sync_added_component", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg.Payload)
						err = s.SyncRouteAdded(ctx, elems[1])
						//ctx.Log.Info("op", "ListenForSyncRequests", "action", "sync_added_component_done", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg.Payload)
					} else if elems[0] == EARS_REDIS_REMOVE_ROUTE_CMD {
						//ctx.Log.Info("op", "ListenForSyncRequests", "action", "sync_removed_component", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg.Payload)
						err = s.SyncRouteRemoved(ctx, elems[1])
						//ctx.Log.Info("op", "ListenForSyncRequests", "action", "sync_removed_component_done", "channel", FSC_COMPLIB_SYNC_CHANNEL, "msg", msg.Payload)
					} else {
						err = errors.New("bad command " + elems[0])
					}
					if err != nil {
						//ctx.Log.Error("op", "ListenForSyncRequests", "action", "sync_with_s3", "err", err.Error())
					} else {
						//ctx.Log.Info("op", "ListenForSyncRequests", "action", "sync_with_s3")
					}
				} else {
					//ctx.Log.Info("op", "ListenForSyncRequests", "action", "no_sync_required")
				}
			}
			/*err = s.SyncAllRoutes(ctx)
			if err != nil {
				ctx.Log.Error("action", "comp_lib_sync", "op", "reload_comp_lib_mgr", "err", err.Error())
			}*/
			s.PublishAckMessage(ctx)
		}
	}()
}

// PublishAckMessage confirm successful syncing of routing table

func (s *RedisTableSyncer) PublishAckMessage(ctx context.Context) error {
	hostname, _ := os.Hostname()
	err := s.client.Publish(EARS_REDIS_ACK_CHANNEL, hostname).Err()
	if err != nil {
		//ctx.Log.Error("op", "PublishAckMessage", "action", "publish_ack_message", "channel", FSC_COMPLIB_ACK_CHANNEL, "msg", hostname, "error", err.Error())
	} else {
		//ctx.Log.Info("op", "PublishAckMessage", "action", "publish_ack_message", "channel", FSC_COMPLIB_ACK_CHANNEL, "msg", hostname)
	}
	return err
}

// PublishPings publish heart beat messages

func (s *RedisTableSyncer) PublishPings(ctx context.Context) {
	hostname, _ := os.Hostname()
	//ctx.Log.Info("op", "PublishPings", "action", "start")
	go func() {
		for {
			err := s.client.Publish(EARS_REDIS_PING_CHANNEL, hostname).Err()
			if err != nil {
				//ctx.Log.Error("op", "PublishPings", "action", "publish_ping_message", "channel", FSC_COMPLIB_PING_CHANNEL, "msg", hostname, "error", err.Error())
			} else {
				//ctx.Log.Info("op", "PublishPings", "action", "publish_ping_message", "channel", FSC_COMPLIB_PING_CHANNEL, "msg", hostname)
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

// ListenForPingMessages listens for heart beat messages

func (s *RedisTableSyncer) ListenForPingMessages(ctx context.Context) {
	//ctx.Log.Info("op", "ListenForPingMessages", "action", "start")
	go func() {
		lrc := redis.NewClient(&redis.Options{
			Addr:     s.redisEndpoint,
			Password: "",
			DB:       0,
		})
		defer lrc.Close()
		pubsub := lrc.Subscribe(EARS_REDIS_PING_CHANNEL)
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				//ctx.Log.Error("op", "ListenForPingMessages", "action", "receive_ping_message", "channel", FSC_COMPLIB_PING_CHANNEL, "error", err.Error())
				time.Sleep(EARS_REDIS_RETRY_INTERVAL_SECONDS)
			} else {
				nowsec := time.Now().Unix()
				if msg.Payload != "" {
					s.subscribers[msg.Payload] = nowsec
				}
				// remove dead subscribers from list
				for sid, ts := range s.subscribers {
					if nowsec-ts > 120 {
						delete(s.subscribers, sid)
					}
				}
				//ctx.Log.Info("op", "ListenForPingMessages", "action", "receive_ping_message", "channel", FSC_COMPLIB_PING_CHANNEL, "msg", msg.Payload, "current_subscriber_count", len(s.subscribers))
			}
		}
	}()
}

// GetSubscriberCount gets number of live ears instances

func (s *RedisTableSyncer) GetInstanceCount(ctx context.Context) int {
	return len(s.subscribers)
}

// PublishMutationMessage lets other interested parties know that updates are available

func (s *RedisTableSyncer) PublishMutationMessage(ctx context.Context, routeId string, add bool) error {
	hostname, _ := os.Hostname()
	msg := ""
	if add {
		msg = EARS_REDIS_ADD_ROUTE_CMD
	} else {
		msg = EARS_REDIS_REMOVE_ROUTE_CMD
	}
	msg += "," + routeId + "," + hostname
	//ctx.Log.Info("op", "PublishNewVersionAvailableMessage", "action", "publish_new_version_available", "channel", FSC_COMPLIB_NEW_VERSION_AVAILABLE_CHANNEL, "msg", msg)
	return s.client.Publish(EARS_REDIS_MUTATION_CHANNEL, msg).Err()
}

// SyncRouteAdded

func (s *RedisTableSyncer) SyncRouteAdded(ctx context.Context, routeId string) error {
	//TODO
	return errors.New("not implemented")
}

// SyncRouteRemoved

func (s *RedisTableSyncer) SyncRouteRemoved(ctx context.Context, routeId string) error {
	//TODO
	return errors.New("not implemented")
}

// SyncAllRoutes

func (s *RedisTableSyncer) SyncAllRoutes(ctx context.Context) error {
	//TODO
	return errors.New("not implemented")
}
