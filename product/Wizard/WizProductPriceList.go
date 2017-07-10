package product

import (
	"github.com/hexya-erp/hexya/pool"
)

func init() {

	pool.WizProductPriceList().DeclareTransientModel()
}
