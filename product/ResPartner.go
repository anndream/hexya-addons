package product

import (
	"github.com/hexya-erp/hexya/pool"
	"github.com/hexya-erp/hexya/hexya/models"
)

func init() {

	pool.Partner().AddMany2OneField("PropertyProductPricelist" , models.ForeignKeyFieldParams{String:"PropertyProductPricelist", RelationModel:pool.ProductPriceList(),
		// Compute:"ComputeProductPricelist"})
		Help:"This pricelist will be used, instead of the default one, for sales to the current partner" })//, inverse:"InverseProductPricelist"})

}
