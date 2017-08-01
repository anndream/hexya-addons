package product

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {

	pool.ProductTemplate().DeclareModel()
	pool.ProductTemplate().AddCharField("Name", models.StringFieldParams{String: "Name", Index: true, Required: true, Translate: true})
	pool.ProductTemplate().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 1
	}, Help: "Gives the sequence order when displaying a product list"})
	pool.ProductTemplate().AddTextField("Description", models.StringFieldParams{String: "Description", Translate: true,
		Help: "A precise description of the Product, used only for internal information purposes."})
	pool.ProductTemplate().AddTextField("DescriptionPurchase", models.StringFieldParams{String: "Purchase Description", Translate: true,
		Help: "A description of the Product that you want to communicate to your vendors. This description will be copied to every Purchase Order, " +
			"Receipt and Vendor Bill/Refund."})
	pool.ProductTemplate().AddTextField("DescriptionSale", models.StringFieldParams{String: "Sale Description", Translate: true,
		Help: "A description of the Product that you want to communicate to your customers .This description will be copied to every Sale Order, " +
			"Delivery Order and Customer Invoice/Refund."})
	pool.ProductTemplate().AddSelectionField("Type", models.SelectionFieldParams{String: "Type", Selection: types.Selection{
		"consu":   "Consumable",
		"service": "Service",
	},
		Help: "A stockable product is a product for which you manage stock. The 'Inventory' app has to be installed.\n" +
			"A consumable product, on the other hand, is a product for which stock is not managed.\n" +
			"A service is a non-material product you provide.\n" +
			"A digital content is a non-material product you sell online. The files attached to the products are the one that are sold on " +
			"the e-commerce such as e-books, music, pictures,... The 'Digital Product' module has to be installed."})
	pool.ProductTemplate().AddBooleanField("Rental", models.SimpleFieldParams{String: "Can Be Rent"})
	pool.ProductTemplate().AddMany2OneField("CategId", models.ForeignKeyFieldParams{String: "Internal Category", RelationModel: pool.ProductCategory(),
		Required: true, Help: "Select category for the current product"}) //, ChangeDefault:true ,Default:"getDefaultCategoryId"  ,  Filter:})
	pool.ProductTemplate().AddMany2OneField("CurrencyId", models.ForeignKeyFieldParams{String: "Currency", RelationModel: pool.Currency()})
	//, Compute: "ComputeCurrencyId"})
	pool.ProductTemplate().AddFloatField("Price", models.FloatFieldParams{String: "Price"}) //,Compute:"ComputeTemplatePrice" , inverse:"SetTemplatePrice" ,
	// Digits:dp.getPrecision("ProductPrice")})
	pool.ProductTemplate().AddFloatField("ListPrice", models.FloatFieldParams{String: "SalePrice",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 1.0
	}, Help: "Base price to compute the customer price. Sometimes called the catalog price."}) // , Digits:dp.getPrecision("ProductPrice")})
	pool.ProductTemplate().AddFloatField("LstPrice", models.FloatFieldParams{String: "Public Price", Related: "ListPrice"})
	//,Digits:dp.getPrecision("ProductPrice")})
	pool.ProductTemplate().AddFloatField("StandardPrice", models.FloatFieldParams{String: "Cost", GroupOperator: "base.GroupUser",
		Help: "Cost of the product, in the default unit of measure of the product."})
	//,inverse:"setStandartPrice" , Digits:dp.getPrecision("ProductPrice")})
	pool.ProductTemplate().AddFloatField("Volume", models.FloatFieldParams{String: "Volume", Help: "The volume in m3."})
	// Stored: true , Compute:"ComputeVolume" , inverse:"setVolume"})
	pool.ProductTemplate().AddFloatField("Weight", models.FloatFieldParams{String: "Weight", Help: "The weight of the contents in Kg, " +
		"not including any packaging, etc.", Stored: true}) //inverse:"setWeight", Digits:dp.getPrecision("Stock Weight") })
	pool.ProductTemplate().AddFloatField("Warranty", models.FloatFieldParams{String: "Warranty"})
	pool.ProductTemplate().AddBooleanField("ScaleOk", models.SimpleFieldParams{String: "Can be Sold",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	}, Help: "Specify if the product can be selected in a sales order line."})
	pool.ProductTemplate().AddBooleanField("PurchaseOk", models.SimpleFieldParams{String: "Can be purchased",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	}})
	pool.ProductTemplate().AddMany2OneField("PriceListId", models.ForeignKeyFieldParams{String: "Price List", RelationModel: pool.ProductPriceList(),
		//, Stored: true,
		Help: "Technical field. Used for searching on pricelists, not stored in database."})
	pool.ProductTemplate().AddMany2OneField("UomId", models.ForeignKeyFieldParams{String: "Unit of Measure", RelationModel: pool.ProductUoM(), Required: true,
		Help: "Default Unit of Measure used for all stock operation."}) //, Default:"getDefaultUomId"})
	pool.ProductTemplate().AddMany2OneField("UomPoId", models.ForeignKeyFieldParams{String: "Purchase Unit Of Measure", //Default:"getDefaultUomId",
		RelationModel: pool.ProductUoM(),
		Help:          "Default Unit of Measure used for purchase orders. It must be in the same category than the default unit of measure."})
	pool.ProductTemplate().AddMany2OneField("CompanyId", models.ForeignKeyFieldParams{String: "CompanyId", RelationModel: pool.Company()})
	//Default:lambda self: self.env['res.company']._company_default_get('product.template'), index=1),

	pool.ProductTemplate().AddOne2ManyField("PackagingIds", models.ReverseFieldParams{String: "Logistical Units", RelationModel: pool.ProductPackaging(),
		ReverseFK: "ProductTmplId", Help: "Gives the different ways to package the same product. This has no impact on the picking order and is mainly used if " +
			"you use the EDI module."})
	//pool.ProductTemplate().AddOne2ManyField("SellerIds", models.ReverseFieldParams{String: "Vendors", RelationModel:pool.ProductSupplierInfo() ,
	// ReverseFK:"ProductTmplID" })
	pool.ProductTemplate().AddBooleanField("Active", models.SimpleFieldParams{String: "Active",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	}, Help: "If unchecked, it will allow you to hide the product without removing it."})
	pool.ProductTemplate().AddIntegerField("Color", models.SimpleFieldParams{String: "Color Index"})
	pool.ProductTemplate().AddOne2ManyField("AttributeLineIds", models.ReverseFieldParams{String: "Product Attributes",
		RelationModel: pool.ProductAttribute(), //.Line() ,
		ReverseFK: "ProductTmplId"})
	pool.ProductTemplate().AddOne2ManyField("ProductVariantIds", models.ReverseFieldParams{String: "Products", RelationModel: pool.ProductProduct(),
		ReverseFK: "ProductTmplId"})
	pool.ProductTemplate().AddMany2OneField("ProductVariantID", models.ForeignKeyFieldParams{String: "Product", RelationModel: pool.ProductProduct()})
	//, Compute:"ComputeProductVariantId"})
	pool.ProductTemplate().AddIntegerField("ProductVariantCount", models.SimpleFieldParams{String: "# Product Variants"})
	//, Compute:"ComputeProductVariantCount"})
	pool.ProductTemplate().AddCharField("Barcode", models.StringFieldParams{String: "Barcode"}) //, Related:"pool.ProductProduct().Barcode()"})
	pool.ProductTemplate().AddCharField("DefaultCode", models.StringFieldParams{String: "Internal Reference",  Stored:true })
	//, Compute:"ComputeDefaultCode" , inverse:"setDefaultCode"})
	pool.ProductTemplate().AddOne2ManyField("ItemIds", models.ReverseFieldParams{String: "Pricelist Items", RelationModel: pool.ProductPriceList(),
		ReverseFK: "ProductTmplId"}) //.Item()})
	pool.ProductTemplate().AddBinaryField("Image", models.SimpleFieldParams{String: "Image", Help: "This field holds the image used as image " +
		"for the product, limited to 1024x1024px."}) //, attachement:true})
	pool.ProductTemplate().AddBinaryField("ImageMedium", models.SimpleFieldParams{String: "Medium-sized Image",
		Help: "Medium-sized image of the product. It is automatically resized as a 128x128px image, with aspect ratio preserved, only when the image exceeds " +
			"one of those sizes. Use this field in form views or some kanban views."}) //, attachment:true})
	pool.ProductTemplate().AddBinaryField("ImageSmall", models.SimpleFieldParams{String: "Small-sized Image",
		Help: "Small-sized image of the product. It is automatically resized as a 64x64px image, with aspect ratio preserved. Use this field anywhere " +
			"a small image is required."})


}
