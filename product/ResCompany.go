package product

import "github.com/hexya-erp/hexya/pool"

func init() {

	pool.ResCompany().DeclareModel()
	pool.ResCompany().InheritModel(pool.Company())
}