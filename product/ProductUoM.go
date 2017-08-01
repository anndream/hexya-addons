package product

import (
	"github.com/hexya-erp/hexya/pool"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/security"
)

func init() {

	//class ProductUoMCategory
	pool.ProductUoMCategory().DeclareModel()
	pool.ProductUoMCategory().AddCharField("Name" , models.StringFieldParams{String:"Name" , Required:true,Translate:true})

	//class ProductUoM
	pool.ProductUoM().DeclareModel()
	pool.ProductUoM().AddCharField("Name" , models.StringFieldParams{String:"Unit of Measure" , Required:true,Translate:true})
	pool.ProductUoM().AddMany2OneField("CategoryId" , models.ForeignKeyFieldParams{String:"Category" , RelationModel:pool.ProductUoMCategory(),
		Required:true, OnDelete:models.Cascade ,
	Help:"Conversion between Units of Measure can only occur if they belong to the same category. The conversion will be made based on the ratios."})
	// force NUMERIC with unlimited precision
	pool.ProductUoM().AddFloatField("Factor" , models.FloatFieldParams{String:"Ratio", Required:true,
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 1.0
	} , Digits:types.Digits{0,0},Help:"How much bigger or smaller this unit is compared to the reference Unit of Measure for this category:" +
			" 1 * (reference unit) = ratio * (this unit)"})
	// force NUMERIC with unlimited precision
	pool.ProductUoM().AddFloatField("FactorInv" , models.FloatFieldParams{String:"Bigger Ratio" , Required:true, Digits:types.Digits{0,0} ,
	Help:"How many times this Unit of Measure is bigger than the reference Unit of Measure in this category: 1 * (this unit) = ratio * (reference unit)"})
	//, Compute:"ComputeFactorInv"})
	//ReadOnly FactorInv
	pool.ProductUoM().Fields().FactorInv().RevokeAccess(security.GroupEveryone,security.Write)
	pool.ProductUoM().AddFloatField("Rounding" , models.FloatFieldParams{String:"Rounding Precision" , Required:true,
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 0.01
	} ,Digits:types.Digits{0,0} ,
		Help:"The computed quantity will be a multiple of this value. Use 1.0 for a Unit of Measure that cannot be further split, such as a piece."})
	pool.ProductUoM().AddBooleanField("Active", models.SimpleFieldParams{String:"Active" ,
		Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	} , Help:"Uncheck the active field to disable a unit of measure without deleting it."})
	pool.ProductUoM().AddSelectionField("UoMType", models.SelectionFieldParams{String: "Type", Selection: types.Selection{
		"bigger" : "Bigger than the reference Unit of Measure",
		"reference": "Reference Unit of Measure for this category",
		"smaller": "Smaller than the reference Unit of Measure",
	} , Default: func(models.Environment, models.FieldMap) interface{} {
		return "reference"
	} , Required:true , JSON:"uom_type"})
	pool.ProductUoM().AddSQLConstraint("FactorGtZero", "CHECK (factor!=0)", ("The conversion ratio for a unit of measure cannot be 0!"))
	pool.ProductUoM().AddSQLConstraint("RoundingGtZero", "CHECK (rounding>0)", ("The rounding precision must be greater than 0!"))
}
