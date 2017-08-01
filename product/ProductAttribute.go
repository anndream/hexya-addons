package product

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool"
)

func init() {

	//class ProductAttribute

	pool.ProductAttribute().DeclareModel()
	pool.ProductAttribute().AddCharField("Name", models.StringFieldParams{String: "Name", Required: true, Translate: true})
	pool.ProductAttribute().AddOne2ManyField("ValueIds", models.ReverseFieldParams{String: "Values", RelationModel: pool.ProductAttributeValue() ,
	ReverseFK: "AttributeId", NoCopy: false})
	pool.ProductAttribute().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence", Help: "Determine the display order"})
	pool.ProductAttribute().AddOne2ManyField("AttributeLineIds", models.ReverseFieldParams{String: "Lines", RelationModel: pool.ProductAttributeLine() ,
	ReverseFK: "AttributeId"})
	pool.ProductAttribute().AddBooleanField("CreateVariant", models.SimpleFieldParams{String: "CreateVariant",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return true
	}, Help: "Check this if you want to create multiple variants for this attribute."})

	//class ProductAttributeValue

	pool.ProductAttributeValue().DeclareModel()
	pool.ProductAttributeValue().AddCharField("Name", models.StringFieldParams{String: "Value", Required: true, Translate: true})
	pool.ProductAttributeValue().AddIntegerField("Sequence", models.SimpleFieldParams{String: "Sequence", Help: "Determine the display order"})
	pool.ProductAttributeValue().AddMany2OneField("AttributeId", models.ForeignKeyFieldParams{String: "Attribute", RelationModel: pool.ProductAttribute(),
		Required: true, OnDelete: models.Cascade})
	pool.ProductAttributeValue().AddMany2ManyField("ProductsIds", models.Many2ManyFieldParams{String: "Variant", RelationModel: pool.ProductProduct()})
	//ProductIds ReadOnly
	pool.ProductAttributeValue().Fields().ProductsIds().RevokeAccess(security.GroupEveryone, security.Write)
	pool.ProductAttributeValue().AddFloatField("PriceExtra", models.FloatFieldParams{String: "Attribute Price Extra",
		Default: func(models.Environment, models.FieldMap) interface{} {
		return 0.0
	}, Help: "Price Extra: Extra price for the variant with this attribute value on sale price. eg. 200 price extra, 1000 + 200 = 1200."})
	//Compute:"ComputePriceExtra" , inverse:"setPriceExtra" , Digits:dp.get_precision("Product Price")})
	pool.ProductAttributeValue().AddSQLConstraint("ValueCompanyUniq", "unique(Name,attributeId)", "This attribute value already exists !")


	//class ProductAttributePrice

	pool.ProductAttributePrice().DeclareModel()
	pool.ProductAttributePrice().AddMany2OneField("ProductTmplId" , models.ForeignKeyFieldParams{String:"Product Template" ,
		RelationModel:pool.ProductTemplate() , Required:true ,
		OnDelete:models.Cascade})
	pool.ProductAttributePrice().AddMany2OneField("ValueId" , models.ForeignKeyFieldParams{String:"Product Attribute Value" ,
		RelationModel:pool.ProductAttributeValue() , Required:true ,
		OnDelete:models.Cascade})
	pool.ProductAttributePrice().AddFloatField("PriceExtra" , models.FloatFieldParams{String:"Price Extra" })//, Digits:dp.get_precision("Product Price")})

	//class ProductAttributeLine

	pool.ProductAttributeLine().DeclareModel()
	pool.ProductAttributeLine().AddMany2OneField("ProductTmplId" , models.ForeignKeyFieldParams{String:"Product Template" ,
		RelationModel:pool.ProductTemplate() , Required:true ,
	OnDelete:models.Cascade})
	pool.ProductAttributeLine().AddMany2OneField("AttributeId" , models.ForeignKeyFieldParams{String:"Attribute" ,
		RelationModel:pool.ProductAttribute() , Required:true ,
	 OnDelete:models.Restrict})
	pool.ProductAttributeLine().AddMany2ManyField("ValueIds" , models.Many2ManyFieldParams{String:"Attribute Value" ,
		RelationModel:pool.ProductAttributeValue() , Required:true })
	//OnDelete:models.Cascade})
}
