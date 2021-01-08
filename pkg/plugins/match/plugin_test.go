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

package match_test

import (
	"testing"

	"github.com/xmidt-org/ears/pkg/plugins/match"

	. "github.com/onsi/gomega"
)

func TestNewPluginer(t *testing.T) {

	a := NewWithT(t)

	plugin := match.NewPlugin()
	a.Expect(plugin).ToNot(BeNil())

	a.Expect(plugin.Name()).To(Equal(match.Name))
	a.Expect(plugin.Version()).ToNot(Equal(""))
	a.Expect(plugin.Config()).To(Equal(""))

	// TODO: This is fragile, but fine for now.  It will be more of an issue
	//   when the plugin needs to be persisted and as the hasher is improved
	a.Expect(plugin.PluginerHash(nil)).To(Equal("60046f14c917c18a9a0f923e191ba0dc"))
	a.Expect(plugin.FiltererHash(nil)).To(Equal("60046f14c917c18a9a0f923e191ba0dc"))

	{
		plug, err := plugin.NewPluginer(nil)
		a.Expect(plug).ToNot(BeNil())
		a.Expect(err).To(BeNil())
		a.Expect(plug.Name()).To(Equal(match.Name))
		a.Expect(plug.Version()).ToNot(Equal(""))
		a.Expect(plug.Config()).To(Equal(""))
	}

	{
		plug, err := plugin.NewPluginer("myconfig")
		a.Expect(plug).ToNot(BeNil())
		a.Expect(err).To(BeNil())
		a.Expect(plug.Name()).To(Equal(match.Name))
		a.Expect(plug.Version()).ToNot(Equal(""))
		a.Expect(plug.Config()).To(Equal("myconfig"))
	}

	{
		plug, err := plugin.NewPluginer(`} bad yaml`)
		a.Expect(plug).ToNot(BeNil())
		a.Expect(err).To(BeNil())
		a.Expect(plug.Name()).To(Equal(match.Name))
		a.Expect(plug.Version()).ToNot(Equal(""))
		a.Expect(plug.Config()).To(Equal(`"} bad yaml"`))
	}

}
