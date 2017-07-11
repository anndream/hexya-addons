package Wizard

import (
	"github.com/hexya-erp/hexya/pool"
	"github.com/hexya-erp/hexya/hexya/models"
)

func init() {

	pool.WizProductPriceList().DeclareTransientModel()
	pool.WizProductPriceList().AddMany2OneField("PriceList" , models.ForeignKeyFieldParams{String:"PriceList" , RelationModel:pool.ProductPriceList() ,
	Required:true})
	pool.WizProductPriceList().AddIntegerField("Qty1", models.SimpleFieldParams{String:"Quantity-1",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 1
	}})
	pool.WizProductPriceList().AddIntegerField("Qty2", models.SimpleFieldParams{String:"Quantity-2",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 5
		}})
	pool.WizProductPriceList().AddIntegerField("Qty3", models.SimpleFieldParams{String:"Quantity-3",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 10
		}})
	pool.WizProductPriceList().AddIntegerField("Qty4", models.SimpleFieldParams{String:"Quantity-4",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 0
		}})
	pool.WizProductPriceList().AddIntegerField("Qty5", models.SimpleFieldParams{String:"Quantity-5",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 0
		}})
}
