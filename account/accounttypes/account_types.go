// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package accounttypes

import "github.com/hexya-erp/hexya/hexya/models/types/dates"

// A PaymentDueDates gives the amount due of an invoice at the given date
type PaymentDueDates struct {
	Date   dates.Date
	Amount float64
}

// A TaxGroup holds an amount for a given group name
type TaxGroup struct {
	GroupName string
	TaxAmount float64
}

// A DataForReconciliationWidget holds data for the reconciliation widget
type DataForReconciliationWidget struct {
	Customers []map[string]interface{} `json:"customers"`
	Suppliers []map[string]interface{} `json:"suppliers"`
	Accounts  []map[string]interface{} `json:"accounts"`
}
