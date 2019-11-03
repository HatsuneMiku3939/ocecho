// Copyright 2017, ocecho Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package ocecho

import (
	"net/http"

	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

const (
	HostAttribute       = "http.host"
	MethodAttribute     = "http.method"
	PathAttribute       = "http.path"
	URLAttribute        = "http.url"
	UserAgentAttribute  = "http.user_agent"
	StatusCodeAttribute = "http.status_code"
)

type addedTagsKey struct{}

type addedTags struct {
	t []tag.Mutator
}

func requestAttrs(path string, r *http.Request) []trace.Attribute {
	userAgent := r.UserAgent()

	attrs := make([]trace.Attribute, 0, 5)
	attrs = append(attrs,
		trace.StringAttribute(PathAttribute, path),
		trace.StringAttribute(URLAttribute, r.URL.String()),
		trace.StringAttribute(HostAttribute, r.Host),
		trace.StringAttribute(MethodAttribute, r.Method),
	)

	if userAgent != "" {
		attrs = append(attrs, trace.StringAttribute(UserAgentAttribute, userAgent))
	}

	return attrs
}
