package product

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {

	//Class ProductCategory

	pool.ProductCategory().DeclareModel()
	pool.ProductCategory().AddCharField("Name", models.StringFieldParams{String: "Name", Index: true, Required: true, Translate: true})
	pool.ProductCategory().AddMany2OneField("Parent", models.ForeignKeyFieldParams{String: "Parent",
		RelationModel: pool.ProductCategory(), Index: true, OnDelete: models.Cascade})
	pool.ProductCategory().AddMany2OneField("Child", models.ForeignKeyFieldParams{String: "Child",
		RelationModel: pool.ProductCategory()})
	pool.ProductCategory().AddSelectionField("Type", models.SelectionFieldParams{String: "Type", Selection: types.Selection{
		"view":   "View",
		"normal": "Normal",
	}})
	pool.ProductCategory().AddIntegerField("ProductCount", models.SimpleFieldParams{String: "# Products"}) //, Compute: "compute_product_count"})

	//Class ProductPriceHistory

	pool.ProductPriceHistory().DeclareModel()
	pool.ProductPriceHistory().AddMany2OneField("CompanyId", models.ForeignKeyFieldParams{String: "Company", RelationModel: pool.Company(), Required: true})
	//, Default: "GetDefaultCompanyId", })
	pool.ProductPriceHistory().AddMany2OneField("ProductId", models.ForeignKeyFieldParams{String: "Product", RelationModel: pool.ProductProduct(),
		OnDelete: models.Cascade, Required: true})
	pool.ProductPriceHistory().AddDateTimeField("DateTime", models.SimpleFieldParams{String: "Date",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return types.Now()

		}})
	pool.ProductPriceHistory().AddFloatField("Cost", models.FloatFieldParams{String: "Cost"}) //, Digits: dp.getPrecision("Product Price")})

	//Class ProductProduct
	pool.ProductProduct().DeclareModel()
	pool.ProductProduct().InheritModel(pool.ProductTemplate())
	pool.ProductProduct().AddFloatField("Price", models.FloatFieldParams{String: "Price"}) //, Digits: dp.getPrecision("Product Price"),
	// inverse: "setProductPrice"})
	pool.ProductProduct().AddFloatField("PriceExtra", models.FloatFieldParams{String: "Variant Price Extra"}) //, Digits: dp.getPrecision("Product Price"),
	// inverse: "setProductLstPrice", Help: "This is the sum of the extra price of all attributes"})
	pool.ProductProduct().AddFloatField("LstPrice", models.FloatFieldParams{String: "Sale Price"}) //Compute: "ComputeProductLstPrice",
	// Digits: dp.getPrecision("Product Price"),
	//inverse: "setProductLstPrice", Help: "The sale price is managed from the product template."+
	// "Click on the 'Variant Prices' button to set the extra attribute prices."})
	pool.ProductProduct().AddCharField("DefaultCode", models.StringFieldParams{String: "Internal Reference", Index: true})
	pool.ProductProduct().AddCharField("Code", models.StringFieldParams{String: "Internal Reference"}) //, Compute: "ComputeProductCode"})
	pool.ProductProduct().AddCharField("PartnerRef", models.StringFieldParams{String: "Customer Ref"}) //, Compute: "ComputePartnerRef"})
	pool.ProductProduct().AddBooleanField("Active", models.SimpleFieldParams{String: "Active",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return true
		}, Help: "If unchecked, it will allow you to hide the product without removing it."})
	pool.ProductProduct().AddMany2OneField("ProductTmplId", models.ForeignKeyFieldParams{String: "Product Template",
		RelationModel: pool.ProductTemplate(), OnDelete: models.Cascade,
		Index: true, Required: true})
	pool.ProductProduct().AddCharField("Barcode", models.StringFieldParams{String: "Barcode", NoCopy: true,
		Help: "International Article Number used for product identification."})
	pool.ProductProduct().AddMany2ManyField("AttributeValueIds", models.Many2ManyFieldParams{String: "Attributes",
		RelationModel: pool.ProductAttributeValue()}) //, OnDelete: "restrict"})
	pool.ProductProduct().AddBinaryField("ImageVariant", models.SimpleFieldParams{String: "Variant Image",
		Help: "This field holds the image used as image for the product variant, limited to 1024x1024px."})
	pool.ProductProduct().AddBinaryField("ImageSmall", models.SimpleFieldParams{String: "Small-Sized Image",
		//, Compute: "ComputeImages", inverse: "setImageSmall",
		Help: "Image of the product variant (Small-sized image of product template if false)."})
	pool.ProductProduct().AddBinaryField("Image", models.SimpleFieldParams{String: "Big-sized Image", //, Compute: "ComputeImages", inverse: "setImage",
		Help: "Image of the product variant (Big-sized image of product template if false). " +
			"It is automatically resized as a 1024x1024px image, with aspect ratio preserved."})
	pool.ProductProduct().AddBinaryField("ImageMedium", models.SimpleFieldParams{String: "Medium-sized Image", //, Compute: "ComputeImages",
		// inverse: "setImageMedium",
		Help: "Image of the product variant (Medium-sized image of product template if false)."})
	pool.ProductProduct().AddFloatField("StandardPrice", models.FloatFieldParams{String: "Cost", //, company_dependent: True,
		// Digits: dp.get_precision("Product Price"),
		GroupOperator: "base.GroupUser", Help: "Cost of the product template used for standard stock valuation in accounting and used as a base price " +
			"on purchase orders Expressed in the default unit of measure of the product."})
	pool.ProductProduct().AddFloatField("Volume", models.FloatFieldParams{String: "Volume", Help: "The volume in m3."})
	pool.ProductProduct().AddFloatField("Weight", models.FloatFieldParams{String: "Weight", //, Digits: dp.getPrecision("Stock Weight")
		Help: "The weight of the contents in Kg, not including any packaging, etc."})
	pool.ProductProduct().AddMany2ManyField("PricelistItemIds", models.Many2ManyFieldParams{String: "Pricelist Items",
		RelationModel: pool.ProductPriceListItem()}) //, Compute: "getPricelistItems"})
	pool.ProductProduct().AddSQLConstraint("BarcodeUniq", "unique(barcode)", ("A barcode can only be assigned to one product !"))

	//Class ProductPackaging

	pool.ProductPackaging().DeclareModel()
	pool.ProductPackaging().AddCharField("Name", models.StringFieldParams{String: "Name", Required: true})
	pool.ProductPackaging().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 1
		}, Help: "The first in the sequence is the default one."})
	pool.ProductPackaging().AddMany2OneField("ProductTmplId", models.ForeignKeyFieldParams{String: "Product", RelationModel: pool.ProductTemplate()})
	pool.ProductPackaging().AddFloatField("Qty", models.FloatFieldParams{String: "Quantity Per Package",
		Help: "The total number of products you can have per pallet or box."})

	//Class ProductSupplierInfo

	pool.ProductSupplierInfo().DeclareModel()
	pool.ProductSupplierInfo().AddMany2OneField("Name", models.ForeignKeyFieldParams{String: "Vendor", RelationModel: pool.Partner(),
		//Filter:pool.Partner().Supplier().Equals(),
		OnDelete: models.Cascade, Required: true, Help: "Vendor of this product"})
	pool.ProductSupplierInfo().AddCharField("ProductName", models.StringFieldParams{String: "Vendor Product Name",
		Help: "This vendor's product name will be used when printing a request for quotation. Keep empty to use the internal one."})
	pool.ProductSupplierInfo().AddCharField("ProductCode", models.StringFieldParams{String: "Vendor Product Code",
		Help: "This vendor's product name will be used when printing a request for quotation. Keep empty to use the internal one."})
	pool.ProductSupplierInfo().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 1
		}, Help: "Assigns the priority to the list of product vendor."})
	pool.ProductSupplierInfo().AddMany2OneField("ProductUom", models.ForeignKeyFieldParams{String: "Vendor Unit of Measure", RelationModel: pool.ProductUoM(),
		Related: "ProductTmplId.UomPoId", Help: "This comes from the product form."})
	// ProductUom ReadOnly
	pool.ProductSupplierInfo().Fields().ProductUom().RevokeAccess(security.GroupEveryone, security.Write)
	pool.ProductSupplierInfo().AddFloatField("MinQty", models.FloatFieldParams{String: "Minimal Quantity",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 0.0
		}, Required: true, Help: "The minimal quantity to purchase from this vendor, expressed in the vendor Product Unit of Measure if not any, " +
			"in the default unit of measure of the product otherwise."})
	pool.ProductSupplierInfo().AddFloatField("Price", models.FloatFieldParams{String: "Price",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 0.0
		}, Required: true, Help: "The price to purchase a product"}) //, Digits: dp.getPrecision("Product Price")})
	pool.ProductSupplierInfo().AddMany2OneField("CompanyId", models.ForeignKeyFieldParams{String: "Company", RelationModel: pool.Company(), Index: true,
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 1 //pool.User().Company().Id()
		}})
	pool.ProductSupplierInfo().AddMany2OneField("CurrencyId", models.ForeignKeyFieldParams{String: "Currency", RelationModel: pool.Currency(),
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 1 //pool.User().Company().Currency().Id()
		}, Required: true})
	pool.ProductSupplierInfo().AddDateField("DateStart", models.SimpleFieldParams{String: "Start Date", Help: "Start date for this vendor price"})
	pool.ProductSupplierInfo().AddDateField("DateEnd", models.SimpleFieldParams{String: "End Date", Help: "End date for this vendor price"})
	pool.ProductSupplierInfo().AddMany2OneField("ProductId", models.ForeignKeyFieldParams{String: "Product Variant", RelationModel: pool.ProductProduct(),
		Help: "When this field is filled in, the vendor data will only apply to the variant."})
	pool.ProductSupplierInfo().AddMany2OneField("ProductTmplId", models.ForeignKeyFieldParams{String: "Product Template", RelationModel: pool.ProductTemplate(),
		Index: true, OnDelete: models.Cascade})
	pool.ProductSupplierInfo().AddIntegerField("Delay", models.SimpleFieldParams{String: "Delivery Lead Time",
		Default: func(models.Environment, models.FieldMap) interface{} {
			return 1
		}, Required: true, Help: "Lead time in days between the confirmation of the purchase order and the receipt of the products in your warehouse. " +
			"Used by the scheduler for automatic computation of the purchase order planning."})

}
