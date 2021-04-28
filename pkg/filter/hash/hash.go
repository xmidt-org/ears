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

package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"hash/fnv"
	"strings"
)

func NewFilter(config interface{}) (*Filter, error) {
	cfg, err := NewConfig(config)
	if err != nil {
		return nil, &filter.InvalidConfigError{
			Err: err,
		}
	}
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	f := &Filter{
		config: *cfg,
	}
	return f, nil
}

func (f *Filter) Filter(evt event.Event) []event.Event {
	//TODO: add validation logic to filter
	if f == nil {
		evt.Nack(&filter.InvalidConfigError{
			Err: fmt.Errorf("<nil> pointer filter"),
		})
		return nil
	}
	obj := evt.Payload()
	var parent interface{}
	var key string
	if f.config.HashPath == "" {
	} else if f.config.HashPath == "." {
	} else {
		path := strings.Split(f.config.HashPath, ".")
		for _, p := range path {
			var ok bool
			parent = obj
			key = p
			obj, ok = obj.(map[string]interface{})[p]
			if !ok {
				evt.Nack(errors.New("invalid object at hash path " + f.config.HashPath))
				return []event.Event{}
			}
		}
	}
	if obj == nil {
		evt.Nack(errors.New("nil object at hash path " + f.config.HashPath))
		return []event.Event{}
	}
	buf, err := json.Marshal(obj)
	if err != nil {
		evt.Nack(err)
		return []event.Event{}
	}
	var output interface{}
	switch f.config.HashAlgorithm {
	case "fnv":
		h := fnv.New32a()
		h.Write(buf)
		fnvHash := int(h.Sum32())
		if *f.config.Mod > 0 {
			fnvHash = fnvHash % (*f.config.Mod)
		}
		output = fnvHash
		break
	case "md5":
		md5Hash := md5.Sum(buf)
		output = md5Hash[:]
		break
	case "sha1":
		sha1Hash := sha1.Sum(buf)
		output = sha1Hash[:]
		break
	case "sha256":
		sha256Hash := sha256.Sum256(buf)
		output = sha256Hash[:]
		break
	case "hmac-md5":
		if f.config.Key == "" {
			evt.Nack(errors.New("key required for hmac"))
			return []event.Event{}
		}
		h := hmac.New(md5.New, []byte(f.config.Key))
		h.Write(buf)
		md5Hash := h.Sum(nil)
		output = md5Hash[:]
		break
	case "hmac-sha1":
		if f.config.Key == "" {
			evt.Nack(errors.New("key required for hmac"))
			return []event.Event{}
		}
		h := hmac.New(sha1.New, []byte(f.config.Key))
		h.Write(buf)
		sha1Hash := h.Sum(nil)
		output = sha1Hash[:]
		break
	case "hmac-sha256":
		if f.config.Key == "" {
			evt.Nack(errors.New("key required for hmac"))
			return []event.Event{}
		}
		h := hmac.New(sha256.New, []byte(f.config.Key))
		h.Write(buf)
		sha256Hash := h.Sum(nil)
		output = sha256Hash[:]
		break
	default:
		evt.Nack(errors.New("unsupported hashing algorithm " + f.config.HashAlgorithm))
		return []event.Event{}
	}
	if f.config.Encoding == "base64" {
		output = base64.StdEncoding.EncodeToString(output.([]byte))
	} else if f.config.Encoding == "hex" {
		output = hex.EncodeToString(output.([]byte))
	}
	// adjust key and parent if we have a result path
	if f.config.ResultPath != "" {
		if *f.config.Metadata {
			if evt.Metadata() == nil {
				evt.SetMetadata(make(map[string]interface{}))
				obj = evt.Metadata()
			}
		} else {
			obj = evt.Payload()
		}
		path := strings.Split(f.config.ResultPath, ".")
		for _, p := range path {
			var ok bool
			parent = obj
			key = p
			obj, ok = obj.(map[string]interface{})[p]
			if !ok {
				evt.Nack(errors.New("invalid object at result path " + f.config.ResultPath))
				return []event.Event{}
			}
		}
	}
	// insert in result in desired place
	if key != "" {
		parentMap, is := parent.(map[string]interface{})
		if !is {
			evt.Nack(errors.New("parent is not a map at decode path " + f.config.HashPath))
			return []event.Event{}
		}
		parentMap[key] = output
	} else {
		evt.SetPayload(output)
	}
	return []event.Event{evt}
}

func (f *Filter) Config() Config {
	if f == nil {
		return Config{}
	}
	return f.config
}
