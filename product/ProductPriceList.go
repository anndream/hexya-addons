package product

import (
	"github.com/hexya-erp/hexya/pool"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/security"
)

func init() {

	//class ProductPriceList
	pool.ProductPriceList().DeclareModel()
	pool.ProductPriceList().AddCharField("Name", models.StringFieldParams{String: "PriceListName", Required: true, Translate: true})
	pool.ProductPriceList().AddBooleanField("Active", models.SimpleFieldParams{String: "Active",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	}, Help: "If unchecked, it will allow you to hide the PriceList without removing it."})
	pool.ProductPriceList().AddOne2ManyField("ItemIds", models.ReverseFieldParams{String: "PriceList Items", RelationModel: pool.ProductPriceListItem(),
		ReverseFK: "PriceListId", NoCopy: false}) //, Default:"getDefaultItemIds"})
	pool.ProductPriceList().AddMany2OneField("CurrencyId", models.ForeignKeyFieldParams{String: "Currency", RelationModel: pool.Currency(),
		Required: true}) //,Default:"getDefaultCurrencyId" })
	pool.ProductPriceList().AddMany2OneField("CompanyId", models.ForeignKeyFieldParams{String: "Company", RelationModel: pool.Company()})
	pool.ProductPriceList().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 16
	}})
	pool.ProductPriceList().AddMany2ManyField("CountryGroupsIds", models.Many2ManyFieldParams{String: "Country Groups", RelationModel: pool.CountryGroup()})

	//class CountryGroup
	pool.CountryGroup().AddMany2ManyField("PriceListIds" , models.Many2ManyFieldParams{String:"PriceLists" , RelationModel:pool.ProductPriceList()})

	//class PriceListItem
	pool.ProductPriceListItem().DeclareModel()
	pool.ProductPriceListItem().AddMany2OneField("ProductTmplId", models.ForeignKeyFieldParams{String: "ProductTmplId", RelationModel: pool.ProductTemplate(),
		OnDelete: models.Cascade , Help:"Specify a template if this rule only applies to one product template. Keep empty otherwise."})
	pool.ProductPriceListItem().AddMany2OneField("ProductId", models.ForeignKeyFieldParams{String: "Product", RelationModel: pool.ProductProduct(),
		OnDelete: models.Cascade , Help:"Specify a product if this rule only applies to one product. Keep empty otherwise."})
	pool.ProductPriceListItem().AddMany2OneField("CategId", models.ForeignKeyFieldParams{String: "CategId", RelationModel: pool.ProductPriceList(),
		OnDelete: models.Cascade , Help:"Specify a product category if this rule only applies to products belonging to this category or its children categories. " +
			"Keep empty otherwise."})
	pool.ProductPriceListItem().AddIntegerField("MinQuantity", models.SimpleFieldParams{String: "Min. Quantity",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 1
	},
		Help: "For the rule to apply, bought/sold quantity must be greater than or equal to the minimum quantity specified in this field.\n" +
			"Expressed in the default unit of measure of the product."})
	pool.ProductPriceListItem().AddSelectionField("AppliedOn", models.SelectionFieldParams{String: "Apply On", Selection: types.Selection{
		"3_global":           "Global",
		"2_product_category": " Product Category",
		"1_product":          "Product",
		"0_product_variant":  "Product Variant",
	}, Default: func(models.Environment, models.FieldMap) interface{} {
		return "3_global"
	}, Required: true, Help: "PriceList Item applicable on selected option"})
	pool.ProductPriceListItem().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence", Required:true ,
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 5
	} ,Help:"Gives the order in which the PriceList items will be checked. The evaluation gives highest priority to lowest sequence and stops as soon as " +
			"a matching item is found."})
	pool.ProductPriceListItem().AddSelectionField("Base", models.SelectionFieldParams{String: "Base On", Selection: types.Selection{
		"list_price": "Public Price",
		"standard_price": "Cost",
		"PriceList": "Other PriceList",
	}, Default: func(models.Environment, models.FieldMap) interface{} {
		return "list_price"
	}, Required: true, Help: "Base price for computation.\nPublic Price: The base price will be the Sale/public Price.\n " +
		"Cost Price : The base price will be the cost price.\n Other PriceList : Computation of the base price based on another PriceList."})
	pool.ProductPriceListItem().AddMany2OneField("PriceListId", models.ForeignKeyFieldParams{String: "PriceList",
		RelationModel: pool.ProductPriceList(), Index: true,
		OnDelete: models.Cascade})
	pool.ProductPriceListItem().AddFloatField("PriceSurcharge", models.FloatFieldParams{String: "Price Surcharge",
	Help:"Specify the fixed amount to add or substract(if negative) to the amount calculated with the discount."}) //, Digits:get_precision('Product Price')})
	pool.ProductPriceListItem().AddFloatField("PriceDiscount", models.FloatFieldParams{String: "Price Discount", Digits:types.Digits{16,2},
	Default: func(models.Environment, models.FieldMap) interface{} {
		return 0
	}})
	pool.ProductPriceListItem().AddFloatField("PriceRound", models.FloatFieldParams{String: "Price Rounding",
		Help:"Sets the price so that it is a multiple of this value.\n Rounding is applied after the discount and before the surcharge.\n"+
		"To have prices that end in 9.99, set rounding 10, surcharge -0.01"}) //, Digits:get_precision('Product Price')})
	 pool.ProductPriceListItem().AddFloatField("PriceMinMargin", models.FloatFieldParams{String: "Min. Price Margin",
		Help:"Specify the minimum amount of margin over the base price."}) //, Digits:get_precision('Product Price')})
	pool.ProductPriceListItem().AddFloatField("PriceMaxMargin", models.FloatFieldParams{String: "Max. Price Margin",
		Help:"Specify the maximum amount of margin over the base price."}) //, Digits:get_precision('Product Price')})
	pool.ProductPriceListItem().AddMany2OneField("CompanyId", models.ForeignKeyFieldParams{String: "Company", RelationModel: pool.Company(),
		Required:true , Stored:true , Related:"PriceListId.CompanyId" })
	//CompanyId ReadOnly
	pool.ProductPriceListItem().Fields().CompanyId().RevokeAccess(security.GroupEveryone,security.Write)
	pool.ProductPriceListItem().AddMany2OneField("CurrencyId", models.ForeignKeyFieldParams{String: "Currency", RelationModel: pool.Currency() ,
		Stored:true  , Related:"PriceListId.CurrencyId"})
	//CurrencyId ReadOnly
	pool.ProductPriceListItem().Fields().CurrencyId().RevokeAccess(security.GroupEveryone,security.Write)
	pool.ProductPriceListItem().AddDateField("DateStart", models.SimpleFieldParams{String: "Start Date",
		Help: "Starting date for the PriceList item validation"})
	pool.ProductPriceListItem().AddDateField("DateEnd", models.SimpleFieldParams{String: "End Date", Help: "Ending valid for the PriceList item validation"})
	pool.ProductPriceListItem().AddSelectionField("ComputePrice", models.SelectionFieldParams{String: "ComputePrice", Selection: types.Selection{
		"fixed": "Fix Price",
		"percentage": "Percentage (discount)",
		"formula": "Formula",
	}, Default: func(models.Environment, models.FieldMap) interface{} {
		return "fixed"
	}, Index:true })
	pool.ProductPriceListItem().AddFloatField("FixedPrice", models.FloatFieldParams{String: "Fixed Price"}) //,  Digits:dp.get_precision("Product Price")} )
	pool.ProductPriceListItem().AddFloatField("PercentPrice" , models.FloatFieldParams{String:"Percentage Price"})
	//functional fields used for usability purposes
	pool.ProductPriceListItem().AddCharField("Name" , models.StringFieldParams{String:"Name", Help:"Explicit rule name for this PriceList line."})
		//, Compute:"getPriceListItemNamePrice"})
	pool.ProductPriceListItem().AddCharField("Price" , models.StringFieldParams{String:"Price" , Help:"Explicit rule name for this PriceList line."})
		//, Compute:"getPriceListItemNamePrice"})
}
