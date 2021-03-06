// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package account

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {

	pool.AccountCommonAccountReport().DeclareMixinModel()
	pool.AccountCommonAccountReport().InheritModel(pool.AccountCommonReport())

	pool.AccountCommonAccountReport().AddFields(map[string]models.FieldDefinition{
		"DisplayAccount": models.SelectionField{String: "Display Accounts", Selection: types.Selection{
			"all":      "All",
			"movement": "With movements",
			"not_zero": "With balance is not equal to 0",
			/*[('all','All'  ('movement','With movements'  ('not_zero','With balance is not equal to 0' ]*/}, /*[]*/ Required: true, Default: models.DefaultValue("movement")},
	})
	pool.AccountCommonAccountReport().Methods().PrePrintReport().DeclareMethod(
		`PrePrintReport`,
		func(rs pool.AccountCommonAccountReportSet, args struct {
			Data interface{}
		}) {
			//@api.multi
			/*def pre_print_report(self, data):
			  data['form'].update(self.read(['display_account'])[0])
			  return data
			*/
		})

}
