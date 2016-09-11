// Copyright 2016 NDP Systèmes. All Rights Reserved.
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

package methods

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/models"
)

// GetFilters returns the filters for the given model and actionID for the current user
func GetFilters(rs *models.RecordCollection, modelName, actionID string) []*pool.IrFilters {
	var res []*pool.IrFilters
	actRef := ir.MakeActionRef(actionID)
	rs.Filter("Model", "=", modelName).Filter("ActionId", "=", actRef.String()).Filter("User.ID", "=", rs.Env().Uid()).ReadAll(&res)
	return res
}

func initFilters() {
	models.CreateMethod("IrFilters", "GetFilters", GetFilters)
}