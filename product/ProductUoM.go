package product

import (
	"github.com/hexya-erp/hexya/pool"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/security"
)

func init() {

	//class ProductUomCategory
	pool.ProductUomCategory().DeclareModel()
	pool.ProductUomCategory().AddCharField("Name" , models.StringFieldParams{String:"Name" , Required:true,Translate:true})

	//class ProductUom
	pool.ProductUom().DeclareModel()
	pool.ProductUom().AddCharField("Name" , models.StringFieldParams{String:"Unit of Measure" , Required:true,Translate:true})
	pool.ProductUom().AddMany2OneField("CategoryId" , models.ForeignKeyFieldParams{String:"Category" , RelationModel:pool.ProductUomCategory(), Required:true, OnDelete:models.Cascade ,
	Help:"Conversion between Units of Measure can only occur if they belong to the same category. The conversion will be made based on the ratios."})
	// force NUMERIC with unlimited precision
	pool.ProductUom().AddFloatField("Factor" , models.FloatFieldParams{String:"Ratio", Required:true, Default: func(models.Environment, models.FieldMap) interface{} {
		return 1.0
	} , Digits:types.Digits{0,0},Help:"How much bigger or smaller this unit is compared to the reference Unit of Measure for this category: 1 * (reference unit) = ratio * (this unit)"})
	// force NUMERIC with unlimited precision
	pool.ProductUom().AddFloatField("FactorInv" , models.FloatFieldParams{String:"Bigger Ratio" , Required:true, Digits:types.Digits{0,0} ,
	Help:"How many times this Unit of Measure is bigger than the reference Unit of Measure in this category: 1 * (this unit) = ratio * (reference unit)"}) //, Compute:"ComputeFactorInv"})
	//ReadOnly FactorInv
	pool.ProductUom().Fields().FactorInv().RevokeAccess(security.GroupEveryone,security.Write)
	pool.ProductUom().AddFloatField("Rounding" , models.FloatFieldParams{String:"Rounding Precision" , Required:true, Default: func(models.Environment, models.FieldMap) interface{} {
		return 0.01
	} ,Digits:types.Digits{0,0} ,
		Help:"The computed quantity will be a multiple of this value. Use 1.0 for a Unit of Measure that cannot be further split, such as a piece."})
	pool.ProductUom().AddBooleanField("Active", models.SimpleFieldParams{String:"Active" , Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	} , Help:"Uncheck the active field to disable a unit of measure without deleting it."})
	pool.ProductUom().AddSelectionField("UomType", models.SelectionFieldParams{String: "Type", Selection: types.Selection{
		"bigger" : "Bigger than the reference Unit of Measure",
		"reference": "Reference Unit of Measure for this category",
		"smaller": "Smaller than the reference Unit of Measure",
	} , Default: func(models.Environment, models.FieldMap) interface{} {
		return "reference"
	} , Required:true})
	pool.ProductUom().AddSQLConstraint("FactorGtZero", "CHECK (factor!=0)", ("The conversion ratio for a unit of measure cannot be 0!"))
	pool.ProductUom().AddSQLConstraint("RoundingGtZero", "CHECK (rounding>0)", ("The rounding precision must be greater than 0!"))
}
