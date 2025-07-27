// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
