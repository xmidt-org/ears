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

package fragmentstorerfx

import (
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/db"
	"github.com/xmidt-org/ears/internal/pkg/db/dynamo"
	"github.com/xmidt-org/ears/pkg/fragments"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		ProvideFragmentStorer,
	),
)

type StorageIn struct {
	fx.In
	Config config.Config
	Logger *zerolog.Logger
}

type StorageOut struct {
	fx.Out
	FragmentStorer fragments.FragmentStorer
}

func ProvideFragmentStorer(in StorageIn) (StorageOut, error) {
	out := StorageOut{}
	storageType := in.Config.GetString("ears.storage.fragment.type")
	switch storageType {
	case "inmemory":
		out.FragmentStorer = db.NewInMemoryFragmentStorer(in.Config)
	case "dynamodb":
		fragmentStorer, err := dynamo.NewDynamoDbFragmentStorer(in.Config)
		if err != nil {
			return out, err
		}
		out.FragmentStorer = fragmentStorer
	default:
		return out, &UnsupportedFragmentStorageError{storageType}
	}
	return out, nil
}
