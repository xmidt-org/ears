// Copyright 2021 Comcast Cable Communications Management, LLC
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

package bolt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/xmidt-org/ears/pkg/route"
)

type BoltDbStorer struct {
	fileName string
	db       *bolt.DB
}

type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

//TODO: configuration
//TODO: clean shutdown of bolt db

func NewBoltDbStorer(config Config) (*BoltDbStorer, error) {
	bdb := &BoltDbStorer{
		fileName: "bolt.dat",
	}
	var err error
	bdb.db, err = bolt.Open(bdb.fileName, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}
	err = bdb.db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("ROUTES"))
		if err != nil {
			return fmt.Errorf("could not create routes bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	fmt.Printf("Started Bolt DB")
	return bdb, nil
}

func (d *BoltDbStorer) GetRoute(ctx context.Context, id string) (route.Config, error) {
	route := route.Config{}
	err := d.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket([]byte("ROUTES")).Get([]byte(id))
		if buf == nil {
			return fmt.Errorf("no data in bolt for key " + id)
		}
		err := json.Unmarshal(buf, &route)
		if err != nil {
			return err
		}
		return nil
	})
	return route, err
}

func (d *BoltDbStorer) GetAllRoutes(ctx context.Context) ([]route.Config, error) {
	routes := make([]route.Config, 0)
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("ROUTES"))
		b.ForEach(func(k, v []byte) error {
			var route route.Config
			err := json.Unmarshal(v, &route)
			if err != nil {
				return err
			}
			routes = append(routes, route)
			return nil
		})
		return nil
	})
	return routes, err
}

func (d *BoltDbStorer) SetRoute(ctx context.Context, r route.Config) error {
	if r.Id == "" {
		return fmt.Errorf("no route to store in bolt")
	}
	val, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = d.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("ROUTES")).Put([]byte(r.Id), val)
		if err != nil {
			return fmt.Errorf("could not insert route: %v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *BoltDbStorer) SetRoutes(ctx context.Context, routes []route.Config) error {
	if routes == nil {
		return fmt.Errorf("no routes to store in bolt")
	}
	for _, r := range routes {
		err := d.SetRoute(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *BoltDbStorer) DeleteRoute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("no route to delete in bolt")
	}
	err := d.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("ROUTES")).Delete([]byte(id))
		if err != nil {
			return fmt.Errorf("could not insert route: %v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *BoltDbStorer) DeleteRoutes(ctx context.Context, ids []string) error {
	for _, id := range ids {
		err := d.DeleteRoute(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}
