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

package tiles

import (
	"omg-cli/config"
	"omg-cli/opsman"
)

// TileInstaller defines and configures an Ops Manager Tile.
type TileInstaller interface {

	// Definition creates a Tile definition.
	Definition(envConfig *config.EnvConfig) config.Tile

	// Configure applies configuration to a tile via the Ops Manager SDK.
	Configure(envConfig *config.EnvConfig, cfg *config.Config, om *opsman.Sdk) error

	// BuiltIn is true if a tile is built in to the Ops Manager.
	BuiltIn() bool
}
