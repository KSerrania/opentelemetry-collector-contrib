// Copyright 2020, OpenTelemetry Authors
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

package splunkhecreceiver

import (
	"fmt"
	"net"
	"strconv"

	"github.com/gobwas/glob"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk"
)

// Config defines configuration for the Splunk HEC receiver.
type Config struct {
	config.ReceiverSettings       `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	confighttp.HTTPServerSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	splunk.AccessTokenPassthroughConfig `mapstructure:",squash"`
	// Path we will listen on, defaults to `*` (anything matches)
	Path     string `mapstructure:"path"`
	pathGlob glob.Glob
	// HecToOtelAttrs creates a mapping from HEC metadata to attributes.
	HecToOtelAttrs splunk.HecToOtelAttrs `mapstructure:"hec_metadata_to_otel_attrs"`
}

// initialize the configuration
func (c *Config) initialize() error {
	if c.HecToOtelAttrs.Source == "" {
		c.HecToOtelAttrs.Source = splunk.DefaultSourceLabel
	}
	if c.HecToOtelAttrs.SourceType == "" {
		c.HecToOtelAttrs.SourceType = splunk.DefaultSourceTypeLabel
	}
	if c.HecToOtelAttrs.Index == "" {
		c.HecToOtelAttrs.Index = splunk.DefaultIndexLabel
	}
	if c.HecToOtelAttrs.Host == "" {
		c.HecToOtelAttrs.Host = conventions.AttributeHostName
	}

	path := c.Path
	if path == "" {
		path = "*"
	}
	glob, err := glob.Compile(path)
	if err != nil {
		return err
	}
	c.pathGlob = glob
	_, err = extractPortFromEndpoint(c.Endpoint)
	return err
}

// extract the port number from string in "address:port" format. If the
// port number cannot be extracted returns an error.
func extractPortFromEndpoint(endpoint string) (int, error) {
	_, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return 0, fmt.Errorf("endpoint is not formatted correctly: %s", err.Error())
	}
	port, err := strconv.ParseInt(portStr, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("endpoint port is not a number: %s", err.Error())
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port number must be between 1 and 65535")
	}
	return int(port), nil
}
