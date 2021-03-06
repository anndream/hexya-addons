// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package account

import (
	"github.com/hexya-erp/hexya-addons/account/accounttypes"
	"github.com/hexya-erp/hexya-addons/decimalPrecision"
	"github.com/hexya-erp/hexya-base/web/webdata"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	"github.com/hexya-erp/hexya/pool"
)

var ReferenceType = types.Selection{
	"none": "Free Reference",
}

func init() {

	pool.AccountInvoice().DeclareModel()
	pool.AccountInvoice().SetDefaultOrder("DateInvoice DESC", "Number DESC", "ID DESC")

	pool.AccountInvoice().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Reference/Description", Index: true, /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			NoCopy: true, Help: "The name that will be used on account move lines"},
		"Origin": models.CharField{String: "Source Document",
			Help: "Reference of the document that produced this invoice." /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/},
		"Type": models.SelectionField{Selection: types.Selection{
			"out_invoice": "Customer Invoice",
			"in_invoice":  "Vendor Bill",
			"out_refund":  "Customer Refund",
			"in_refund":   "Vendor Refund",
		}, /*[readonly True]*/ Index: true,
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				if env.Context().HasKey("type") {
					return env.Context().GetString("type")
				}
				return "out_invoice"
			} /*[ track_visibility 'always']*/},
		"RefundInvoice": models.Many2OneField{String: "Invoice for which this invoice is the refund",
			RelationModel: pool.AccountInvoice()},
		"Number": models.CharField{Related: "Move.Name" /*[ readonly True]*/, NoCopy: true},
		"MoveName": models.CharField{String: "Journal Entry Name" /*[ readonly False]*/, NoCopy: true,
			Default: models.DefaultValue(false),
			Help: `Technical field holding the number given to the invoice, automatically set when the invoice is
validated then stored to set the same number again if the invoice is cancelled,
set to draft and re-validated.`},
		"Reference": models.CharField{String: "Vendor Reference",
			Help: "The partner reference of this invoice." /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/},
		"ReferenceType": models.SelectionField{Selection: ReferenceType,
			String:  "Payment Reference", /* states draft:"readonly": "False" */
			Default: models.DefaultValue("none"), Required: true},
		"Comment": models.TextField{String: "Additional Information" /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/},
		"State": models.SelectionField{String: "Status", Selection: types.Selection{
			"draft":     "Draft",
			"proforma":  "Pro-forma",
			"proforma2": "Pro-forma",
			"open":      "Open",
			"paid":      "Paid",
			"cancel":    "Cancelled",
		}, Index: true /*[ readonly True]*/, Default: models.DefaultValue("draft"), /*[ track_visibility 'onchange']*/
			NoCopy: true,
			Help: ` * The 'Draft' status is used when a user is encoding a new and unconfirmed Invoice.
 * The 'Pro-forma' status is used when the invoice does not have an invoice number.
 * The 'Open' status is used when user creates invoice an invoice number is generated.
   It stays in the open status till the user pays the invoice.
 * The 'Paid' status is set automatically when the invoice is paid. Its related journal
   entries may or may not be reconciled.
 * The 'Cancelled' status is used when user cancel invoice.`},
		"Sent": models.BooleanField{ /*[readonly True]*/ Default: models.DefaultValue(false),
			NoCopy: true, Help: "It indicates that the invoice has been sent."},
		"DateInvoice": models.DateField{String: "Invoice Date", /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Index: true, Help: "Keep empty to use the current date", NoCopy: true,
			Constraint: pool.AccountInvoice().Methods().OnchangePaymentTermDateInvoice()},
		"DateDue": models.DateField{String: "Due Date", /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Index: true, NoCopy: true,
			Help: `If you use payment terms, the due date will be computed automatically at the generation
of accounting entries. The payment term may compute several due dates, for example 50%
now and 50% in one month, but if you want to force a due date, make sure that the payment
term is not set on the invoice. If you keep the payment term and the due date empty, it
means direct payment.`},
		"Partner": models.Many2OneField{RelationModel: pool.Partner(), Required: true, /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/ /*[ track_visibility 'always']*/
			OnChange: pool.AccountInvoice().Methods().OnchangePartner()},
		"PaymentTerm": models.Many2OneField{String: "Payment Terms",
			RelationModel: pool.AccountPaymentTerm(), /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Constraint:    pool.AccountInvoice().Methods().OnchangePaymentTermDateInvoice(),
			Help: `If you use payment terms, the due date will be computed automatically at the generation
of accounting entries. If you keep the payment term and the due date empty, it means direct payment.
The payment term may compute several due dates, for example 50% now, 50% in one month.`},
		"Date": models.DateField{String: "Accounting Date", NoCopy: true,
			Help: "Keep empty to use the invoice date." /*[ readonly True]*/ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/},
		"Account": models.Many2OneField{String: "Account", RelationModel: pool.AccountAccount(),
			Required: true, /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Filter:   pool.AccountAccount().Deprecated().Equals(false),
			Help:     "The partner account used for this invoice."},
		"InvoiceLines": models.One2ManyField{String: "Invoice Lines", RelationModel: pool.AccountInvoiceLine(),
			ReverseFK: "Invoice", JSON: "invoice_line_ids", /* readonly */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			OnChange: pool.AccountInvoice().Methods().OnchangeInvoiceLines(),
			NoCopy:   false},
		"TaxLines": models.One2ManyField{RelationModel: pool.AccountInvoiceTax(), ReverseFK: "Invoice",
			JSON: "tax_line_ids" /* readonly */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/, NoCopy: false},
		"Move": models.Many2OneField{String: "Journal Entry", RelationModel: pool.AccountMove(), /* readonly=true */
			Index: true, OnDelete: models.Restrict, NoCopy: true,
			Help: "Link to the automatically generated Journal Items."},
		"AmountUntaxed": models.FloatField{String: "Untaxed Amount", Stored: true,
			Compute: pool.AccountInvoice().Methods().ComputeAmount(), /*[ track_visibility 'always']*/
			Depends: []string{"InvoiceLines.PriceSubtotal", "TaxLines.Amount", "Currency", "Company", "DateInvoice", "Type"}},
		"AmountUntaxedSigned": models.FloatField{String: "Untaxed Amount in Company Currency", Stored: true,
			Compute: pool.AccountInvoice().Methods().ComputeAmount(),
			Depends: []string{"InvoiceLines.PriceSubtotal", "TaxLines.Amount", "Currency", "Company", "DateInvoice", "Type"}},
		"AmountTax": models.FloatField{String: "Tax", Stored: true,
			Compute: pool.AccountInvoice().Methods().ComputeAmount(),
			Depends: []string{"InvoiceLines.PriceSubtotal", "TaxLines.Amount", "Currency", "Company", "DateInvoice", "Type"}},
		"AmountTotal": models.FloatField{String: "Total", Stored: true,
			Compute: pool.AccountInvoice().Methods().ComputeAmount(),
			Depends: []string{"InvoiceLines.PriceSubtotal", "TaxLines.Amount", "Currency", "Company", "DateInvoice", "Type"}},
		"AmountTotalSigned": models.FloatField{String: "Total in Invoice Currency", Stored: true,
			Compute: pool.AccountInvoice().Methods().ComputeAmount(),
			Depends: []string{"InvoiceLines.PriceSubtotal", "TaxLines.Amount", "Currency", "Company", "DateInvoice", "Type"},
			Help:    "Total amount in the currency of the invoice, negative for credit notes."},
		"AmountTotalCompanySigned": models.FloatField{String: "Total in Company Currency", Stored: true,
			Compute: pool.AccountInvoice().Methods().ComputeAmount(),
			Depends: []string{"InvoiceLines.PriceSubtotal", "TaxLines.Amount", "Currency", "Company", "DateInvoice", "Type"},
			Help:    "Total amount in the currency of the company, negative for credit notes."},
		"Currency": models.Many2OneField{RelationModel: pool.Currency(),
			Required: true, /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				journal := pool.AccountInvoice().NewSet(env).DefaultJournal()
				if !journal.Currency().IsEmpty() {
					return journal.Currency()
				}
				if !journal.Company().Currency().IsEmpty() {
					return journal.Company().Currency()
				}
				return pool.User().NewSet(env).CurrentUser().Company().Currency()
			} /*[ track_visibility 'always']*/},
		"CompanyCurrency": models.Many2OneField{RelationModel: pool.Currency(),
			Related: "Company.Currency" /* readonly=true */},
		"Journal": models.Many2OneField{RelationModel: pool.AccountJournal(), Required: true, /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			OnChange: pool.AccountInvoice().Methods().OnchangeJournal(),
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.AccountInvoice().NewSet(env).DefaultJournal()
			} /*Filter: "[('type'*/ /*[ 'in']*/ /*[ {'out_invoice': ['sale']]*/ /*[ 'out_refund': ['sale']]*/ /*[ 'in_refund': ['purchase']] [ 'in_invoice': ['purchase']}.get(type,  ('company_id']*/ /*[ ' ']*/ /*[ company_id)]"]*/},
		"Company": models.Many2OneField{RelationModel: pool.Company(), Required: true, /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.Company().NewSet(env).CompanyDefaultGet()
			}, OnChange: pool.AccountInvoice().Methods().OnchangePartner()},
		"Reconciled": models.BooleanField{String: "Paid/Reconciled", Stored: true, /*[ readonly True]*/
			Compute: pool.AccountInvoice().Methods().ComputeResidual(),
			Depends: []string{"State", "Currency", "InvoiceLines.PriceSubtotal", "Move.Lines.AmountResidual", "Move.Lines.Currency"},
			Help: `It indicates that the invoice has been paid and the journal entry of the invoice
has been reconciled with one or several journal entries of payment.`},
		"PartnerBank": models.Many2OneField{String: "Bank Account", RelationModel: pool.BankAccount(),
			Help: `Bank Account Number to which the invoice will be paid.
A Company bank account if this is a Customer Invoice or Vendor Refund, otherwise a Partner bank account number.`, /* readonly=True */ /* states={'draft': [('readonly', False)]}" */
		},
		"Residual": models.FloatField{String: "Amount Due",
			Compute: pool.AccountInvoice().Methods().ComputeResidual(), Stored: true, Help: "Remaining amount due.",
			Depends: []string{"State", "Currency", "InvoiceLines.PriceSubtotal", "Move.Lines.AmountResidual", "Move.Lines.Currency"}},
		"ResidualSigned": models.FloatField{String: "Amount Due in Invoice Currency",
			Compute: pool.AccountInvoice().Methods().ComputeResidual(), Stored: true,
			Depends: []string{"State", "Currency", "InvoiceLines.PriceSubtotal", "Move.Lines.AmountResidual", "Move.Lines.Currency"},
			Help:    "Remaining amount due in the currency of the invoice."},
		"ResidualCompanySigned": models.FloatField{String: "Amount Due in Company Currency",
			Compute: pool.AccountInvoice().Methods().ComputeResidual(), Stored: true,
			Depends: []string{"State", "Currency", "InvoiceLines.PriceSubtotal", "Move.Lines.AmountResidual", "Move.Lines.Currency"},
			Help:    "Remaining amount due in the currency of the company."},
		"Payments": models.Many2ManyField{RelationModel: pool.AccountPayment(), JSON: "payment_ids",
			NoCopy: true /*[ readonly True]*/},
		"PaymentMoveLines": models.Many2ManyField{String: "Payment Move Lines", RelationModel: pool.AccountMoveLine(),
			JSON: "payment_move_line_ids", Compute: pool.AccountInvoice().Methods().ComputePayments(), Stored: true,
			Depends: []string{"Move.Lines.AmountResidual"}},
		"User": models.Many2OneField{String: "Salesperson", RelationModel: pool.User(), /*[ track_visibility 'onchange']*/
			/* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.User().NewSet(env).CurrentUser()
			}},
		"FiscalPosition": models.Many2OneField{RelationModel: pool.AccountFiscalPosition() /* readonly=true */ /*[ states {'draft': [('readonly']*/ /*[ False)]}]*/},
		"CommercialPartner": models.Many2OneField{String: "Commercial Entity",
			RelationModel: pool.Partner(), Related: "Partner.CommercialPartner", /* readonly=true */
			Help: "The commercial entity that will be used on Journal Entries for this invoice"},
		"OutstandingCreditsDebitsWidget": models.TextField{Compute: pool.AccountInvoice().Methods().GetOutstandingInfoJSON()},
		"PaymentsWidget": models.TextField{Compute: pool.AccountInvoice().Methods().GetPaymentInfoJSON(),
			Depends: []string{"PaymentMoveLines.AmountResidual"}},
		"HasOutstanding": models.BooleanField{Compute: pool.AccountInvoice().Methods().GetOutstandingInfoJSON()},
	})

	// TODO implement as constraint
	//pool.AccountInvoice().AddSQLConstraint("number_uniq", "unique(number, company_id, journal_id, type)",
	//	"Invoice Number must be unique per Company!")

	pool.AccountInvoice().Methods().ComputeAmount().DeclareMethod(
		`ComputeAmount`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.depends('invoice_line_ids.price_subtotal','tax_line_ids.amount','currency_id','company_id','date_invoice','type')
			/*def _compute_amount(self):
			  self.amount_untaxed = sum(line.price_subtotal for line in self.invoice_line_ids)
			  self.amount_tax = sum(line.amount for line in self.tax_line_ids)
			  self.amount_total = self.amount_untaxed + self.amount_tax
			  amount_total_company_signed = self.amount_total
			  amount_untaxed_signed = self.amount_untaxed
			  if self.currency_id and self.company_id and self.currency_id != self.company_id.currency_id:
			      currency_id = self.currency_id.with_context(date=self.date_invoice)
			      amount_total_company_signed = currency_id.compute(self.amount_total, self.company_id.currency_id)
			      amount_untaxed_signed = currency_id.compute(self.amount_untaxed, self.company_id.currency_id)
			  sign = self.type in ['in_refund', 'out_refund'] and -1 or 1
			  self.amount_total_company_signed = amount_total_company_signed * sign
			  self.amount_total_signed = self.amount_total * sign
			  self.amount_untaxed_signed = amount_untaxed_signed * sign

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().DefaultJournal().DeclareMethod(
		`DefaultJournal`,
		func(rs pool.AccountInvoiceSet) pool.AccountJournalSet {
			//@api.model
			/*def _default_journal(self):
			  if self._context.get('default_journal_id', False):
			      return self.env['account.journal'].browse(self._context.get('default_journal_id'))
			  inv_type = self._context.get('type', 'out_invoice')
			  inv_types = inv_type if isinstance(inv_type, list) else [inv_type]
			  company_id = self._context.get('company_id', self.env.user.company_id.id)
			  domain = [
			      ('type', 'in', filter(None, map(TYPE2JOURNAL.get, inv_types))),
			      ('company_id', '=', company_id),
			  ]
			  return self.env['account.journal'].search(domain, limit=1)

			*/
			return pool.AccountJournal().NewSet(rs.Env())
		})

	pool.AccountInvoice().Methods().ComputeResidual().DeclareMethod(
		`ComputeResidual`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			/*def _compute_residual(self):
			  residual = 0.0
			  residual_company_signed = 0.0
			  sign = self.type in ['in_refund', 'out_refund'] and -1 or 1
			  for line in self.sudo().move_id.line_ids:
			      if line.account_id.internal_type in ('receivable', 'payable'):
			          residual_company_signed += line.amount_residual
			          if line.currency_id == self.currency_id:
			              residual += line.amount_residual_currency if line.currency_id else line.amount_residual
			          else:
			              from_currency = (line.currency_id and line.currency_id.with_context(date=line.date)) or line.company_id.currency_id.with_context(date=line.date)
			              residual += from_currency.compute(line.amount_residual, self.currency_id)
			  self.residual_company_signed = abs(residual_company_signed) * sign
			  self.residual_signed = abs(residual) * sign
			  self.residual = abs(residual)
			  digits_rounding_precision = self.currency_id.rounding
			  if float_is_zero(self.residual, precision_rounding=digits_rounding_precision):
			      self.reconciled = True
			  else:
			      self.reconciled = False

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().GetOutstandingInfoJSON().DeclareMethod(
		`GetOutstandingInfoJSON`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.one
			/*def _get_outstanding_info_JSON(self):
			  self.outstanding_credits_debits_widget = json.dumps(False)
			  if self.state == 'open':
			      domain = [('account_id', '=', self.account_id.id), ('partner_id', '=', self.env['res.partner']._find_accounting_partner(self.partner_id).id), ('reconciled', '=', False), ('amount_residual', '!=', 0.0)]
			      if self.type in ('out_invoice', 'in_refund'):
			          domain.extend([('credit', '>', 0), ('debit', '=', 0)])
			          type_payment = _('Outstanding credits')
			      else:
			          domain.extend([('credit', '=', 0), ('debit', '>', 0)])
			          type_payment = _('Outstanding debits')
			      info = {'title': '', 'outstanding': True, 'content': [], 'invoice_id': self.id}
			      lines = self.env['account.move.line'].search(domain)
			      currency_id = self.currency_id
			      if len(lines) != 0:
			          for line in lines:
			              # get the outstanding residual value in invoice currency
			              if line.currency_id and line.currency_id == self.currency_id:
			                  amount_to_show = abs(line.amount_residual_currency)
			              else:
			                  amount_to_show = line.company_id.currency_id.with_context(date=line.date).compute(abs(line.amount_residual), self.currency_id)
			              if float_is_zero(amount_to_show, precision_rounding=self.currency_id.rounding):
			                  continue
			              info['content'].append({
			                  'journal_name': line.ref or line.move_id.name,
			                  'amount': amount_to_show,
			                  'currency': currency_id.symbol,
			                  'id': line.id,
			                  'position': currency_id.position,
			                  'digits': [69, self.currency_id.decimal_places],
			              })
			          info['title'] = type_payment
			          self.outstanding_credits_debits_widget = json.dumps(info)
			          self.has_outstanding = True

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().GetPaymentInfoJSON().DeclareMethod(
		`GetPaymentInfoJSON`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.depends('payment_move_line_ids.amount_residual')
			/*def _get_payment_info_JSON(self):
			  self.payments_widget = json.dumps(False)
			  if self.payment_move_line_ids:
			      info = {'title': _('Less Payment'), 'outstanding': False, 'content': []}
			      currency_id = self.currency_id
			      for payment in self.payment_move_line_ids:
			          payment_currency_id = False
			          if self.type in ('out_invoice', 'in_refund'):
			              amount = sum([p.amount for p in payment.matched_debit_ids if p.debit_move_id in self.move_id.line_ids])
			              amount_currency = sum([p.amount_currency for p in payment.matched_debit_ids if p.debit_move_id in self.move_id.line_ids])
			              if payment.matched_debit_ids:
			                  payment_currency_id = all([p.currency_id == payment.matched_debit_ids[0].currency_id for p in payment.matched_debit_ids]) and payment.matched_debit_ids[0].currency_id or False
			          elif self.type in ('in_invoice', 'out_refund'):
			              amount = sum([p.amount for p in payment.matched_credit_ids if p.credit_move_id in self.move_id.line_ids])
			              amount_currency = sum([p.amount_currency for p in payment.matched_credit_ids if p.credit_move_id in self.move_id.line_ids])
			              if payment.matched_credit_ids:
			                  payment_currency_id = all([p.currency_id == payment.matched_credit_ids[0].currency_id for p in payment.matched_credit_ids]) and payment.matched_credit_ids[0].currency_id or False
			          # get the payment value in invoice currency
			          if payment_currency_id and payment_currency_id == self.currency_id:
			              amount_to_show = amount_currency
			          else:
			              amount_to_show = payment.company_id.currency_id.with_context(date=payment.date).compute(amount, self.currency_id)
			          if float_is_zero(amount_to_show, precision_rounding=self.currency_id.rounding):
			              continue
			          payment_ref = payment.move_id.name
			          if payment.move_id.ref:
			              payment_ref += ' (' + payment.move_id.ref + ')'
			          info['content'].append({
			              'name': payment.name,
			              'journal_name': payment.journal_id.name,
			              'amount': amount_to_show,
			              'currency': currency_id.symbol,
			              'digits': [69, currency_id.decimal_places],
			              'position': currency_id.position,
			              'date': payment.date,
			              'payment_id': payment.id,
			              'move_id': payment.move_id.id,
			              'ref': payment_ref,
			          })
			      self.payments_widget = json.dumps(info)

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().ComputePayments().DeclareMethod(
		`ComputePayments`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.depends('move_id.line_ids.amount_residual')
			/*def _compute_payments(self):
			  payment_lines = []
			  for line in self.move_id.line_ids:
			      payment_lines.extend(filter(None, [rp.credit_move_id.id for rp in line.matched_credit_ids]))
			      payment_lines.extend(filter(None, [rp.debit_move_id.id for rp in line.matched_debit_ids]))
			  self.payment_move_line_ids = self.env['account.move.line'].browse(list(set(payment_lines)))

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().Create().Extend("",
		func(rs pool.AccountInvoiceSet, data *pool.AccountInvoiceData) pool.AccountInvoiceSet {
			//@api.model
			/*def create(self, vals):
			  onchanges = {
			      '_onchange_partner_id': ['account_id', 'payment_term_id', 'fiscal_position_id', 'partner_bank_id'],
			      '_onchange_journal_id': ['currency_id'],
			  }
			  for onchange_method, changed_fields in onchanges.items():
			      if any(f not in vals for f in changed_fields):
			          invoice = self.new(vals)
			          getattr(invoice, onchange_method)()
			          for field in changed_fields:
			              if field not in vals and invoice[field]:
			                  vals[field] = invoice._fields[field].convert_to_write(invoice[field], invoice)
			  if not vals.get('account_id',False):
			      raise UserError(_('Configuration error!\nCould not find any account to create the invoice, are you sure you have a chart of account installed?'))

			  invoice = super(AccountInvoice, self.with_context(mail_create_nolog=True)).create(vals)

			  if any(line.invoice_line_tax_ids for line in invoice.invoice_line_ids) and not invoice.tax_line_ids:
			      invoice.compute_taxes()

			  return invoice

			*/
			return rs.Super().Create(data)
		})

	pool.AccountInvoice().Methods().Write().Extend("",
		func(rs pool.AccountInvoiceSet, vals *pool.AccountInvoiceData, fieldsToReset ...models.FieldNamer) bool {
			//@api.multi
			/*def _write(self, vals):
			  pre_not_reconciled = self.filtered(lambda invoice: not invoice.reconciled)
			  pre_reconciled = self - pre_not_reconciled
			  res = super(AccountInvoice, self)._write(vals)
			  reconciled = self.filtered(lambda invoice: invoice.reconciled)
			  not_reconciled = self - reconciled
			  (reconciled & pre_reconciled).filtered(lambda invoice: invoice.state == 'open').action_invoice_paid()
			  (not_reconciled & pre_not_reconciled).filtered(lambda invoice: invoice.state == 'paid').action_invoice_re_open()
			  return res

			*/
			return rs.Super().Write(vals, fieldsToReset...)
		})

	pool.AccountInvoice().Methods().FieldsViewGet().Extend("",
		func(rs pool.AccountInvoiceSet, params webdata.FieldsViewGetParams) *webdata.FieldsViewData {
			//@api.model
			/*
				  def fields_view_get(self, view_id=None, view_type='form', toolbar=False, submenu=False):
					  def get_view_id(xid, name):
						  try:
							  return self.env.ref('account.' + xid)
						  except ValueError:
							  view = self.env['ir.ui.view'].search([('name', '=', name)], limit=1)
							  if not view:
								  return False
							  return view.id

					  context = self._context
					  if context.get('active_model') == 'res.partner' and context.get('active_ids'):
						  partner = self.env['res.partner'].browse(context['active_ids'])[0]
						  if not view_type:
							  view_id = get_view_id('invoice_tree', 'account.invoice.tree')
							  view_type = 'tree'
						  elif view_type == 'form':
							  if partner.supplier and not partner.customer:
								  view_id = get_view_id('invoice_supplier_form', 'account.invoice.supplier.form').id
							  elif partner.customer and not partner.supplier:
								  view_id = get_view_id('invoice_form', 'account.invoice.form').id
					  return super(AccountInvoice, self).fields_view_get(view_id=view_id, view_type=view_type, toolbar=toolbar, submenu=submenu)
			*/
			return rs.Super().FieldsViewGet(params)
		})

	pool.AccountInvoice().Methods().InvoicePrint().DeclareMethod(
		`InvoicePrint`,
		func(rs pool.AccountInvoiceSet) *actions.Action {
			//@api.multi
			/*def invoice_print(self):
			  """ Print the invoice and mark it as sent, so that we can see more
			      easily the next step of the workflow
			  """
			  self.ensure_one()
			  self.sent = True
			  return self.env['report'].get_action(self, 'account.report_invoice')

			*/
			return &actions.Action{
				Type: actions.ActionCloseWindow,
			}
		})

	pool.AccountInvoice().Methods().ActionInvoiceSent().DeclareMethod(
		`ActionInvoiceSent`,
		func(rs pool.AccountInvoiceSet) *actions.Action {
			//@api.multi
			/*def action_invoice_sent(self):
			  """ Open a window to compose an email, with the edi invoice template
			      message loaded by default
			  """
			  self.ensure_one()
			  template = self.env.ref('account.email_template_edi_invoice', False)
			  compose_form = self.env.ref('mail.email_compose_message_wizard_form', False)
			  ctx = dict(
			      default_model='account.invoice',
			      default_res_id=self.id,
			      default_use_template=bool(template),
			      default_template_id=template and template.id or False,
			      default_composition_mode='comment',
			      mark_invoice_as_sent=True,
			      custom_layout="account.mail_template_data_notification_email_account_invoice"
			  )
			  return {
			      'name': _('Compose Email'),
			      'type': 'ir.actions.act_window',
			      'view_type': 'form',
			      'view_mode': 'form',
			      'res_model': 'mail.compose.message',
			      'views': [(compose_form.id, 'form')],
			      'view_id': compose_form.id,
			      'target': 'new',
			      'context': ctx,
			  }

			*/
			return &actions.Action{
				Type: actions.ActionCloseWindow,
			}
		})

	pool.AccountInvoice().Methods().ComputeTaxes().DeclareMethod(
		`ComputeTaxes`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def compute_taxes(self):
			  """Function used in other module to compute the taxes on a fresh invoice created (onchanges did not applied)"""
			  account_invoice_tax = self.env['account.invoice.tax']
			  ctx = dict(self._context)
			  for invoice in self:
			      # Delete non-manual tax lines
			      self._cr.execute("DELETE FROM account_invoice_tax WHERE invoice_id=%s AND manual is False", (invoice.id,))
			      self.invalidate_cache()

			      # Generate one tax line per tax, however many invoice lines it's applied to
			      tax_grouped = invoice.get_taxes_values()

			      # Create new tax lines
			      for tax in tax_grouped.values():
			          account_invoice_tax.create(tax)

			  # dummy write on self to trigger recomputations
			  return self.with_context(ctx).write({'invoice_line_ids': []})

			*/
			return true
		})

	pool.AccountInvoice().Methods().Unlink().Extend("",
		func(rs pool.AccountInvoiceSet) int64 {
			//@api.multi
			/*def unlink(self):
			  for invoice in self:
			      if invoice.state not in ('draft', 'cancel'):
			          raise UserError(_('You cannot delete an invoice which is not draft or cancelled. You should refund it instead.'))
			      elif invoice.move_name:
			          raise UserError(_('You cannot delete an invoice after it has been validated (and received a number). You can set it back to "Draft" state and modify its content, then re-confirm it.'))
			  return super(AccountInvoice, self).unlink()

			*/
			return rs.Super().Unlink()
		})

	pool.AccountInvoice().Methods().OnchangeInvoiceLines().DeclareMethod(
		`OnchangeInvoiceLines`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.onchange('invoice_line_ids')
			/*def _onchange_invoice_line_ids(self):
			  taxes_grouped = self.get_taxes_values()
			  tax_lines = self.tax_line_ids.filtered('manual')
			  for tax in taxes_grouped.values():
			      tax_lines += tax_lines.new(tax)
			  self.tax_line_ids = tax_lines
			  return

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().OnchangePartner().DeclareMethod(
		`OnchangePartner`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.onchange('partner_id','company_id')
			/*def _onchange_partner_id(self):
			  account_id = False
			  payment_term_id = False
			  fiscal_position = False
			  bank_id = False
			  warning = {}
			  domain = {}
			  company_id = self.company_id.id
			  p = self.partner_id if not company_id else self.partner_id.with_context(force_company=company_id)
			  type = self.type
			  if p:
			      rec_account = p.property_account_receivable_id
			      pay_account = p.property_account_payable_id
			      if not rec_account and not pay_account:
			          action = self.env.ref('account.action_account_config')
			          msg = _('Cannot find a chart of accounts for this company, You should configure it. \nPlease go to Account Configuration.')
			          raise RedirectWarning(msg, action.id, _('Go to the configuration panel'))

			      if type in ('out_invoice', 'out_refund'):
			          account_id = rec_account.id
			          payment_term_id = p.property_payment_term_id.id
			      else:
			          account_id = pay_account.id
			          payment_term_id = p.property_supplier_payment_term_id.id

			      delivery_partner_id = self.get_delivery_partner_id()
			      fiscal_position = self.env['account.fiscal.position'].get_fiscal_position(self.partner_id.id, delivery_id=delivery_partner_id)

			      # If partner has no warning, check its company
			      if p.invoice_warn == 'no-message' and p.parent_id:
			          p = p.parent_id
			      if p.invoice_warn != 'no-message':
			          # Block if partner only has warning but parent company is blocked
			          if p.invoice_warn != 'block' and p.parent_id and p.parent_id.invoice_warn == 'block':
			              p = p.parent_id
			          warning = {
			              'title': _("Warning for %s") % p.name,
			              'message': p.invoice_warn_msg
			              }
			          if p.invoice_warn == 'block':
			              self.partner_id = False

			  self.account_id = account_id
			  self.payment_term_id = payment_term_id
			  self.fiscal_position_id = fiscal_position

			  if type in ('in_invoice', 'out_refund'):
			      bank_ids = p.commercial_partner_id.bank_ids
			      bank_id = bank_ids[0].id if bank_ids else False
			      self.partner_bank_id = bank_id
			      domain = {'partner_bank_id': [('id', 'in', bank_ids.ids)]}

			  res = {}
			  if warning:
			      res['warning'] = warning
			  if domain:
			      res['domain'] = domain
			  return res

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().GetDeliveryPartner().DeclareMethod(
		`GetDeliveryPartner`,
		func(rs pool.AccountInvoiceSet) pool.PartnerSet {
			//@api.multi
			/*def get_delivery_partner_id(self):
			  self.ensure_one()
			  return self.partner_id.address_get(['delivery'])['delivery']

			*/
			return pool.Partner().NewSet(rs.Env())
		})

	pool.AccountInvoice().Methods().OnchangeJournal().DeclareMethod(
		`OnchangeJournal`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.onchange('journal_id')
			/*def _onchange_journal_id(self):
			  if self.journal_id:
			      self.currency_id = self.journal_id.currency_id.id or self.journal_id.company_id.currency_id.id

			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().OnchangePaymentTermDateInvoice().DeclareMethod(
		`OnchangePaymentTermDateInvoice`,
		func(rs pool.AccountInvoiceSet) (*pool.AccountInvoiceData, []models.FieldNamer) {
			//@api.onchange('payment_term_id','date_invoice')
			/*
				def _onchange_payment_term_date_invoice(self):
					date_invoice = self.date_invoice
					if not date_invoice:
						date_invoice = fields.Date.context_today(self)
					if not self.payment_term_id:
						# When no payment term defined
						self.date_due = self.date_due or self.date_invoice
					else:
						pterm = self.payment_term_id
						pterm_list = pterm.with_context(currency_id=self.company_id.currency_id.id).compute(value=1, date_ref=date_invoice)[0]
						self.date_due = max(line[0] for line in pterm_list)
			*/
			return new(pool.AccountInvoiceData), []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().ActionInvoiceDraft().DeclareMethod(
		`ActionInvoiceDraft`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_invoice_draft(self):
			  if self.filtered(lambda inv: inv.state != 'cancel'):
			      raise UserError(_("Invoice must be cancelled in order to reset it to draft."))
			  # go from canceled state to draft state
			  self.write({'state': 'draft', 'date': False})
			  # Delete former printed invoice
			  try:
			      report_invoice = self.env['report']._get_report_from_name('account.report_invoice')
			  except IndexError:
			      report_invoice = False
			  if report_invoice and report_invoice.attachment:
			      for invoice in self:
			          with invoice.env.do_in_draft():
			              invoice.number, invoice.state = invoice.move_name, 'open'
			              attachment = self.env['report']._attachment_stored(invoice, report_invoice)[invoice.id]
			          if attachment:
			              attachment.unlink()
			  return True

			*/
			return true
		})

	pool.AccountInvoice().Methods().ActionInvoiceProforma2().DeclareMethod(
		`ActionInvoiceProforma2`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_invoice_proforma2(self):
			  if self.filtered(lambda inv: inv.state != 'draft'):
			      raise UserError(_("Invoice must be a draft in order to set it to Pro-forma."))
			  return self.write({'state': 'proforma2'})

			*/
			return true
		})

	pool.AccountInvoice().Methods().ActionInvoiceOpen().DeclareMethod(
		`ActionInvoiceOpen`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_invoice_open(self):
			  # lots of duplicate calls to action_invoice_open, so we remove those already open
			  to_open_invoices = self.filtered(lambda inv: inv.state != 'open')
			  if to_open_invoices.filtered(lambda inv: inv.state not in ['proforma2', 'draft']):
			      raise UserError(_("Invoice must be in draft or Pro-forma state in order to validate it."))
			  to_open_invoices.action_date_assign()
			  to_open_invoices.action_move_create()
			  return to_open_invoices.invoice_validate()

			*/
			return true
		})

	pool.AccountInvoice().Methods().ActionInvoicePaid().DeclareMethod(
		`ActionInvoicePaid`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_invoice_paid(self):
			  # lots of duplicate calls to action_invoice_paid, so we remove those already paid
			  to_pay_invoices = self.filtered(lambda inv: inv.state != 'paid')
			  if to_pay_invoices.filtered(lambda inv: inv.state != 'open'):
			      raise UserError(_('Invoice must be validated in order to set it to register payment.'))
			  if to_pay_invoices.filtered(lambda inv: not inv.reconciled):
			      raise UserError(_('You cannot pay an invoice which is partially paid. You need to reconcile payment entries first.'))
			  return to_pay_invoices.write({'state': 'paid'})

			*/
			return true
		})

	pool.AccountInvoice().Methods().ActionInvoiceReOpen().DeclareMethod(
		`ActionInvoiceReOpen`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_invoice_re_open(self):
			  if self.filtered(lambda inv: inv.state != 'paid'):
			      raise UserError(_('Invoice must be paid in order to set it to register payment.'))
			  return self.write({'state': 'open'})

			*/
			return true
		})

	pool.AccountInvoice().Methods().ActionInvoiceCancel().DeclareMethod(
		`ActionInvoiceCancel`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_invoice_cancel(self):
			  if self.filtered(lambda inv: inv.state not in ['proforma2', 'draft', 'open']):
			      raise UserError(_("Invoice must be in draft, Pro-forma or open state in order to be cancelled."))
			  return self.action_cancel()

			*/
			return true
		})

	pool.AccountInvoice().Methods().GetFormviewId().Extend("",
		func(rs pool.AccountInvoiceSet) string {
			//@api.multi
			/*def get_formview_id(self):
			  """ Update form view id of action to open the invoice """
			  if self.type in ('in_invoice', 'in_refund'):
			      return self.env.ref('account.invoice_supplier_form').id
			  else:
			      return self.env.ref('account.invoice_form').id

			*/
			return ""
		})

	pool.AccountInvoice().Methods().PrepareTaxLineVals().DeclareMethod(
		`PrepareTaxLineVals`,
		func(rs pool.AccountInvoiceSet, line pool.AccountInvoiceLineSet, tax pool.AccountTaxSet) *pool.AccountInvoiceTaxData {
			/*def _prepare_tax_line_vals(self, line, tax):
			  """ Prepare values to create an account.invoice.tax line

			  The line parameter is an account.invoice.line, and the
			  tax parameter is the output of account.tax.compute_all().
			  """
			  vals = {
			      'invoice_id': self.id,
			      'name': tax['name'],
			      'tax_id': tax['id'],
			      'amount': tax['amount'],
			      'base': tax['base'],
			      'manual': False,
			      'sequence': tax['sequence'],
			      'account_analytic_id': tax['analytic'] and line.account_analytic_id.id or False,
			      'account_id': self.type in ('out_invoice', 'in_invoice') and (tax['account_id'] or line.account_id.id) or (tax['refund_account_id'] or line.account_id.id),
			  }

			  # If the taxes generate moves on the same financial account as the invoice line,
			  # propagate the analytic account from the invoice line to the tax line.
			  # This is necessary in situations were (part of) the taxes cannot be reclaimed,
			  # to ensure the tax move is allocated to the proper analytic account.
			  if not vals.get('account_analytic_id') and line.account_analytic_id and vals['account_id'] == line.account_id.id:
			      vals['account_analytic_id'] = line.account_analytic_id.id

			  return vals

			*/
			return new(pool.AccountInvoiceTaxData)
		})

	pool.AccountInvoice().Methods().GetTaxesValues().DeclareMethod(
		`GetTaxesValues`,
		func(rs pool.AccountInvoiceSet) map[string]*pool.AccountInvoiceTaxData {
			//@api.multi
			/*def get_taxes_values(self):
			  tax_grouped = {}
			  for line in self.invoice_line_ids:
			      price_unit = line.price_unit * (1 - (line.discount or 0.0) / 100.0)
			      taxes = line.invoice_line_tax_ids.compute_all(price_unit, self.currency_id, line.quantity, line.product_id, self.partner_id)['taxes']
			      for tax in taxes:
			          val = self._prepare_tax_line_vals(line, tax)
			          key = self.env['account.tax'].browse(tax['id']).get_grouping_key(val)

			          if key not in tax_grouped:
			              tax_grouped[key] = val
			          else:
			              tax_grouped[key]['amount'] += val['amount']
			              tax_grouped[key]['base'] += val['base']
			  return tax_grouped

			*/
			return make(map[string]*pool.AccountInvoiceTaxData)
		})

	pool.AccountInvoice().Methods().RegisterPayment().DeclareMethod(
		`RegisterPayment`,
		func(rs pool.AccountInvoiceSet, paymentLine pool.AccountMoveLineSet, writeOffAccount pool.AccountAccountSet,
			writeOffJournal pool.AccountJournalSet) bool {
			//@api.multi
			/*def register_payment(self, payment_line, writeoff_acc_id=False, writeoff_journal_id=False):
			  """ Reconcile payable/receivable lines from the invoice with payment_line """
			  line_to_reconcile = self.env['account.move.line']
			  for inv in self:
			      line_to_reconcile += inv.move_id.line_ids.filtered(lambda r: not r.reconciled and r.account_id.internal_type in ('payable', 'receivable'))
			  return (line_to_reconcile + payment_line).reconcile(writeoff_acc_id, writeoff_journal_id)

			*/
			return true
		})

	pool.AccountInvoice().Methods().AssignOutstandingCredit().DeclareMethod(
		`AssignOutstandingCredit`,
		func(rs pool.AccountInvoiceSet, creditAML pool.AccountMoveLineSet) bool {
			//@api.multi
			/*def assign_outstanding_credit(self, credit_aml_id):
			  self.ensure_one()
			  credit_aml = self.env['account.move.line'].browse(credit_aml_id)
			  if not credit_aml.currency_id and self.currency_id != self.company_id.currency_id:
			      credit_aml.with_context(allow_amount_currency=True).write({
			          'amount_currency': self.company_id.currency_id.with_context(date=credit_aml.date).compute(credit_aml.balance, self.currency_id),
			          'currency_id': self.currency_id.id})
			  if credit_aml.payment_id:
			      credit_aml.payment_id.write({'invoice_ids': [(4, self.id, None)]})
			  return self.register_payment(credit_aml)

			*/
			return true
		})

	pool.AccountInvoice().Methods().ActionDateAssign().DeclareMethod(
		`ActionDateAssign`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_date_assign(self):
			  for inv in self:
			      # Here the onchange will automatically write to the database
			      inv._onchange_payment_term_date_invoice()
			  return True

			*/
			return true
		})

	pool.AccountInvoice().Methods().FinalizeInvoiceMoveLines().DeclareMethod(
		`FinalizeInvoiceMoveLines is a hook method to be overridden in additional modules to verify and
		possibly alter the move lines to be created by an invoice, for special cases.`,
		func(rs pool.AccountInvoiceSet, moveLines pool.AccountMoveLineSet) pool.AccountMoveLineSet {
			return moveLines
		})

	pool.AccountInvoice().Methods().ComputeInvoiceTotals().DeclareMethod(
		`ComputeInvoiceTotals`,
		func(rs pool.AccountInvoiceSet, companyCurrency pool.CurrencySet, invoiceMoveLines pool.AccountMoveLineSet) {
			//@api.multi
			/*def compute_invoice_totals(self, company_currency, invoice_move_lines):
						  total = 0
						  total_currency = 0
						  for line in invoice_move_lines:
						      if self.currency_id != company_currency:
						          currency = self.currency_id.with_context(date=self.date or self.date_invoice or fields.Date.context_today(self))
			                  if not (line.get('currency_id') and line.get('amount_currency')):
			                      line['currency_id'] = currency.id
			                      line['amount_currency'] = currency.round(line['price'])
			                      line['price'] = currency.compute(line['price'], company_currency)
			              else:
			                  line['currency_id'] = False
			                  line['amount_currency'] = False
			                  line['price'] = self.currency_id.round(line['price'])
			              if self.type in ('out_invoice', 'in_refund'):
			                  total += line['price']
			                  total_currency += line['amount_currency'] or line['price']
			                  line['price'] = - line['price']
			              else:
			                  total -= line['price']
			                  total_currency -= line['amount_currency'] or line['price']
			          return total, total_currency, invoice_move_lines*/
		})

	pool.AccountInvoice().Methods().InvoiceLineMoveLineGet().DeclareMethod(
		`InvoiceLineMoveLineGet`,
		func(rs pool.AccountInvoiceSet) []*pool.AccountInvoiceLineData {
			//@api.model
			/*def invoice_line_move_line_get(self):
			  res = []
			  for line in self.invoice_line_ids:
			      if line.quantity==0:
			          continue
			      tax_ids = []
			      for tax in line.invoice_line_tax_ids:
			          tax_ids.append((4, tax.id, None))
			          for child in tax.children_tax_ids:
			              if child.type_tax_use != 'none':
			                  tax_ids.append((4, child.id, None))
			      analytic_tag_ids = [(4, analytic_tag.id, None) for analytic_tag in line.analytic_tag_ids]

			      move_line_dict = {
			          'invl_id': line.id,
			          'type': 'src',
			          'name': line.name.split('\n')[0][:64],
			          'price_unit': line.price_unit,
			          'quantity': line.quantity,
			          'price': line.price_subtotal,
			          'account_id': line.account_id.id,
			          'product_id': line.product_id.id,
			          'uom_id': line.uom_id.id,
			          'account_analytic_id': line.account_analytic_id.id,
			          'tax_ids': tax_ids,
			          'invoice_id': self.id,
			          'analytic_tag_ids': analytic_tag_ids
			      }
			      if line['account_analytic_id']:
			          move_line_dict['analytic_line_ids'] = [(0, 0, line._get_analytic_line())]
			      res.append(move_line_dict)
			  return res

			*/
			return []*pool.AccountInvoiceLineData{}
		})

	pool.AccountInvoice().Methods().TaxLineMoveLineGet().DeclareMethod(
		`TaxLineMoveLineGet`,
		func(rs pool.AccountInvoiceSet) []*pool.AccountInvoiceLineData {
			//@api.model
			/*def tax_line_move_line_get(self):
			  res = []
			  # keep track of taxes already processed
			  done_taxes = []
			  # loop the invoice.tax.line in reversal sequence
			  for tax_line in sorted(self.tax_line_ids, key=lambda x: -x.sequence):
			      if tax_line.amount:
			          tax = tax_line.tax_id
			          if tax.amount_type == "group":
			              for child_tax in tax.children_tax_ids:
			                  done_taxes.append(child_tax.id)
			          res.append({
			              'invoice_tax_line_id': tax_line.id,
			              'tax_line_id': tax_line.tax_id.id,
			              'type': 'tax',
			              'name': tax_line.name,
			              'price_unit': tax_line.amount,
			              'quantity': 1,
			              'price': tax_line.amount,
			              'account_id': tax_line.account_id.id,
			              'account_analytic_id': tax_line.account_analytic_id.id,
			              'invoice_id': self.id,
			              'tax_ids': [(6, 0, list(done_taxes))] if tax_line.tax_id.include_base_amount else []
			          })
			          done_taxes.append(tax.id)
			  return res

			*/
			return []*pool.AccountInvoiceLineData{}
		})

	pool.AccountInvoice().Methods().InvLineCharacteristicHashcode().DeclareMethod(
		`InvLineCharacteristicHashcode`,
		func(rs pool.AccountInvoiceSet, invoiceLine *pool.AccountInvoiceLineData) string {
			/*def inv_line_characteristic_hashcode(self, invoice_line):
			  """Overridable hashcode generation for invoice lines. Lines having the same hashcode
			  will be grouped together if the journal has the 'group line' option. Of course a module
			  can add fields to invoice lines that would need to be tested too before merging lines
			  or not."""
			  return "%s-%s-%s-%s-%s-%s-%s" % (
			      invoice_line['account_id'],
			      invoice_line.get('tax_ids', 'False'),
			      invoice_line.get('tax_line_id', 'False'),
			      invoice_line.get('product_id', 'False'),
			      invoice_line.get('analytic_account_id', 'False'),
			      invoice_line.get('date_maturity', 'False'),
			      invoice_line.get('analytic_tag_ids', 'False'),
			  )

			*/
			return ""
		})

	pool.AccountInvoice().Methods().GroupLines().DeclareMethod(
		`GroupLines`,
		func(rs pool.AccountInvoiceSet, iml []*pool.AccountInvoiceLineData, line *pool.AccountMoveLineData) []*pool.AccountMoveLineData {
			/*def group_lines(self, iml, line):
			  """Merge account move lines (and hence analytic lines) if invoice line hashcodes are equals"""
			  if self.journal_id.group_invoice_lines:
			      line2 = {}
			      for x, y, l in line:
			          tmp = self.inv_line_characteristic_hashcode(l)
			          if tmp in line2:
			              am = line2[tmp]['debit'] - line2[tmp]['credit'] + (l['debit'] - l['credit'])
			              line2[tmp]['debit'] = (am > 0) and am or 0.0
			              line2[tmp]['credit'] = (am < 0) and -am or 0.0
			              line2[tmp]['amount_currency'] += l['amount_currency']
			              line2[tmp]['analytic_line_ids'] += l['analytic_line_ids']
			              qty = l.get('quantity')
			              if qty:
			                  line2[tmp]['quantity'] = line2[tmp].get('quantity', 0.0) + qty
			          else:
			              line2[tmp] = l
			      line = []
			      for key, val in line2.items():
			          line.append((0, 0, val))
			  return line

			*/
			return []*pool.AccountMoveLineData{}
		})

	pool.AccountInvoice().Methods().ActionMoveCreate().DeclareMethod(
		`ActionMoveCreate creates invoice related analytics and financial move lines`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_move_create(self):
			  """ Creates invoice related analytics and financial move lines """
			  account_move = self.env['account.move']

			  for inv in self:
			      if not inv.journal_id.sequence_id:
			          raise UserError(_('Please define sequence on the journal related to this invoice.'))
			      if not inv.invoice_line_ids:
			          raise UserError(_('Please create some invoice lines.'))
			      if inv.move_id:
			          continue

			      ctx = dict(self._context, lang=inv.partner_id.lang)

			      if not inv.date_invoice:
			          inv.with_context(ctx).write({'date_invoice': fields.Date.context_today(self)})
				company_currency = inv.company_id.currency_id

				# create move lines (one per invoice line + eventual taxes and analytic lines)
				iml = inv.invoice_line_move_line_get()
				iml += inv.tax_line_move_line_get()

				diff_currency = inv.currency_id != company_currency
				# create one move line for the total and possibly adjust the other lines amount
				total, total_currency, iml = inv.with_context(ctx).compute_invoice_totals(company_currency, iml)

				name = inv.name or '/'
				if inv.payment_term_id:
					totlines = inv.with_context(ctx).payment_term_id.with_context(currency_id=company_currency.id).compute(total, inv.date_invoice)[0]
					res_amount_currency = total_currency
					ctx['date'] = inv.date or inv.date_invoice
					for i, t in enumerate(totlines):
						if inv.currency_id != company_currency:
							amount_currency = company_currency.with_context(ctx).compute(t[1], inv.currency_id)
						else:
							amount_currency = False

						# last line: add the diff
						res_amount_currency -= amount_currency or 0
						if i + 1 == len(totlines):
							amount_currency += res_amount_currency

						iml.append({
							'type': 'dest',
							'name': name,
							'price': t[1],
							'account_id': inv.account_id.id,
							'date_maturity': t[0],
							'amount_currency': diff_currency and amount_currency,
							'currency_id': diff_currency and inv.currency_id.id,
							'invoice_id': inv.id
						})
				else:
					iml.append({
						'type': 'dest',
						'name': name,
						'price': total,
						'account_id': inv.account_id.id,
						'date_maturity': inv.date_due,
						'amount_currency': diff_currency and total_currency,
						'currency_id': diff_currency and inv.currency_id.id,
						'invoice_id': inv.id
					})
				part = self.env['res.partner']._find_accounting_partner(inv.partner_id)
				line = [(0, 0, self.line_get_convert(l, part.id)) for l in iml]
				line = inv.group_lines(iml, line)

				journal = inv.journal_id.with_context(ctx)
				line = inv.finalize_invoice_move_lines(line)

				date = inv.date or inv.date_invoice
				move_vals = {
					'ref': inv.reference,
					'line_ids': line,
					'journal_id': journal.id,
					'date': date,
					'narration': inv.comment,
				}
				ctx['company_id'] = inv.company_id.id
				ctx['invoice'] = inv
				ctx_nolang = ctx.copy()
				ctx_nolang.pop('lang', None)
				move = account_move.with_context(ctx_nolang).create(move_vals)
				# Pass invoice in context in method post: used if you want to get the same
				# account move reference when creating the same invoice after a cancelled one:
				move.post()
				# make the invoice point to that move
				vals = {
					'move_id': move.id,
					'date': date,
					'move_name': move.name,
				}
				inv.with_context(ctx).write(vals)
			return True*/
			return true
		})

	pool.AccountInvoice().Methods().InvoiceValidate().DeclareMethod(
		`InvoiceValidate`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def invoice_validate(self):
			  for invoice in self:
			      #refuse to validate a vendor bill/refund if there already exists one with the same reference for the same partner,
			      #because it's probably a double encoding of the same bill/refund
			      if invoice.type in ('in_invoice', 'in_refund') and invoice.reference:
			          if self.search([('type', '=', invoice.type), ('reference', '=', invoice.reference), ('company_id', '=', invoice.company_id.id), ('commercial_partner_id', '=', invoice.commercial_partner_id.id), ('id', '!=', invoice.id)]):
			              raise UserError(_("Duplicated vendor reference detected. You probably encoded twice the same vendor bill/refund."))
			  return self.write({'state': 'open'})

			*/
			return true
		})

	pool.AccountInvoice().Methods().LineGetConvert().DeclareMethod(
		`LineGetConvert`,
		func(rs pool.AccountInvoiceSet, line *pool.AccountInvoiceLineData, partner pool.PartnerSet) *pool.AccountMoveLineData {
			//@api.model
			/*def line_get_convert(self, line, part):
			  return {
			      'date_maturity': line.get('date_maturity', False),
			      'partner_id': part,
			      'name': line['name'],
			      'debit': line['price'] > 0 and line['price'],
			      'credit': line['price'] < 0 and -line['price'],
			      'account_id': line['account_id'],
			      'analytic_line_ids': line.get('analytic_line_ids', []),
			      'amount_currency': line['price'] > 0 and abs(line.get('amount_currency', False)) or -abs(line.get('amount_currency', False)),
			      'currency_id': line.get('currency_id', False),
			      'quantity': line.get('quantity', 1.00),
			      'product_id': line.get('product_id', False),
			      'product_uom_id': line.get('uom_id', False),
			      'analytic_account_id': line.get('account_analytic_id', False),
			      'invoice_id': line.get('invoice_id', False),
			      'tax_ids': line.get('tax_ids', False),
			      'tax_line_id': line.get('tax_line_id', False),
			      'analytic_tag_ids': line.get('analytic_tag_ids', False),
			  }

			*/
			return new(pool.AccountMoveLineData)
		})

	pool.AccountInvoice().Methods().ActionCancel().DeclareMethod(
		`ActionCancel`,
		func(rs pool.AccountInvoiceSet) bool {
			//@api.multi
			/*def action_cancel(self):
			  moves = self.env['account.move']
			  for inv in self:
			      if inv.move_id:
			          moves += inv.move_id
			      if inv.payment_move_line_ids:
			          raise UserError(_('You cannot cancel an invoice which is partially paid. You need to unreconcile related payment entries first.'))

			  # First, set the invoices as cancelled and detach the move ids
			  self.write({'state': 'cancel', 'move_id': False})
			  if moves:
			      # second, invalidate the move(s)
			      moves.button_cancel()
			      # delete the move this invoice was pointing to
			      # Note that the corresponding move_lines and move_reconciles
			      # will be automatically deleted too
			      moves.unlink()
			  return True
			*/
			return true
		})

	pool.AccountInvoice().Methods().NameGet().Extend("",
		func(rs pool.AccountInvoiceSet) string {
			//@api.multi
			/*def name_get(self):
			  TYPES = {
			      'out_invoice': _('Invoice'),
			      'in_invoice': _('Vendor Bill'),
			      'out_refund': _('Refund'),
			      'in_refund': _('Vendor Refund'),
			  }
			  result = []
			  for inv in self:
			      result.append((inv.id, "%s %s" % (inv.number or TYPES[inv.type], inv.name or '')))
			  return result

			*/
			return rs.Super().NameGet()
		})

	pool.AccountInvoice().Methods().SearchByName().Extend("",
		func(rs pool.AccountInvoiceSet, name string, op operator.Operator, additionalCond pool.AccountInvoiceCondition, limit int) pool.AccountInvoiceSet {
			//@api.model
			/*def name_search(self, name, args=None, operator='ilike', limit=100):
			  args = args or []
			  recs = self.browse()
			  if name:
			      recs = self.search([('number', '=', name)] + args, limit=limit)
			  if not recs:
			      recs = self.search([('name', operator, name)] + args, limit=limit)
			  return recs.name_get()

			*/
			return rs.Super().SearchByName(name, op, additionalCond, limit)
		})

	pool.AccountInvoice().Methods().GetRefundCommonFields().DeclareMethod(
		`GetRefundCommonFields`,
		func(rs pool.AccountInvoiceSet) []models.FieldNamer {
			/*def _get_refund_common_fields(self):
			  return ['partner_id', 'payment_term_id', 'account_id', 'currency_id', 'journal_id']

			*/
			return []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().GetRefundPrepareFields().DeclareMethod(
		`GetRefundPrepareFields`,
		func(rs pool.AccountInvoiceSet) []models.FieldNamer {
			/*def _get_refund_prepare_fields(self):
			  return ['name', 'reference', 'comment', 'date_due']

			*/
			return []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().GetRefundModifyReadFields().DeclareMethod(
		`GetRefundModifyReadFields`,
		func(rs pool.AccountInvoiceSet) []models.FieldNamer {
			/*def _get_refund_modify_read_fields(self):
			  read_fields = ['type', 'number', 'invoice_line_ids', 'tax_line_ids', 'date']
			  return self._get_refund_common_fields() + self._get_refund_prepare_fields() + read_fields

			*/
			return []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().GetRefundCopyFields().DeclareMethod(
		`GetRefundCopyFields`,
		func(rs pool.AccountInvoiceSet) []models.FieldNamer {
			/*def _get_refund_copy_fields(self):
			  copy_fields = ['company_id', 'user_id', 'fiscal_position_id']
			  return self._get_refund_common_fields() + self._get_refund_prepare_fields() + copy_fields

			*/
			return []models.FieldNamer{}
		})

	pool.AccountInvoice().Methods().PrepareRefund().DeclareMethod(
		`PrepareRefund`,
		func(rs pool.AccountInvoiceSet, invoice pool.AccountInvoiceSet, dateInvoice, date dates.Date,
			description string, journal pool.AccountJournalSet) *pool.AccountInvoiceData {
			//@api.model
			/*def _prepare_refund(self, invoice, date_invoice=None, date=None, description=None, journal_id=None):
						  """ Prepare the dict of values to create the new refund from the invoice.
						      This method may be overridden to implement custom
						      refund generation (making sure to call super() to establish
						      a clean extension chain).

						      :param record invoice: invoice to refund
						      :param string date_invoice: refund creation date from the wizard
						      :param integer date: force date from the wizard
						      :param string description: description of the refund from the wizard
						      :param integer journal_id: account.journal from the wizard
						      :return: dict of value to create() the refund
						  """
						  values = {}
						  for field in self._get_refund_copy_fields():
						      if invoice._fields[field].type == 'many2one':
						          values[field] = invoice[field].id
						      else:
						          values[field] = invoice[field] or False

						  values['invoice_line_ids'] = self._refund_cleanup_lines(invoice.invoice_line_ids)

						  tax_lines = invoice.tax_line_ids
						  values['tax_line_ids'] = self._refund_cleanup_lines(tax_lines)

						  if journal_id:
						      journal = self.env['account.journal'].browse(journal_id)
						  elif invoice['type'] == 'in_invoice':
						      journal = self.env['account.journal'].search([('type', '=', 'purchase')], limit=1)
						  else:
						      journal = self.env['account.journal'].search([('type', '=', 'sale')], limit=1)
						  values['journal_id'] = journal.id

						  values['type'] = TYPE2REFUND[invoice['type']]
						  values['date_invoice'] = date_invoice or fields.Date.context_today(invoice)
						  values['state'] = 'draft'
						  values['number'] = False
						  values['origin'] = invoice.number
						  values['refund_invoice_id'] = invoice.id

						  if date:
							  values['date'] = date
						  if description:
							  values['name'] = description
			              return values*/
			return new(pool.AccountInvoiceData)
		})

	pool.AccountInvoice().Methods().Refund().DeclareMethod(
		`Refund`,
		func(rs pool.AccountInvoiceSet, dateInvoice, date dates.Date,
			description string, journal pool.AccountJournalSet) pool.AccountInvoiceSet {
			//@api.returns('self')
			/*def refund(self, date_invoice=None, date=None, description=None, journal_id=None):
			  new_invoices = self.browse()
			  for invoice in self:
			      # create the new invoice
			      values = self._prepare_refund(invoice, date_invoice=date_invoice, date=date,
			                              description=description, journal_id=journal_id)
			      refund_invoice = self.create(values)
			      invoice_type = {'out_invoice': ('customer invoices refund'),
			          'in_invoice': ('vendor bill refund')}
			      message = _("This %s has been created from: <a href=# data-oe-model=account.invoice data-oe-id=%d>%s</a>") % (invoice_type[invoice.type], invoice.id, invoice.number)
			      refund_invoice.message_post(body=message)
			      new_invoices += refund_invoice
			  return new_invoices

			*/
			return pool.AccountInvoice().NewSet(rs.Env())
		})

	pool.AccountInvoice().Methods().PayAndReconcile().DeclareMethod(
		`PayAndReconcile`,
		func(rs pool.AccountInvoiceSet, payJournal pool.AccountJournalSet, payAmount float64,
			date dates.Date, writeoffAcc pool.AccountAccountSet) bool {
			//@api.multi
			/*def pay_and_reconcile(self, pay_journal, pay_amount=None, date=None, writeoff_acc=None):
			  """ Create and post an account.payment for the invoice self, which creates a journal entry that reconciles the invoice.

			      :param pay_journal: journal in which the payment entry will be created
			      :param pay_amount: amount of the payment to register, defaults to the residual of the invoice
			      :param date: payment date, defaults to fields.Date.context_today(self)
			      :param writeoff_acc: account in which to create a writeoff if pay_amount < self.residual, so that the invoice is fully paid
				"""
				if isinstance( pay_journal, ( int, long ) ):
					pay_journal = self.env['account.journal'].browse([pay_journal])
				assert len(self) == 1, "Can only pay one invoice at a time."
				payment_type = self.type in ('out_invoice', 'in_refund') and 'inbound' or 'outbound'
				if payment_type == 'inbound':
					payment_method = self.env.ref('account.account_payment_method_manual_in')
					journal_payment_methods = pay_journal.inbound_payment_method_ids
				else:
					payment_method = self.env.ref('account.account_payment_method_manual_out')
					journal_payment_methods = pay_journal.outbound_payment_method_ids
				if payment_method not in journal_payment_methods:
					raise UserError(_('No appropriate payment method enabled on journal %s') % pay_journal.name)

				communication = self.type in ('in_invoice', 'in_refund') and self.reference or self.number
				if self.origin:
					communication = '%s (%s)' % (communication, self.origin)

				payment = self.env['account.payment'].create({
					'invoice_ids': [(6, 0, self.ids)],
					'amount': pay_amount or self.residual,
					'payment_date': date or fields.Date.context_today(self),
					'communication': communication,
					'partner_id': self.partner_id.id,
					'partner_type': self.type in ('out_invoice', 'out_refund') and 'customer' or 'supplier',
					'journal_id': pay_journal.id,
					'payment_type': payment_type,
					'payment_method_id': payment_method.id,
					'payment_difference_handling': writeoff_acc and 'reconcile' or 'open',
					'writeoff_account_id': writeoff_acc and writeoff_acc.id or False,
				})
				payment.post()

				return True
			*/
			return true
		})

	pool.AccountInvoice().Methods().GetTaxAmountByGroup().DeclareMethod(
		`GetTaxAmountByGroup`,
		func(rs pool.AccountInvoiceSet) []accounttypes.TaxGroup {
			//@api.multi
			/*def _get_tax_amount_by_group(self):
			  self.ensure_one()
			  res = {}
			  currency = self.currency_id or self.company_id.currency_id
			  for line in self.tax_line_ids:
			      res.setdefault(line.tax_id.tax_group_id, 0.0)
			      res[line.tax_id.tax_group_id] += line.amount
			  res = sorted(res.items(), key=lambda l: l[0].sequence)
			  res = map(lambda l: (l[0].name, l[1]), res)
			  return res


			*/
			return []accounttypes.TaxGroup{}
		})

	pool.AccountInvoiceLine().DeclareModel()
	pool.AccountInvoiceLine().SetDefaultOrder("Invoice", "Sequence", "ID")

	pool.AccountInvoiceLine().AddFields(map[string]models.FieldDefinition{
		"Name": models.TextField{String: "Description", Required: true},
		"Origin": models.CharField{String: "Source Document",
			Help: "Reference of the document that produced this invoice."},
		"Sequence": models.IntegerField{Default: models.DefaultValue(10),
			Help: "Gives the sequence of this line when displaying the invoice."},
		"Invoice": models.Many2OneField{String: "Invoice Reference", RelationModel: pool.AccountInvoice(),
			OnDelete: models.Cascade, Index: true},
		"Uom": models.Many2OneField{String: "Unit of Measure", RelationModel: pool.ProductUom(),
			OnDelete: models.SetNull, Index: true,
			OnChange: pool.AccountInvoiceLine().Methods().OnchangeUom()},
		"Product": models.Many2OneField{String: "Product", RelationModel: pool.ProductProduct(),
			OnDelete: models.Restrict, Index: true,
			OnChange: pool.AccountInvoiceLine().Methods().OnchangeProduct()},
		"Account": models.Many2OneField{String: "Account", RelationModel: pool.AccountAccount(),
			Required: true, Filter: pool.AccountAccount().Deprecated().Equals(false),
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				if !env.Context().HasKey("journal_id") {
					return pool.AccountJournal().NewSet(env)
				}
				journal := pool.AccountJournal().Browse(env, []int64{env.Context().GetInteger("journal_id")})
				if env.Context().GetString("type") == "out_invoice" || env.Context().GetString("type") == "in_refund" {
					return journal.DefaultCreditAccount()
				}
				return journal.DefaultDebitAccount()
			}, Help: "The income or expense account related to the selected product.",
			OnChange: pool.AccountInvoiceLine().Methods().OnchangeAccount()},
		"PriceUnit": models.FloatField{String: "Unit Price", Required: true,
			Digits: decimalPrecision.GetPrecision("Product Price")},
		"PriceSubtotal": models.FloatField{String: "Amount", Stored: true,
			Compute: pool.AccountInvoiceLine().Methods().ComputePrice(),
			Depends: []string{"PriceUnit", "Discount", "InvoiceLineTaxes", "Quantity", "Product", "Invoice.Partner",
				"Invoice.Currency", "Invoice.Company", "Invoice.DateInvoice"}},
		"PriceSubtotalSigned": models.FloatField{String: "Amount Signed", /*[ currency_field 'company_currency_id']*/
			Stored: true, Compute: pool.AccountInvoiceLine().Methods().ComputePrice(),
			Depends: []string{"PriceUnit", "Discount", "InvoiceLineTaxes", "Quantity", "Product", "Invoice.Partner",
				"Invoice.Currency", "Invoice.Company", "Invoice.DateInvoice"},
			Help: "Total amount in the currency of the company, negative for credit notes."},
		"Quantity": models.FloatField{Digits: decimalPrecision.GetPrecision("Product Unit of Measure"),
			Required: true, Default: models.DefaultValue(1)},
		"Discount": models.FloatField{String: "Discount (%)", Digits: decimalPrecision.GetPrecision("Discount"),
			Default: models.DefaultValue(0.0)},
		"InvoiceLineTaxes": models.Many2ManyField{String: "Taxes", RelationModel: pool.AccountTax(),
			JSON: "invoice_line_tax_ids", Filter: pool.AccountTax().TypeTaxUse().NotEquals("none").AndCond(
				pool.AccountTax().Active().Equals(true).Or().Active().Equals(false))},
		"AccountAnalytic": models.Many2OneField{String: "Analytic Account", RelationModel: pool.AccountAnalyticAccount()},
		"AnalyticTags":    models.Many2ManyField{RelationModel: pool.AccountAnalyticTag(), JSON: "analytic_tag_ids"},
		"Company":         models.Many2OneField{RelationModel: pool.Company(), Related: "Invoice.Company" /* readonly=true */},
		"Partner": models.Many2OneField{String: "Partner", RelationModel: pool.Partner(),
			Related: "Invoice.Partner" /* readonly=true */},
		"Currency":        models.Many2OneField{RelationModel: pool.Currency(), Related: "Invoice.Currency"},
		"CompanyCurrency": models.Many2OneField{RelationModel: pool.Currency(), Related: "Invoice.CompanyCurrency" /* readonly=true */},
	})

	pool.AccountInvoiceLine().Methods().GetAnalyticLine().DeclareMethod(
		`GetAnalyticLine`,
		func(rs pool.AccountInvoiceLineSet) *pool.AccountAnalyticLineData {
			//@api.multi
			/*def _get_analytic_line(self):
			  ref = self.invoice_id.number
			  return {
			      'name': self.name,
			      'date': self.invoice_id.date_invoice,
			      'account_id': self.account_analytic_id.id,
			      'unit_amount': self.quantity,
			      'amount': self.price_subtotal_signed,
			      'product_id': self.product_id.id,
			      'product_uom_id': self.uom_id.id,
			      'general_account_id': self.account_id.id,
			      'ref': ref,
			  }

			*/
			return new(pool.AccountAnalyticLineData)
		})

	pool.AccountInvoiceLine().Methods().ComputePrice().DeclareMethod(
		`ComputePrice`,
		func(rs pool.AccountInvoiceLineSet) (*pool.AccountInvoiceLineData, []models.FieldNamer) {
			/*def _compute_price(self):
			  currency = self.invoice_id and self.invoice_id.currency_id or None
			  price = self.price_unit * (1 - (self.discount or 0.0) / 100.0)
			  taxes = False
			  if self.invoice_line_tax_ids:
			      taxes = self.invoice_line_tax_ids.compute_all(price, currency, self.quantity, product=self.product_id, partner=self.invoice_id.partner_id)
			  self.price_subtotal = price_subtotal_signed = taxes['total_excluded'] if taxes else self.quantity * price
			  if self.invoice_id.currency_id and self.invoice_id.company_id and self.invoice_id.currency_id != self.invoice_id.company_id.currency_id:
			      price_subtotal_signed = self.invoice_id.currency_id.with_context(date=self.invoice_id.date_invoice).compute(price_subtotal_signed, self.invoice_id.company_id.currency_id)
			  sign = self.invoice_id.type in ['in_refund', 'out_refund'] and -1 or 1
			  self.price_subtotal_signed = price_subtotal_signed * sign

			*/
			return new(pool.AccountInvoiceLineData), []models.FieldNamer{}
		})

	pool.AccountInvoiceLine().Methods().FieldsViewGet().Extend("",
		func(rs pool.AccountInvoiceLineSet, args webdata.FieldsViewGetParams) *webdata.FieldsViewData {
			//@api.model
			/*def fields_view_get(self, view_id=None, view_type='form', toolbar=False, submenu=False):
			res = super(AccountInvoiceLine, self).fields_view_get(
				view_id=view_id, view_type=view_type, toolbar=toolbar, submenu=submenu)
			if self._context.get('type'):
				doc = etree.XML(res['arch'])
				for node in doc.xpath("//field[@name='product_id']"):
					if self._context['type'] in ('in_invoice', 'in_refund'):
						# Hack to fix the stable version 8.0 -> saas-12
						# purchase_ok will be moved from purchase to product in master #13271
						if 'purchase_ok' in self.env['product.template']._fields:
							node.set('domain', "[('purchase_ok', '=', True)]")
					else:
						node.set('domain', "[('sale_ok', '=', True)]")
				res['arch'] = etree.tostring(doc)
			return res*/
			return rs.Super().FieldsViewGet(args)
		})

	pool.AccountInvoiceLine().Methods().GetInvoiceLineAccount().DeclareMethod(
		`GetInvoiceLineAccount`,
		func(rs pool.AccountInvoiceLineSet, typ string, product pool.ProductProductSet, fPos pool.AccountFiscalPositionSet,
			company pool.CompanySet) pool.AccountAccountSet {
			//@api.v8
			/*def get_invoice_line_account(self, type, product, fpos, company):
			  accounts = product.product_tmpl_id.get_product_accounts(fpos)
			  if type in ('out_invoice', 'out_refund'):
			      return accounts['income']
			  return accounts['expense']

			*/
			return pool.AccountAccount().NewSet(rs.Env())
		})

	pool.AccountInvoiceLine().Methods().DefineTaxes().DeclareMethod(
		`DefineTaxes is used in Onchange to set taxes and price.`,
		func(rs pool.AccountInvoiceLineSet) {
			/*def _set_taxes(self):
			  """ Used in on_change to set taxes and price."""
			  if self.invoice_id.type in ('out_invoice', 'out_refund'):
			      taxes = self.product_id.taxes_id or self.account_id.tax_ids
			  else:
			      taxes = self.product_id.supplier_taxes_id or self.account_id.tax_ids

			  # Keep only taxes of the company
			  company_id = self.company_id or self.env.user.company_id
			  taxes = taxes.filtered(lambda r: r.company_id == company_id)

			  self.invoice_line_tax_ids = fp_taxes = self.invoice_id.fiscal_position_id.map_tax(taxes, self.product_id, self.invoice_id.partner_id)

			  fix_price = self.env['account.tax']._fix_tax_included_price
			  if self.invoice_id.type in ('in_invoice', 'in_refund'):
			      prec = self.env['decimal.precision'].precision_get('Product Price')
			      if not self.price_unit or float_compare(self.price_unit, self.product_id.standard_price, precision_digits=prec) == 0:
			          self.price_unit = fix_price(self.product_id.standard_price, taxes, fp_taxes)
			  else:
			      self.price_unit = fix_price(self.product_id.lst_price, taxes, fp_taxes)

			*/
		})

	pool.AccountInvoiceLine().Methods().OnchangeProduct().DeclareMethod(
		`OnchangeProduct`,
		func(rs pool.AccountInvoiceLineSet) (*pool.AccountInvoiceLineData, []models.FieldNamer) {
			//@api.onchange('product_id')
			/*def _onchange_product_id(self):
			  domain = {}
			  if not self.invoice_id:
			      return

			  part = self.invoice_id.partner_id
			  fpos = self.invoice_id.fiscal_position_id
			  company = self.invoice_id.company_id
			  currency = self.invoice_id.currency_id
			  type = self.invoice_id.type

			  if not part:
			      warning = {
			              'title': _('Warning!'),
			              'message': _('You must first select a partner!'),
			          }
			      return {'warning': warning}

			  if not self.product_id:
			      if type not in ('in_invoice', 'in_refund'):
			          self.price_unit = 0.0
			      domain['uom_id'] = []
			  else:
			      if part.lang:
			          product = self.product_id.with_context(lang=part.lang)
			      else:
			          product = self.product_id

			      self.name = product.partner_ref
			      account = self.get_invoice_line_account(type, product, fpos, company)
			      if account:
			          self.account_id = account.id
			      self._set_taxes()

			      if type in ('in_invoice', 'in_refund'):
			          if product.description_purchase:
			              self.name += '\n' + product.description_purchase
			      else:
			          if product.description_sale:
			              self.name += '\n' + product.description_sale

			      if not self.uom_id or product.uom_id.category_id.id != self.uom_id.category_id.id:
			          self.uom_id = product.uom_id.id
			      domain['uom_id'] = [('category_id', '=', product.uom_id.category_id.id)]

			      if company and currency:
			          if company.currency_id != currency:
			              self.price_unit = self.price_unit * currency.with_context(dict(self._context or {}, date=self.invoice_id.date_invoice)).rate

			          if self.uom_id and self.uom_id.id != product.uom_id.id:
			              self.price_unit = product.uom_id._compute_price(self.price_unit, self.uom_id)
			  return {'domain': domain}

			*/
			return new(pool.AccountInvoiceLineData), []models.FieldNamer{}
		})

	pool.AccountInvoiceLine().Methods().OnchangeAccount().DeclareMethod(
		`OnchangeAccount`,
		func(rs pool.AccountInvoiceLineSet) (*pool.AccountInvoiceLineData, []models.FieldNamer) {
			//@api.onchange('account_id')
			/*def _onchange_account_id(self):
			  if not self.account_id:
			      return
			  if not self.product_id:
			      fpos = self.invoice_id.fiscal_position_id
			      self.invoice_line_tax_ids = fpos.map_tax(self.account_id.tax_ids, partner=self.partner_id).ids
			  elif not self.price_unit:
			      self._set_taxes()

			*/
			return new(pool.AccountInvoiceLineData), []models.FieldNamer{}
		})

	pool.AccountInvoiceLine().Methods().OnchangeUom().DeclareMethod(
		`OnchangeUom`,
		func(rs pool.AccountInvoiceLineSet) (*pool.AccountInvoiceLineData, []models.FieldNamer) {
			//@api.onchange('uom_id')
			/*def _onchange_uom_id(self):
			  warning = {}
			  result = {}
			  if not self.uom_id:
			      self.price_unit = 0.0
			  if self.product_id and self.uom_id:
			      if self.product_id.uom_id.category_id.id != self.uom_id.category_id.id:
			          warning = {
			              'title': _('Warning!'),
			              'message': _('The selected unit of measure is not compatible with the unit of measure of the product.'),
			          }
			          self.uom_id = self.product_id.uom_id.id
			  if warning:
			      result['warning'] = warning
			  return result

			*/
			return new(pool.AccountInvoiceLineData), []models.FieldNamer{}
		})

	pool.AccountInvoiceLine().Methods().DefineAdditionalFields().DeclareMethod(
		`DefineAdditionalFields`,
		func(rs pool.AccountInvoiceLineSet, invoice pool.AccountInvoiceSet) pool.AccountInvoiceLineSet {
			/*def _set_additional_fields(self, invoice):
			  """ Some modules, such as Purchase, provide a feature to add automatically pre-filled
			      invoice lines. However, these modules might not be aware of extra fields which are
			      added by extensions of the accounting module.
			      This method is intended to be overridden by these extensions, so that any new field can
			      easily be auto-filled as well.
			      :param invoice : account.invoice corresponding record
			      :rtype line : account.invoice.line record
			  """
			  pass

			*/
			return pool.AccountInvoiceLine().NewSet(rs.Env())
		})

	pool.AccountInvoiceLine().Methods().Unlink().Extend("",
		func(rs pool.AccountInvoiceLineSet) int64 {
			//@api.multi
			/*def unlink(self):
			  if self.filtered(lambda r: r.invoice_id and r.invoice_id.state != 'draft'):
			      raise UserError(_('You can only delete an invoice line if the invoice is in draft state.'))
			  return super(AccountInvoiceLine, self).unlink()

			*/
			return rs.Super().Unlink()
		})

	pool.AccountInvoiceTax().DeclareModel()
	pool.AccountInvoiceTax().SetDefaultOrder("Sequence")

	pool.AccountInvoiceTax().AddFields(map[string]models.FieldDefinition{
		"Invoice": models.Many2OneField{RelationModel: pool.AccountInvoice(), OnDelete: models.Cascade,
			Index: true},
		"Name": models.CharField{String: "Tax Description", Required: true},
		"Tax":  models.Many2OneField{RelationModel: pool.AccountTax(), OnDelete: models.Restrict},
		"Account": models.Many2OneField{String: "Tax Account", RelationModel: pool.AccountAccount(),
			Required: true, Filter: pool.AccountAccount().Deprecated().Equals(false)},
		"AccountAnalytic": models.Many2OneField{String: "Analytic account", RelationModel: pool.AccountAnalyticAccount()},
		"Amount":          models.FloatField{},
		"Manual":          models.BooleanField{Default: models.DefaultValue(true)},
		"Sequence":        models.IntegerField{Help: "Gives the sequence order when displaying a list of invoice tax."},
		"Company":         models.Many2OneField{RelationModel: pool.Company(), Related: "Account.Company" /* readonly=true */},
		"Currency":        models.Many2OneField{RelationModel: pool.Currency(), Related: "Invoice.Currency" /* readonly=true */},
		"Base":            models.FloatField{Compute: pool.AccountInvoiceTax().Methods().ComputeBaseAmount()},
	})

	pool.AccountInvoiceTax().Methods().ComputeBaseAmount().DeclareMethod(
		`ComputeBaseAmount`,
		func(rs pool.AccountInvoiceTaxSet) (*pool.AccountInvoiceTaxData, []models.FieldNamer) {
			/*def _compute_base_amount(self):
			  tax_grouped = {}
			  for invoice in self.mapped('invoice_id'):
			      tax_grouped[invoice.id] = invoice.get_taxes_values()
			  for tax in self:
			      tax.base = 0.0
			      if tax.tax_id:
			          key = tax.tax_id.get_grouping_key({
			              'tax_id': tax.tax_id.id,
			              'account_id': tax.account_id.id,
			              'account_analytic_id': tax.account_analytic_id.id,
			          })
			          if tax.invoice_id and key in tax_grouped[tax.invoice_id.id]:
			              tax.base = tax_grouped[tax.invoice_id.id][key]['base']
			          else:
			              _logger.warning('Tax Base Amount not computable probably due to a change in an underlying tax (%s).', tax.tax_id.name)
			*/
			return new(pool.AccountInvoiceTaxData), []models.FieldNamer{}
		})

	pool.AccountPaymentTerm().DeclareModel()
	pool.AccountPaymentTerm().SetDefaultOrder("Name")

	pool.AccountPaymentTerm().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Payment Terms", Translate: true, Required: true},
		"Active": models.BooleanField{String: "Active", Default: models.DefaultValue(true),
			Help: "If the active field is set to False, it will allow you to hide the payment term without removing it."},
		"Note": models.TextField{String: "Description on the Invoice", Translate: true},
		"Lines": models.One2ManyField{RelationModel: pool.AccountPaymentTermLine(), ReverseFK: "Payment",
			JSON: "line_ids", String: "Terms",
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.AccountPaymentTermLine().Create(env, &pool.AccountPaymentTermLineData{
					Value:       "balance",
					ValueAmount: 0,
					Sequence:    9,
					Days:        0,
					Option:      "day_after_invoice_date",
				})
			}, Constraint: pool.AccountPaymentTerm().Methods().CheckLines()},
		"Company": models.Many2OneField{RelationModel: pool.Company(), Required: true,
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.User().NewSet(env).CurrentUser().Company()
			}},
	})

	pool.AccountPaymentTerm().Methods().CheckLines().DeclareMethod(
		`CheckLines`,
		func(rs pool.AccountPaymentTermSet) {
			//@api.one
			/*def _check_lines(self):
			  payment_term_lines = self.line_ids.sorted()
			  if payment_term_lines and payment_term_lines[-1].value != 'balance':
			      raise ValidationError(_('A Payment Term should have its last line of type Balance.'))
			  lines = self.line_ids.filtered(lambda r: r.value == 'balance')
			  if len(lines) > 1:
			      raise ValidationError(_('A Payment Term should have only one line of type Balance.'))

			*/
		})

	pool.AccountPaymentTerm().Methods().Compute().DeclareMethod(
		`Compute`,
		func(rs pool.AccountPaymentTermSet, value float64, dateRef dates.Date) []accounttypes.PaymentDueDates {
			//@api.one
			/*def compute(self, value, date_ref=False):
				date_ref = date_ref or fields.Date.today()
				amount = value
				result = []
				if self.env.context.get('currency_id'):
					currency = self.env['res.currency'].browse(self.env.context['currency_id'])
				else:
					currency = self.env.user.company_id.currency_id
				prec = currency.decimal_places
				for line in self.line_ids:
					if line.value == 'fixed':
						amt = round(line.value_amount, prec)
					elif line.value == 'percent':
						amt = round(value * (line.value_amount / 100.0), prec)
					elif line.value == 'balance':
						amt = round(amount, prec)
					if amt:
						next_date = fields.Date.from_string(date_ref)
						if line.option == 'day_after_invoice_date':
							next_date += relativedelta(days=line.days)
						elif line.option == 'fix_day_following_month':
							next_first_date = next_date + relativedelta(day=1, months=1)  # Getting 1st of next month
							next_date = next_first_date + relativedelta(days=line.days - 1)
						elif line.option == 'last_day_following_month':
							next_date += relativedelta(day=31, months=1)  # Getting last day of next month
						elif line.option == 'last_day_current_month':
							next_date += relativedelta(day=31, months=0)  # Getting last day of next month
						result.append((fields.Date.to_string(next_date), amt))
						amount -= amt
				amount = reduce(lambda x, y: x + y[1], result, 0.0)
				dist = round(value - amount, prec)
				if dist:
					last_date = result and result[-1][0] or fields.Date.today()
					result.append((last_date, dist))
			return result*/
			return []accounttypes.PaymentDueDates{}
		})

	pool.AccountPaymentTerm().Methods().Unlink().Extend("",
		func(rs pool.AccountPaymentTermSet) int64 {
			//@api.multi
			/*def unlink(self):
			property_recs = self.env['ir.property'].search([('value_reference', 'in', ['account.payment.term,%s'%payment_term.id for payment_term in self])])
			property_recs.unlink()
			return super(AccountPaymentTerm, self).unlink()

			*/
			return rs.Super().Unlink()
		})

	pool.AccountPaymentTermLine().DeclareModel()
	pool.AccountPaymentTermLine().SetDefaultOrder("Sequence", "ID")

	pool.AccountPaymentTermLine().AddFields(map[string]models.FieldDefinition{
		"Value": models.SelectionField{Selection: types.Selection{
			"balance": "Balance",
			"percent": "Percent",
			"fixed":   "Fixed Amount",
		}, String: "Type", Required: true, Default: models.DefaultValue("balance"),
			Constraint: pool.AccountPaymentTermLine().Methods().CheckPercent(),
			Help:       "Select here the kind of valuation related to this payment term line."},
		"ValueAmount": models.FloatField{String: "Value", Digits: decimalPrecision.GetPrecision("Payment Terms"),
			Constraint: pool.AccountPaymentTermLine().Methods().CheckPercent(),
			Help:       "For percent enter a ratio between 0-100."},
		"Days": models.IntegerField{String: "Number of Days", Required: true, Default: models.DefaultValue(0)},
		"Option": models.SelectionField{String: "Options", Selection: types.Selection{
			"day_after_invoice_date":   "Day(s) after the invoice date",
			"fix_day_following_month":  "Day(s) after the end of the invoice month (Net EOM)",
			"last_day_following_month": "Last day of following month",
			"last_day_current_month":   "Last day of current month"},
			Default: models.DefaultValue("day_after_invoice_date"), Required: true,
			OnChange: pool.AccountPaymentTermLine().Methods().OnchangeOption()},
		"Payment": models.Many2OneField{String: "Payment Terms", RelationModel: pool.AccountPaymentTerm(),
			Required: true, Index: true, OnDelete: models.Cascade},
		"Sequence": models.IntegerField{Default: models.DefaultValue(10),
			Help: "Gives the sequence order when displaying a list of payment term lines."},
	})

	pool.AccountPaymentTermLine().Methods().CheckPercent().DeclareMethod(
		`CheckPercent`,
		func(rs pool.AccountPaymentTermLineSet) {
			//@api.constrains('value','value_amount')
			/*def _check_percent(self):
			  if self.value == 'percent' and (self.value_amount < 0.0 or self.value_amount > 100.0):
			      raise ValidationError(_('Percentages for Payment Terms Line must be between 0 and 100.'))

			*/
		})

	pool.AccountPaymentTermLine().Methods().OnchangeOption().DeclareMethod(
		`OnchangeOption`,
		func(rs pool.AccountPaymentTermLineSet) (*pool.AccountPaymentTermLineData, []models.FieldNamer) {
			//@api.onchange('option')
			/*def _onchange_option(self):
			  if self.option in ('last_day_current_month', 'last_day_following_month'):
			      self.days = 0

			*/
			return new(pool.AccountPaymentTermLineData), []models.FieldNamer{}
		})

	//pool.MailComposeMessage().Methods().SendMail().DeclareMethod(
	//	`SendMail`,
	//	func(rs pool.MailComposeMessageSet, args struct {
	//		AutoCommit interface{}
	//	}) {
	//		//@api.multi
	//		/*def send_mail(self, auto_commit=False):
	//		  context = self._context
	//		  if context.get('default_model') == 'account.invoice' and \
	//		          context.get('default_res_id') and context.get('mark_invoice_as_sent'):
	//		      invoice = self.env['account.invoice'].browse(context['default_res_id'])
	//		      invoice = invoice.with_context(mail_post_autofollow=True)
	//		      invoice.sent = True
	//		      invoice.message_post(body=_("Invoice sent"))
	//		  return super(MailComposeMessage, self).send_mail(auto_commit=auto_commit)
	//		*/
	//	})

}
