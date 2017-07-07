package product

import (
	"github.com/hexya-erp/hexya/pool"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
)

func init(){

	pool.BaseConfigSetting().DeclareModel()
	pool.BaseConfigSetting().AddBooleanField("CompanyShareProduct" , models.SimpleFieldParams{String:"Share product to all companies" ,
		Help:"Share your product to all companies defined in your instance.\n"+
		" * Checked : Product are visible for every company, even if a company is defined on the partner.\n"+
		" * Unchecked : Each company can see only its product (product where company is defined). Product not related to a company are visible for all companies."})
	pool.BaseConfigSetting().AddSelectionField("GroupProductVariant", models.SelectionFieldParams{String: "Product Variants", Selection: types.Selection{
		"0" : "No variants on products",
		"1" : "Products can have several attributes, defining variants (Example: size, color,...)",
	}, Help:"Work with product variant allows you to define some variant of the same products, an ease the product management in the ecommerce for example" , })
	//ImpliedGroup:product.group_product_variant})
}