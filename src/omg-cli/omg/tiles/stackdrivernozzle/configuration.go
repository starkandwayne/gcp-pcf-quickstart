/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stackdrivernozzle

import (
	"encoding/json"
	"fmt"

	"omg-cli/config"
	"omg-cli/omg/tiles"
	"omg-cli/opsman"
)

const (
	skipSSLValidation = "true"
)

type properties struct {
	Endpoint          tiles.Value `json:".properties.firehose_endpoint"`
	SkipSSLValidation tiles.Value `json:".properties.firehose_skip_ssl"`
	ServiceAccount    tiles.Value `json:".properties.service_account"`
	ProjectID         tiles.Value `json:".properties.project_id"`
}

type resources struct {
	StackdriverNozzle tiles.Resource `json:"stackdriver-nozzle"`
}

// Configure satisfies TileInstaller interface.
func (t *Tile) Configure(envConfig *config.EnvConfig, cfg *config.Config, om *opsman.Sdk) error {
	if err := om.StageProduct(tile.Product); err != nil {
		return err
	}

	network := tiles.NetworkConfig(cfg.ServicesSubnetName, cfg)

	networkBytes, err := json.Marshal(&network)
	if err != nil {
		return err
	}

	properties := &properties{
		Endpoint:          tiles.Value{Value: fmt.Sprintf("https://api.sys.%s", cfg.DNSSuffix)},
		SkipSSLValidation: tiles.Value{Value: skipSSLValidation},
		ServiceAccount:    tiles.Value{Value: cfg.StackdriverNozzleServiceAccountKey},
		ProjectID:         tiles.Value{Value: cfg.ProjectName},
	}

	propertiesBytes, err := json.Marshal(&properties)
	if err != nil {
		return err
	}

	vmType := ""
	if envConfig.SmallFootprint {
		vmType = "micro"
	}
	resources := resources{
		StackdriverNozzle: tiles.Resource{
			InternetConnected: false,
			VMTypeID:          vmType,
		},
	}
	resourcesBytes, err := json.Marshal(&resources)
	if err != nil {
		return err
	}

	return om.ConfigureProduct(tile.Product.Name, string(networkBytes), string(propertiesBytes), string(resourcesBytes))
}
