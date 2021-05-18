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

package syncer_test

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"testing"
)

func InMemoryConfig() config.Config {
	v := viper.New()
	v.Set("ears.synchronization.active", true)
	return v
}

func newInMemoryDeltaSyncer() syncer.DeltaSyncer {
	return syncer.NewInMemoryDeltaSyncer(&log.Logger, InMemoryConfig())
}

func TestInMemoryDeltaSyncer(t *testing.T) {
	testSyncers(newInMemoryDeltaSyncer, t)
}
