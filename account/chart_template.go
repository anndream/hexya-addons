// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package account

import (
	"github.com/hexya-erp/hexya-base/web/webdata"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/nbutils"
	"github.com/hexya-erp/hexya/pool"
)

func init() {

	pool.AccountAccountTemplate().DeclareModel()
	pool.AccountAccountTemplate().SetDefaultOrder("code")

	pool.AccountAccountTemplate().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{Required: true, Index: true},
		"Currency": models.Many2OneField{String: "Account Currency", RelationModel: pool.Currency(),
			Help: "Forces all moves for this account to have this secondary currency."},
		"Code": models.CharField{String: "Code", Size: 64, Required: true, Index: true},
		"UserType": models.Many2OneField{String: "Type", RelationModel: pool.AccountAccountType(), Required: true,
			Help: `These types are defined according to your country.
The type contains more information about the account and its specificities.`},
		"Reconcile": models.BooleanField{String: "Allow Invoices & payments Matching",
			Default: models.DefaultValue(false),
			Help:    "Check this option if you want the user to reconcile entries in this account."},
		"Note": models.TextField{},
		"Taxes": models.Many2ManyField{String: "Default Taxes", RelationModel: pool.AccountTaxTemplate(),
			JSON: "tax_ids"},
		"Nocreate": models.BooleanField{String: "Optional Create", Default: models.DefaultValue(false),
			Help: "If checked, the new chart of accounts will not contain this by default."},
		"ChartTemplate": models.Many2OneField{RelationModel: pool.AccountChartTemplate(),
			Help: `This optional field allow you to link an account template to a specific chart template that may
differ from the one its root parent belongs to. This allow you
to define chart templates that extend another and complete it with
few new accounts (You don't need to define the whole structure that
is common to both several times).`},
		"Tags": models.Many2ManyField{String: "Account tag", RelationModel: pool.AccountAccountTag(), JSON: "tag_ids",
			Help: "Optional tags you may want to assign for custom reporting"},
	})

	//pool.AccountChartTemplate().Fields().DisplayName().SetDepends([]string{"Name", "Code"})

	pool.AccountAccountTemplate().Methods().NameGet().Extend("",
		func(rs pool.AccountAccountTemplateSet) string {
			//@api.depends('name','code')
			/*def name_get(self):
			  res = []
			  for record in self:
			      name = record.name
			      if record.code:
			          name = record.code + ' ' + name
			      res.append((record.id, name))
			  return res


			*/
			return rs.Super().NameGet()
		})

	pool.AccountChartTemplate().DeclareModel()

	pool.AccountChartTemplate().AddFields(map[string]models.FieldDefinition{
		"Name":    models.CharField{Required: true},
		"Company": models.Many2OneField{RelationModel: pool.Company()},
		"Parent": models.Many2OneField{String: "Parent Chart Template",
			RelationModel: pool.AccountChartTemplate()},
		"CodeDigits": models.IntegerField{String: "# of Digits", Required: true,
			Default: models.DefaultValue(6), Help: "No. of Digits to use for account code"},
		"Visible": models.BooleanField{String: "Can be Visible?",
			Default: models.DefaultValue(true),
			Help: `Set this to False if you don't want this template to be used actively in the wizard that
generate Chart of Accounts from templates, this is useful when you want to generate
accounts of this template only when loading its child template.`},
		"Currency": models.Many2OneField{RelationModel: pool.Currency(), Required: true},
		"UseAngloSaxon": models.BooleanField{String: "Use Anglo-Saxon accounting",
			Default: models.DefaultValue(false)},
		"CompleteTaxSet": models.BooleanField{String: "Complete Set of Taxes",
			Default: models.DefaultValue(true),
			Help: `This boolean helps you to choose if you want to propose to the user to encode the sale and
purchase rates or choose from list  of taxes. This last choice assumes that the set of tax
defined on this template is complete`},
		"Accounts": models.One2ManyField{String: "Associated Account Templates",
			RelationModel: pool.AccountAccountTemplate(), ReverseFK: "ChartTemplate", JSON: "account_ids"},
		"TaxTemplates": models.One2ManyField{String: "Tax Template List",
			RelationModel: pool.AccountTaxTemplate(), ReverseFK: "ChartTemplate", JSON: "tax_template_ids",
			Help: "List of all the taxes that have to be installed by the wizard"},
		"BankAccountCodePrefix": models.CharField{String: "Prefix of the bank accounts"},
		"CashAccountCodePrefix": models.CharField{String: "Prefix of the main cash accounts"},
		"TransferAccount": models.Many2OneField{RelationModel: pool.AccountAccountTemplate(),
			Required: true, Filter: pool.AccountAccountTemplate().Reconcile().Equals(true).
					And().UserTypeFilteredOn(
				pool.AccountAccountType().HexyaExternalID().Equals("account_data_account_type_current_assets")),
			Help: "Intermediary account used when moving money from a liquidity account to another"},
		"IncomeCurrencyExchangeAccount": models.Many2OneField{String: "Gain Exchange Rate Account",
			RelationModel: pool.AccountAccountTemplate()},
		"ExpenseCurrencyExchangeAccount": models.Many2OneField{String: "Loss Exchange Rate Account",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyAccountReceivable": models.Many2OneField{String: "Receivable Account",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyAccountPayable": models.Many2OneField{String: "Payable Account",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyAccountExpenseCateg": models.Many2OneField{String: "Category of Expense Account",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyAccountIncomeCateg": models.Many2OneField{String: "Category of Income Account",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyAccountExpense": models.Many2OneField{String: "Expense Account on Product Template",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyAccountIncome": models.Many2OneField{String: "Income Account on Product Template",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyStockAccountInputCateg": models.Many2OneField{String: "Input Account for Stock Valuation",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyStockAccountOutputCateg": models.Many2OneField{String: "Output Account for Stock Valuation",
			RelationModel: pool.AccountAccountTemplate()},
		"PropertyStockValuationAccount": models.Many2OneField{String: "Account Template for Stock Valuation",
			RelationModel: pool.AccountAccountTemplate()},
	})

	pool.AccountChartTemplate().Methods().TryLoadingForCurrentCompany().DeclareMethod(
		`TryLoadingForCurrentCompany`,
		func(rs pool.AccountChartTemplateSet) *actions.Action {
			//@api.one
			/*def try_loading_for_current_company(self):
			  self.ensure_one()
			  company = self.env.user.company_id
			  # If we don't have any chart of account on this company, install this chart of account
			  if not company.chart_template_id:
			      wizard = self.env['wizard.multi.charts.accounts'].create({
			          'company_id': self.env.user.company_id.id,
			          'chart_template_id': self.id,
			          'code_digits': self.code_digits,
			          'transfer_account_id': self.transfer_account_id.id,
			          'currency_id': self.currency_id.id,
			          'bank_account_code_prefix': self.bank_account_code_prefix,
			          'cash_account_code_prefix': self.cash_account_code_prefix,
			      })
			      wizard.onchange_chart_template_id()
			      wizard.execute()

			*/
			return &actions.Action{
				Type: actions.ActionCloseWindow,
			}
		})

	pool.AccountChartTemplate().Methods().OpenSelectTemplateWizard().DeclareMethod(
		`OpenSelectTemplateWizard`,
		func(rs pool.AccountChartTemplateSet) bool {
			//@api.multi
			/*def open_select_template_wizard(self):
			  # Add action to open wizard to select between several templates
			  if not self.company_id.chart_template_id:
			      todo = self.env['ir.actions.todo']
			      action_rec = self.env['ir.model.data'].xmlid_to_object('account.action_wizard_multi_chart')
			      if action_rec:
			          todo.create({'action_id': action_rec.id, 'name': _('Choose Accounting Template'), 'type': 'automatic'})
			  return True

			*/
			return true
		})

	pool.AccountChartTemplate().Methods().GenerateJournals().DeclareMethod(
		`GenerateJournals`,
		func(rs pool.AccountChartTemplateSet, accTemplateRef map[int64]int64, company pool.CompanySet,
			journalsData []*pool.AccountJournalData) bool {
			//@api.model
			/*def generate_journals(self, acc_template_ref, company, journals_dict=None):
			  """
			  This method is used for creating journals.

			  :param chart_temp_id: Chart Template Id.
			  :param acc_template_ref: Account templates reference.
			  :param company_id: company_id selected from wizard.multi.charts.accounts.
			  :returns: True
			  """
			  JournalObj = self.env['account.journal']
			  for vals_journal in self._prepare_all_journals(acc_template_ref, company, journals_dict=journals_dict):
			      journal = JournalObj.create(vals_journal)
			      if vals_journal['type'] == 'general' and vals_journal['code'] == _('EXCH'):
			          company.write({'currency_exchange_journal_id': journal.id})
			  return True

			*/
			return true
		})

	pool.AccountChartTemplate().Methods().PrepareAllJournals().DeclareMethod(
		`PrepareAllJournals`,
		func(rs pool.AccountChartTemplateSet, accTemplateRef map[int64]int64, company pool.CompanySet,
			journalsData []*pool.AccountJournalData) []*pool.AccountJournalData {
			//@api.multi
			/*def _prepare_all_journals(self, acc_template_ref, company, journals_dict=None):
			 */
			/*def _get_default_account(journal_vals, type='debit'):
			      # Get the default accounts
			      default_account = False
			      if journal['type'] == 'sale':
			          default_account = acc_template_ref.get(self.property_account_income_categ_id.id)
			      elif journal['type'] == 'purchase':
			          default_account = acc_template_ref.get(self.property_account_expense_categ_id.id)
			      elif journal['type'] == 'general' and journal['code'] == _('EXCH'):
			          if type=='credit':
			              default_account = acc_template_ref.get(self.income_currency_exchange_account_id.id)
			          else:
			              default_account = acc_template_ref.get(self.expense_currency_exchange_account_id.id)
			      return default_account

			  journals = [{'name': _('Customer Invoices'), 'type': 'sale', 'code': _('INV'), 'favorite': True, 'sequence': 5},
			              {'name': _('Vendor Bills'), 'type': 'purchase', 'code': _('BILL'), 'favorite': True, 'sequence': 6},
			              {'name': _('Miscellaneous Operations'), 'type': 'general', 'code': _('MISC'), 'favorite': False, 'sequence': 7},
			              {'name': _('Exchange Difference'), 'type': 'general', 'code': _('EXCH'), 'favorite': False, 'sequence': 9},]
			  if journals_dict != None:
			      journals.extend(journals_dict)

			  self.ensure_one()
			  journal_data = []
			  for journal in journals:
			      vals = {
			          'type': journal['type'],
			          'name': journal['name'],
			          'code': journal['code'],
			          'company_id': company.id,
			          'default_credit_account_id': _get_default_account(journal, 'credit'),
			          'default_debit_account_id': _get_default_account(journal, 'debit'),
			          'show_on_dashboard': journal['favorite'],
			          'sequence': journal['sequence']
			      }
			      journal_data.append(vals)
			  return journal_data

			*/
			return []*pool.AccountJournalData{}
		})

	pool.AccountChartTemplate().Methods().GenerateProperties().DeclareMethod(
		`GenerateProperties`,
		func(rs pool.AccountChartTemplateSet, accTemplateRef map[int64]int64, company pool.CompanySet) bool {
			//@api.multi
			/*def generate_properties(self, acc_template_ref, company):
			  """
			  This method used for creating properties.

			  :param self: chart templates for which we need to create properties
			  :param acc_template_ref: Mapping between ids of account templates and real accounts created from them
			  :param company_id: company_id selected from wizard.multi.charts.accounts.
			  :returns: True
			  """
			  self.ensure_one()
			  PropertyObj = self.env['ir.property']
			  todo_list = [
			      ('property_account_receivable_id', 'res.partner', 'account.account'),
			      ('property_account_payable_id', 'res.partner', 'account.account'),
			      ('property_account_expense_categ_id', 'product.category', 'account.account'),
			      ('property_account_income_categ_id', 'product.category', 'account.account'),
			      ('property_account_expense_id', 'product.template', 'account.account'),
			      ('property_account_income_id', 'product.template', 'account.account'),
			  ]
			  for record in todo_list:
			      account = getattr(self, record[0])
			      value = account and 'account.account,' + str(acc_template_ref[account.id]) or False
			      if value:
			          field = self.env['ir.model.fields'].search([('name', '=', record[0]), ('model', '=', record[1]), ('relation', '=', record[2])], limit=1)
			          vals = {
			              'name': record[0],
			              'company_id': company.id,
			              'fields_id': field.id,
			              'value': value,
			          }
			          properties = PropertyObj.search([('name', '=', record[0]), ('company_id', '=', company.id)])
			          if properties:
			              #the property exist: modify it
			              properties.write(vals)
			          else:
			              #create the property
			              PropertyObj.create(vals)
			  stock_properties = [
			      'property_stock_account_input_categ_id',
			      'property_stock_account_output_categ_id',
			      'property_stock_valuation_account_id',
			  ]
			  for stock_property in stock_properties:
			      account = getattr(self, stock_property)
			      value = account and acc_template_ref[account.id] or False
			      if value:
			          company.write({stock_property: value})
			  return True

			*/
			return true
		})

	pool.AccountChartTemplate().Methods().InstallTemplate().DeclareMethod(
		`InstallTemplate`,
		func(rs pool.AccountChartTemplateSet, company pool.CompanySet, codeDigits int64, transferAccount pool.AccountAccountSet,
			objWizard pool.WizardMultiChartsAccountsSet, accRef, taxesRef map[int64]int64) (map[int64]int64, map[int64]int64) {
			//@api.multi
			/*def _install_template(self, company, code_digits=None, transfer_account_id=None, obj_wizard=None, acc_ref=None, taxes_ref=None):
			  """ Recursively load the template objects and create the real objects from them.

			      :param company: company the wizard is running for
			      :param code_digits: number of digits the accounts code should have in the COA
			      :param transfer_account_id: reference to the account template that will be used as intermediary account for transfers between 2 liquidity accounts
			      :param obj_wizard: the current wizard for generating the COA from the templates
			      :param acc_ref: Mapping between ids of account templates and real accounts created from them
			      :param taxes_ref: Mapping between ids of tax templates and real taxes created from them
			      :returns: tuple with a dictionary containing
			          * the mapping between the account template ids and the ids of the real accounts that have been generated
			            from them, as first item,
			          * a similar dictionary for mapping the tax templates and taxes, as second item,
			      :rtype: tuple(dict, dict, dict)
			  """
			  self.ensure_one()
			  if acc_ref is None:
			      acc_ref = {}
			  if taxes_ref is None:
			      taxes_ref = {}
			  if self.parent_id:
			      tmp1, tmp2 = self.parent_id._install_template(company, code_digits=code_digits, transfer_account_id=transfer_account_id, acc_ref=acc_ref, taxes_ref=taxes_ref)
			      acc_ref.update(tmp1)
			      taxes_ref.update(tmp2)
			  tmp1, tmp2 = self._load_template(company, code_digits=code_digits, transfer_account_id=transfer_account_id, account_ref=acc_ref, taxes_ref=taxes_ref)
			  acc_ref.update(tmp1)
			  taxes_ref.update(tmp2)
			  return acc_ref, taxes_ref

			*/
			return accRef, taxesRef
		})

	pool.AccountChartTemplate().Methods().LoadTemplate().DeclareMethod(
		`LoadTemplate`,
		func(rs pool.AccountChartTemplateSet, company pool.CompanySet, codeDigits int64, transferAccount pool.AccountAccountSet,
			accountRef, taxesRef map[int64]int64) (map[int64]int64, map[int64]int64) {
			//@api.multi
			/*def _load_template(self, company, code_digits=None, transfer_account_id=None, account_ref=None, taxes_ref=None):
			  """ Generate all the objects from the templates

			      :param company: company the wizard is running for
			      :param code_digits: number of digits the accounts code should have in the COA
			      :param transfer_account_id: reference to the account template that will be used as intermediary account for transfers between 2 liquidity accounts
			      :param acc_ref: Mapping between ids of account templates and real accounts created from them
			      :param taxes_ref: Mapping between ids of tax templates and real taxes created from them
			      :returns: tuple with a dictionary containing
			          * the mapping between the account template ids and the ids of the real accounts that have been generated
			            from them, as first item,
			          * a similar dictionary for mapping the tax templates and taxes, as second item,
			      :rtype: tuple(dict, dict, dict)
			  """
			  self.ensure_one()
			  if account_ref is None:
			      account_ref = {}
			  if taxes_ref is None:
			      taxes_ref = {}
			  if not code_digits:
			      code_digits = self.code_digits
			  if not transfer_account_id:
			      transfer_account_id = self.transfer_account_id
			  AccountTaxObj = self.env['account.tax']

			  # Generate taxes from templates.
			  generated_tax_res = self.tax_template_ids._generate_tax(company)
			  taxes_ref.update(generated_tax_res['tax_template_to_tax'])

			  # Generating Accounts from templates.
			  account_template_ref = self.generate_account(taxes_ref, account_ref, code_digits, company)
			  account_ref.update(account_template_ref)

			  # writing account values after creation of accounts
			  company.transfer_account_id = account_template_ref[transfer_account_id.id]
			  for key, value in generated_tax_res['account_dict'].items():
			      if value['refund_account_id'] or value['account_id']:
			          AccountTaxObj.browse(key).write({
			              'refund_account_id': account_ref.get(value['refund_account_id'], False),
			              'account_id': account_ref.get(value['account_id'], False),
			          })

			  # Create Journals - Only done for root chart template
			  if not self.parent_id:
			      self.generate_journals(account_ref, company)

			  # generate properties function
			  self.generate_properties(account_ref, company)

			  # Generate Fiscal Position , Fiscal Position Accounts and Fiscal Position Taxes from templates
			  self.generate_fiscal_position(taxes_ref, account_ref, company)

			  # Generate account operation template templates
			  self.generate_account_reconcile_model(taxes_ref, account_ref, company)

			  return account_ref, taxes_ref

			*/
			return accountRef, taxesRef
		})

	pool.AccountChartTemplate().Methods().GetAccountVals().DeclareMethod(
		`GetAccountVals`,
		func(rs pool.AccountChartTemplateSet, company pool.CompanySet, accountTemplate pool.AccountAccountSet,
			codeAcc string, taxTemplateRef map[int64]int64) *pool.AccountAccountData {
			/*def _get_account_vals(self, company, account_template, code_acc, tax_template_ref):
			  """ This method generates a dictionnary of all the values for the account that will be created.
			  """
			  self.ensure_one()
			  tax_ids = []
			  for tax in account_template.tax_ids:
			      tax_ids.append(tax_template_ref[tax.id])
			  val = {
			          'name': account_template.name,
			          'currency_id': account_template.currency_id and account_template.currency_id.id or False,
			          'code': code_acc,
			          'user_type_id': account_template.user_type_id and account_template.user_type_id.id or False,
			          'reconcile': account_template.reconcile,
			          'note': account_template.note,
			          'tax_ids': [(6, 0, tax_ids)],
			          'company_id': company.id,
			          'tag_ids': [(6, 0, [t.id for t in account_template.tag_ids])],
			      }
			  return val

			*/
			return &pool.AccountAccountData{}
		})

	pool.AccountChartTemplate().Methods().GenerateAccount().DeclareMethod(
		`GenerateAccount`,
		func(rs pool.AccountChartTemplateSet, taxTemplateRef, accTemplateRef map[int64]int64, codeDigits int,
			company pool.CompanySet) map[int64]int64 {
			//@api.multi
			/*def generate_account(self, tax_template_ref, acc_template_ref, code_digits, company):
			  """ This method for generating accounts from templates.

			      :param tax_template_ref: Taxes templates reference for write taxes_id in account_account.
			      :param acc_template_ref: dictionary with the mappping between the account templates and the real accounts.
			      :param code_digits: number of digits got from wizard.multi.charts.accounts, this is use for account code.
			      :param company_id: company_id selected from wizard.multi.charts.accounts.
			      :returns: return acc_template_ref for reference purpose.
			      :rtype: dict
			  """
			  self.ensure_one()
			  account_tmpl_obj = self.env['account.account.template']
			  acc_template = account_tmpl_obj.search([('nocreate', '!=', True), ('chart_template_id', '=', self.id)], order='id')
			  for account_template in acc_template:
			      code_main = account_template.code and len(account_template.code) or 0
			      code_acc = account_template.code or ''
			      if code_main > 0 and code_main <= code_digits:
			          code_acc = str(code_acc) + (str('0'*(code_digits-code_main)))
			      vals = self._get_account_vals(company, account_template, code_acc, tax_template_ref)
			      new_account = self.create_record_with_xmlid(company, account_template, 'account.account', vals)
			      acc_template_ref[account_template.id] = new_account
			  return acc_template_ref

			*/
			return accTemplateRef
		})

	pool.AccountChartTemplate().Methods().PrepareReconcileModelVals().DeclareMethod(
		`PrepareReconcileModelVals`,
		func(rs pool.AccountChartTemplateSet, company pool.CompanySet, accountReconcileModel pool.AccountReconcileModelTemplateSet,
			taxTemplateRef, accTemplateRef map[int64]int64) *pool.AccountReconcileModelData {
			/*def _prepare_reconcile_model_vals(self, company, account_reconcile_model, acc_template_ref, tax_template_ref):
			  """ This method generates a dictionnary of all the values for the account.reconcile.model that will be created.
			  """
			  self.ensure_one()
			  return {
			          'name': account_reconcile_model.name,
			          'sequence': account_reconcile_model.sequence,
			          'has_second_line': account_reconcile_model.has_second_line,
			          'company_id': company.id,
			          'account_id': acc_template_ref[account_reconcile_model.account_id.id],
			          'label': account_reconcile_model.label,
			          'amount_type': account_reconcile_model.amount_type,
			          'amount': account_reconcile_model.amount,
			          'tax_id': account_reconcile_model.tax_id and tax_template_ref[account_reconcile_model.tax_id.id] or False,
			          'second_account_id': account_reconcile_model.second_account_id and acc_template_ref[account_reconcile_model.second_account_id.id] or False,
			          'second_label': account_reconcile_model.second_label,
			          'second_amount_type': account_reconcile_model.second_amount_type,
			          'second_amount': account_reconcile_model.second_amount,
			          'second_tax_id': account_reconcile_model.second_tax_id and tax_template_ref[account_reconcile_model.second_tax_id.id] or False,
			      }

			*/
			return &pool.AccountReconcileModelData{}
		})

	pool.AccountChartTemplate().Methods().GenerateAccountReconcileModel().DeclareMethod(
		`GenerateAccountReconcileModel`,
		func(rs pool.AccountChartTemplateSet, taxTemplateRef, accTemplateRef map[int64]int64, company pool.CompanySet) bool {
			//@api.multi
			/*def generate_account_reconcile_model(self, tax_template_ref, acc_template_ref, company):
			  """ This method for generating accounts from templates.

			      :param tax_template_ref: Taxes templates reference for write taxes_id in account_account.
			      :param acc_template_ref: dictionary with the mappping between the account templates and the real accounts.
			      :param company_id: company_id selected from wizard.multi.charts.accounts.
			      :returns: return new_account_reconcile_model for reference purpose.
			      :rtype: dict
			  """
			  self.ensure_one()
			  account_reconcile_models = self.env['account.reconcile.model.template'].search([
			      ('account_id.chart_template_id', '=', self.id)
			  ])
			  for account_reconcile_model in account_reconcile_models:
			      vals = self._prepare_reconcile_model_vals(company, account_reconcile_model, acc_template_ref, tax_template_ref)
			      self.create_record_with_xmlid(company, account_reconcile_model, 'account.reconcile.model', vals)
			  return True

			*/
			return true
		})

	pool.AccountChartTemplate().Methods().GenerateFiscalPosition().DeclareMethod(
		`GenerateFiscalPosition`,
		func(rs pool.AccountChartTemplateSet, taxTemplateRef, accTemplateRef map[int64]int64, company pool.CompanySet) bool {
			//@api.multi
			/*def generate_fiscal_position(self, tax_template_ref, acc_template_ref, company):
			  """ This method generate Fiscal Position, Fiscal Position Accounts and Fiscal Position Taxes from templates.

			      :param chart_temp_id: Chart Template Id.
			      :param taxes_ids: Taxes templates reference for generating account.fiscal.position.tax.
			      :param acc_template_ref: Account templates reference for generating account.fiscal.position.account.
			      :param company_id: company_id selected from wizard.multi.charts.accounts.
			      :returns: True
			  """
			  self.ensure_one()
			  positions = self.env['account.fiscal.position.template'].search([('chart_template_id', '=', self.id)])
			  for position in positions:
			      new_fp = self.create_record_with_xmlid(company, position, 'account.fiscal.position', {'company_id': company.id, 'name': position.name, 'note': position.note})
			      for tax in position.tax_ids:
			          self.create_record_with_xmlid(company, tax, 'account.fiscal.position.tax', {
			              'tax_src_id': tax_template_ref[tax.tax_src_id.id],
			              'tax_dest_id': tax.tax_dest_id and tax_template_ref[tax.tax_dest_id.id] or False,
			              'position_id': new_fp
			          })
			      for acc in position.account_ids:
			          self.create_record_with_xmlid(company, acc, 'account.fiscal.position.account', {
			              'account_src_id': acc_template_ref[acc.account_src_id.id],
			              'account_dest_id': acc_template_ref[acc.account_dest_id.id],
			              'position_id': new_fp
			          })
			  return True


			*/
			return true
		})

	pool.AccountTaxTemplate().DeclareModel()

	pool.AccountTaxTemplate().AddFields(map[string]models.FieldDefinition{
		"ChartTemplate": models.Many2OneField{RelationModel: pool.AccountChartTemplate(), Required: true},
		"Name":          models.CharField{String: "Tax Name", Required: true},
		"TypeTaxUse": models.SelectionField{String: "Tax Scope", Selection: types.Selection{
			"sale":     "Sales",
			"purchase": "Purchases",
			"none":     "None",
		}, Required: true, Default: models.DefaultValue("sale"),
			Help: `Determines where the tax is selectable.
Note : 'None' means a tax can't be used by itself however it can still be used in a group.`},
		"AmountType": models.SelectionField{String: "Tax Computation",
			Selection: types.Selection{
				"group":    "Group of Taxes",
				"fixed":    "Fixed",
				"percent":  "Percentage of Price",
				"division": "Percentage of Price Tax Included",
			}, Default: models.DefaultValue("percent"), Required: true},
		"Active": models.BooleanField{Default: models.DefaultValue(true),
			Help: "Set active to false to hide the tax without removing it."},
		"Company": models.Many2OneField{RelationModel: pool.Company(), Required: true,
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.User().NewSet(env).CurrentUser().Company()
			}},
		"ChildrenTaxes": models.Many2ManyField{RelationModel: pool.AccountTaxTemplate(), JSON: "children_tax_ids",
			M2MTheirField: "ChildTax", M2MOurField: "ParentTax"},
		"Sequence": models.IntegerField{Required: true, Default: models.DefaultValue(1),
			Help: "The sequence field is used to define order in which the tax lines are applied."},
		"Amount": models.FloatField{Required: true, Digits: nbutils.Digits{Precision: 16, Scale: 4}},
		"Account": models.Many2OneField{String: "Tax Account", RelationModel: pool.AccountAccountTemplate(),
			OnDelete: models.Restrict,
			Help:     "Account that will be set on invoice tax lines for invoices. Leave empty to use the expense account."},
		"RefundAccount": models.Many2OneField{String: "Tax Account on Refunds",
			RelationModel: pool.AccountAccountTemplate(), OnDelete: models.Restrict,
			Help: "Account that will be set on invoice tax lines for refunds. Leave empty to use the expense account."},
		"Description": models.CharField{String: "Display on Invoices"},
		"PriceInclude": models.BooleanField{String: "Included in Price", Default: models.DefaultValue(false),
			Help: "Check this if the price you use on the product and invoices includes this tax."},
		"IncludeBaseAmount": models.BooleanField{String: "Affect Subsequent Taxes",
			Default: models.DefaultValue(false),
			Help:    "If set, taxes which are computed after this one will be computed based on the price tax included."},
		"Analytic": models.BooleanField{String: "Analytic Cost",
			Help: `If set, the amount computed by this tax will be assigned to
the same analytic account as the invoice line (if any)`},
		"Tags": models.Many2ManyField{String: "Account tag", RelationModel: pool.AccountAccountTag(),
			JSON: "tag_ids", Help: "Optional tags you may want to assign for custom reporting"},
		"TaxGroup":      models.Many2OneField{RelationModel: pool.AccountTaxGroup()},
		"TaxAdjustment": models.BooleanField{Default: models.DefaultValue(false)},
	})

	pool.AccountTaxTemplate().AddSQLConstraint("name_company_uniq",
		"unique(name, company_id, type_tax_use)",
		"Tax names must be unique !")

	pool.AccountTaxTemplate().Methods().NameGet().Extend("",
		func(rs pool.AccountTaxTemplateSet) string {
			//@api.depends('name','description')
			/*
				def name_get(self):
					res = []
					for record in self:
						name = record.description and record.description or record.name
						res.append((record.id, name))
					return res

			*/
			return rs.Super().NameGet()
		})

	pool.AccountTaxTemplate().Methods().GetTaxVals().DeclareMethod(
		`GetTaxVals`,
		func(rs pool.AccountTaxTemplateSet, company pool.CompanySet) *pool.AccountTaxData {
			/*def _get_tax_vals(self, company):
			  """ This method generates a dictionnary of all the values for the tax that will be created.
			  """
			  self.ensure_one()
			  val = {
			      'name': self.name,
			      'type_tax_use': self.type_tax_use,
			      'amount_type': self.amount_type,
			      'active': self.active,
			      'company_id': company.id,
			      'sequence': self.sequence,
			      'amount': self.amount,
			      'description': self.description,
			      'price_include': self.price_include,
			      'include_base_amount': self.include_base_amount,
			      'analytic': self.analytic,
			      'tag_ids': [(6, 0, [t.id for t in self.tag_ids])],
			      'tax_adjustment': self.tax_adjustment,
			  }
			  if self.tax_group_id:
			      val['tax_group_id'] = self.tax_group_id.id
			  return val

			*/
			return &pool.AccountTaxData{}
		})

	pool.AccountTaxTemplate().Methods().GenerateTax().DeclareMethod(
		`GenerateTax`,
		func(rs pool.AccountTaxTemplateSet, company pool.CompanySet) (map[int64]int64, map[int64]struct {
			AccountID       int64
			RefundAccountID int64
		}) {
			//@api.multi
			/*def _generate_tax(self, company):
			  """ This method generate taxes from templates.

			      :param company: the company for which the taxes should be created from templates in self
			      :returns: {
			          'tax_template_to_tax': mapping between tax template and the newly generated taxes corresponding,
			          'account_dict': dictionary containing a to-do list with all the accounts to assign on new taxes
			      }
			  """
			  todo_dict = {}
			  tax_template_to_tax = {}
			  for tax in self:
			      # Compute children tax ids
			      children_ids = []
			      for child_tax in tax.children_tax_ids:
			          if tax_template_to_tax.get(child_tax.id):
			              children_ids.append(tax_template_to_tax[child_tax.id])
			      vals_tax = tax._get_tax_vals(company)
			      vals_tax['children_tax_ids'] = children_ids and [(6, 0, children_ids)] or []
			      new_tax = self.env['account.chart.template'].create_record_with_xmlid(company, tax, 'account.tax', vals_tax)
			      tax_template_to_tax[tax.id] = new_tax
			      # Since the accounts have not been created yet, we have to wait before filling these fields
			      todo_dict[new_tax] = {
			          'account_id': tax.account_id.id,
			          'refund_account_id': tax.refund_account_id.id,
			      }

			  return {
			      'tax_template_to_tax': tax_template_to_tax,
			      'account_dict': todo_dict
			  }

			*/
			return make(map[int64]int64), make(map[int64]struct {
				AccountID       int64
				RefundAccountID int64
			})
		})

	pool.AccountFiscalPositionTemplate().DeclareModel()

	pool.AccountFiscalPositionTemplate().AddFields(map[string]models.FieldDefinition{
		"Name":          models.CharField{String: "Fiscal Position Template", Required: true},
		"ChartTemplate": models.Many2OneField{RelationModel: pool.AccountChartTemplate(), Required: true},
		"Accounts": models.One2ManyField{String: "Account Mapping",
			RelationModel: pool.AccountFiscalPositionAccountTemplate(), ReverseFK: "Position", JSON: "account_ids"},
		"Taxes": models.One2ManyField{String: "Tax Mapping", RelationModel: pool.AccountFiscalPositionTaxTemplate(),
			ReverseFK: "Position", JSON: "tax_ids"},
		"Note": models.TextField{String: "Notes"},
	})

	pool.AccountFiscalPositionTaxTemplate().DeclareModel()

	pool.AccountFiscalPositionTaxTemplate().AddFields(map[string]models.FieldDefinition{
		"Position": models.Many2OneField{String: "Fiscal Position", RelationModel: pool.AccountFiscalPositionTemplate(),
			Required: true, OnDelete: models.Cascade},
		"TaxSrc":  models.Many2OneField{String: "Tax Source", RelationModel: pool.AccountTaxTemplate(), Required: true},
		"TaxDest": models.Many2OneField{String: "Replacement Tax", RelationModel: pool.AccountTaxTemplate()},
	})

	pool.AccountFiscalPositionTaxTemplate().Methods().NameGet().Extend("",
		func(rs pool.AccountFiscalPositionTaxTemplateSet) string {
			return rs.Position().NameGet()
		})

	pool.AccountFiscalPositionAccountTemplate().DeclareModel()

	pool.AccountFiscalPositionAccountTemplate().AddFields(map[string]models.FieldDefinition{
		"Position": models.Many2OneField{String: "Fiscal Mapping",
			RelationModel: pool.AccountFiscalPositionTemplate(), Required: true, OnDelete: models.Cascade},
		"AccountSrc": models.Many2OneField{String: "Account Source", RelationModel: pool.AccountAccountTemplate(),
			Required: true},
		"AccountDest": models.Many2OneField{String: "Account Destination", RelationModel: pool.AccountAccountTemplate(),
			Required: true},
	})

	pool.AccountFiscalPositionAccountTemplate().Methods().NameGet().Extend("",
		func(rs pool.AccountFiscalPositionAccountTemplateSet) string {
			return rs.Position().NameGet()
		})

	pool.WizardMultiChartsAccounts().DeclareTransientModel()
	//pool.WizardMultiChartsAccounts().InheritModel(ResConfig)

	pool.WizardMultiChartsAccounts().AddFields(map[string]models.FieldDefinition{
		"Company": models.Many2OneField{RelationModel: pool.Company(), Required: true},
		"Currency": models.Many2OneField{RelationModel: pool.Currency(),
			Help: "Currency as per company's country.", Required: true},
		"OnlyOneChartTemplate": models.BooleanField{String: "Only One Chart Template Available"},
		"ChartTemplate": models.Many2OneField{String: "Chart Template",
			RelationModel: pool.AccountChartTemplate(), Required: true,
			OnChange: pool.WizardMultiChartsAccounts().Methods().OnchangeChartTemplate()},
		"BankAccounts": models.One2ManyField{String: "Cash and Banks",
			RelationModel: pool.AccountBankAccountsWizard(), ReverseFK: "BankAccount", JSON: "bank_account_ids",
			Required: true},
		"BankAccountCodePrefix": models.CharField{String: "Bank Accounts Prefix"},
		"CashAccountCodePrefix": models.CharField{String: "Cash Accounts Prefix"},
		"CodeDigits": models.IntegerField{String: "# of Digits", Required: true,
			Help: "No. of Digits to use for account code"},
		"SaleTax": models.Many2OneField{String: "Default Sales Tax",
			RelationModel: pool.AccountTaxTemplate()},
		"PurchaseTax": models.Many2OneField{String: "Default Purchase Tax",
			RelationModel: pool.AccountTaxTemplate()},
		"SaleTaxRate": models.FloatField{String: "Sales Tax(%)",
			OnChange: pool.WizardMultiChartsAccounts().Methods().OnchangeTaxRate()},
		"UseAngloSaxon": models.BooleanField{String: "Use Anglo-Saxon Accounting",
			Related: "ChartTemplate.UseAngloSaxon"},
		"TransferAccount": models.Many2OneField{RelationModel: pool.AccountAccountTemplate(),
			Required: true, Filter: pool.AccountAccountTemplate().Reconcile().Equals(true).
					And().UserTypeFilteredOn(
				pool.AccountAccountType().HexyaExternalID().Equals("account_data_account_type_current_assets")),
			Help: "Intermediary account used when moving money from a liquidity account to another"},
		"PurchaseTaxRate": models.FloatField{String: "Purchase Tax(%)"},
		"CompleteTaxSet": models.BooleanField{String: "Complete Set of Taxes",
			Help: `This boolean helps you to choose if you want to propose to the user to encode the sales and
purchase rates or use the usual m2o fields. This last choice assumes that the
set of tax defined for the chosen template is complete`},
	})

	pool.WizardMultiChartsAccounts().Methods().GetChartParentIds().DeclareMethod(
		`GetChartParentIds`,
		func(rs pool.WizardMultiChartsAccountsSet, chartTemplate pool.AccountChartTemplateSet) pool.AccountChartTemplateSet {
			//@api.model
			/*def _get_chart_parent_ids(self, chart_template):
			  """ Returns the IDs of all ancestor charts, including the chart itself.
			      (inverse of child_of operator)

			      :param browse_record chart_template: the account.chart.template record
			      :return: the IDS of all ancestor charts, including the chart itself.
			  """
			  result = [chart_template.id]
			  while chart_template.parent_id:
			      chart_template = chart_template.parent_id
			      result.append(chart_template.id)
			  return result

			*/
			return pool.AccountChartTemplate().NewSet(rs.Env())
		})

	pool.WizardMultiChartsAccounts().Methods().OnchangeTaxRate().DeclareMethod(
		`OnchangeTaxRate`,
		func(rs pool.WizardMultiChartsAccountsSet) (*pool.WizardMultiChartsAccountsData, []models.FieldNamer) {
			//@api.onchange('sale_tax_rate')
			/*def onchange_tax_rate(self):
			  self.purchase_tax_rate = self.sale_tax_rate or False

			*/
			return &pool.WizardMultiChartsAccountsData{}, []models.FieldNamer{}
		})

	pool.WizardMultiChartsAccounts().Methods().OnchangeChartTemplate().DeclareMethod(
		`OnchangeChartTemplateId`,
		func(rs pool.WizardMultiChartsAccountsSet) (*pool.WizardMultiChartsAccountsData, []models.FieldNamer) {
			//@api.onchange('chart_template_id')
			/*def onchange_chart_template_id(self):
			  res = {}
			  tax_templ_obj = self.env['account.tax.template']
			  if self.chart_template_id:
			      currency_id = self.chart_template_id.currency_id and self.chart_template_id.currency_id.id or self.env.user.company_id.currency_id.id
			      self.complete_tax_set = self.chart_template_id.complete_tax_set
			      self.currency_id = currency_id
			      if self.chart_template_id.complete_tax_set:
			      # default tax is given by the lowest sequence. For same sequence we will take the latest created as it will be the case for tax created while isntalling the generic chart of account
			          chart_ids = self._get_chart_parent_ids(self.chart_template_id)
			          base_tax_domain = [('chart_template_id', 'parent_of', chart_ids)]
			          sale_tax_domain = base_tax_domain + [('type_tax_use', '=', 'sale')]
			          purchase_tax_domain = base_tax_domain + [('type_tax_use', '=', 'purchase')]
			          sale_tax = tax_templ_obj.search(sale_tax_domain, order="sequence, id desc", limit=1)
			          purchase_tax = tax_templ_obj.search(purchase_tax_domain, order="sequence, id desc", limit=1)
			          self.sale_tax_id = sale_tax.id
			          self.purchase_tax_id = purchase_tax.id
			          res.setdefault('domain', {})
			          res['domain']['sale_tax_id'] = repr(sale_tax_domain)
			          res['domain']['purchase_tax_id'] = repr(purchase_tax_domain)
			      if self.chart_template_id.transfer_account_id:
			          self.transfer_account_id = self.chart_template_id.transfer_account_id.id
			      if self.chart_template_id.code_digits:
			          self.code_digits = self.chart_template_id.code_digits
			      if self.chart_template_id.bank_account_code_prefix:
			          self.bank_account_code_prefix = self.chart_template_id.bank_account_code_prefix
			      if self.chart_template_id.cash_account_code_prefix:
			          self.cash_account_code_prefix = self.chart_template_id.cash_account_code_prefix
			  return res

			*/
			return &pool.WizardMultiChartsAccountsData{}, []models.FieldNamer{}
		})

	pool.WizardMultiChartsAccounts().Methods().GetDefaultBankAccountIds().DeclareMethod(
		`GetDefaultBankAccountIds`,
		func(rs pool.WizardMultiChartsAccountsSet) pool.AccountBankAccountsWizardSet {
			//@api.model
			/*def _get_default_bank_account_ids(self):
			  return [{'acc_name': _('Cash'), 'account_type': 'cash'}, {'acc_name': _('Bank'), 'account_type': 'bank'}]

			*/
			return pool.AccountBankAccountsWizard().NewSet(rs.Env())
		})

	pool.WizardMultiChartsAccounts().Methods().DefaultGet().Extend("",
		func(rs pool.WizardMultiChartsAccountsSet) models.FieldMap {
			//@api.model
			/*def default_get(self, fields):
			  context = self._context or {}
			  res = super(WizardMultiChartsAccounts, self).default_get(fields)
			  tax_templ_obj = self.env['account.tax.template']
			  account_chart_template = self.env['account.chart.template']

			  if 'bank_account_ids' in fields:
			      res.update({'bank_account_ids': self._get_default_bank_account_ids()})
			  if 'company_id' in fields:
			      res.update({'company_id': self.env.user.company_id.id})
			  if 'currency_id' in fields:
			      company_id = res.get('company_id') or False
			      if company_id:
			          company = self.env['res.company'].browse(company_id)
			          currency_id = company.on_change_country(company.country_id.id)['value']['currency_id']
			          res.update({'currency_id': currency_id})

			  chart_templates = account_chart_template.search([('visible', '=', True)])
			  if chart_templates:
			      #in order to set default chart which was last created set max of ids.
			      chart_id = max(chart_templates.ids)
			      if context.get("default_charts"):
			          model_data = self.env['ir.model.data'].search_read([('model', '=', 'account.chart.template'), ('module', '=', context.get("default_charts"))], ['res_id'])
			          if model_data:
			              chart_id = model_data[0]['res_id']
			      chart = account_chart_template.browse(chart_id)
			      chart_hierarchy_ids = self._get_chart_parent_ids(chart)
			      if 'chart_template_id' in fields:
			          res.update({'only_one_chart_template': len(chart_templates) == 1,
			                      'chart_template_id': chart_id})
			      if 'sale_tax_id' in fields:
			          sale_tax = tax_templ_obj.search([('chart_template_id', 'in', chart_hierarchy_ids),
			                                                        ('type_tax_use', '=', 'sale')], limit=1, order='sequence')
			          res.update({'sale_tax_id': sale_tax and sale_tax.id or False})
			      if 'purchase_tax_id' in fields:
			          purchase_tax = tax_templ_obj.search([('chart_template_id', 'in', chart_hierarchy_ids),
			                                                            ('type_tax_use', '=', 'purchase')], limit=1, order='sequence')
			          res.update({'purchase_tax_id': purchase_tax and purchase_tax.id or False})
			  res.update({
			      'purchase_tax_rate': 15.0,
			      'sale_tax_rate': 15.0,
			  })
			  return res

			*/
			return rs.Super().DefaultGet()
		})

	pool.WizardMultiChartsAccounts().Methods().FieldsViewGet().Extend("",
		func(rs pool.WizardMultiChartsAccountsSet, args webdata.FieldsViewGetParams) *webdata.FieldsViewData {
			//@api.model
			/*def fields_view_get(self, view_id=None, view_type='form', toolbar=False, submenu=False):
			  context = self._context or {}
			  res = super(WizardMultiChartsAccounts, self).fields_view_get(view_id=view_id, view_type=view_type, toolbar=toolbar, submenu=False)
			  cmp_select = []
			  CompanyObj = self.env['res.company']

			  companies = CompanyObj.search([])
			  #display in the widget selection of companies, only the companies that haven't been configured yet (but don't care about the demo chart of accounts)
			  self._cr.execute("SELECT company_id FROM account_account WHERE deprecated = 'f' AND name != 'Chart For Automated Tests' AND name NOT LIKE '%(test)'")
			  configured_cmp = [r[0] for r in self._cr.fetchall()]
			  unconfigured_cmp = list(set(companies.ids) - set(configured_cmp))
			  for field in res['fields']:
			      if field == 'company_id':
			          res['fields'][field]['domain'] = [('id', 'in', unconfigured_cmp)]
			          res['fields'][field]['selection'] = [('', '')]
			          if unconfigured_cmp:
			              cmp_select = [(line.id, line.name) for line in CompanyObj.browse(unconfigured_cmp)]
			              res['fields'][field]['selection'] = cmp_select
			  return res

			*/
			return rs.Super().FieldsViewGet(args)
		})

	pool.WizardMultiChartsAccounts().Methods().CreateTaxTemplatesFromRates().DeclareMethod(
		`CreateTaxTemplatesFromRates`,
		func(rs pool.WizardMultiChartsAccountsSet, company pool.CompanySet) bool {
			//@api.one
			/*def _create_tax_templates_from_rates(self, company_id):
			  '''
			  This function checks if the chosen chart template is configured as containing a full set of taxes, and if
			  it's not the case, it creates the templates for account.tax object accordingly to the provided sale/purchase rates.
			  Then it saves the new tax templates as default taxes to use for this chart template.

			  :param company_id: id of the company for wich the wizard is running
			  :return: True
			  '''
			  obj_tax_temp = self.env['account.tax.template']
			  all_parents = self._get_chart_parent_ids(self.chart_template_id)
			  # create tax templates from purchase_tax_rate and sale_tax_rate fields
			  if not self.chart_template_id.complete_tax_set:
			      value = self.sale_tax_rate
			      ref_taxs = obj_tax_temp.search([('type_tax_use', '=', 'sale'), ('chart_template_id', 'in', all_parents)], order="sequence, id desc", limit=1)
			      ref_taxs.write({'amount': value, 'name': _('Tax %.2f%%') % value, 'description': '%.2f%%' % value})
			      value = self.purchase_tax_rate
			      ref_taxs = obj_tax_temp.search([('type_tax_use', '=', 'purchase'), ('chart_template_id', 'in', all_parents)], order="sequence, id desc", limit=1)
			      ref_taxs.write({'amount': value, 'name': _('Tax %.2f%%') % value, 'description': '%.2f%%' % value})
			  return True

			*/
			return true
		})

	pool.WizardMultiChartsAccounts().Methods().Execute().DeclareMethod(
		`Execute`,
		func(rs pool.WizardMultiChartsAccountsSet) bool {
			//@api.multi
			/*def execute(self):
			  '''
			  This function is called at the confirmation of the wizard to generate the COA from the templates. It will read
			  all the provided information to create the accounts, the banks, the journals, the taxes, the
			  accounting properties... accordingly for the chosen company.
			  '''
			  if len(self.env['account.account'].search([('company_id', '=', self.company_id.id)])) > 0:
			      # We are in a case where we already have some accounts existing, meaning that user has probably
			      # created its own accounts and does not need a coa, so skip installation of coa.
			      _logger.info('Could not install chart of account since some accounts already exists for the company (%s)', (self.company_id.id,))
			      return {}
			  if not self.env.user._is_admin():
			      raise AccessError(_("Only administrators can change the settings"))
			  ir_values_obj = self.env['ir.values']
			  company = self.company_id
			  self.company_id.write({'currency_id': self.currency_id.id,
			                         'accounts_code_digits': self.code_digits,
			                         'anglo_saxon_accounting': self.use_anglo_saxon,
			                         'bank_account_code_prefix': self.bank_account_code_prefix,
			                         'cash_account_code_prefix': self.cash_account_code_prefix,
			                         'chart_template_id': self.chart_template_id.id})

			  #set the coa currency to active
			  self.currency_id.write({'active': True})

			  # When we install the CoA of first company, set the currency to price types and pricelists
			  if company.id == 1:
			      for reference in ['product.list_price', 'product.standard_price', 'product.list0']:
			          try:
			              tmp2 = self.env.ref(reference).write({'currency_id': self.currency_id.id})
			          except ValueError:
			              pass

			  # If the floats for sale/purchase rates have been filled, create templates from them
			  self._create_tax_templates_from_rates(company.id)

			  # Install all the templates objects and generate the real objects
			  acc_template_ref, taxes_ref = self.chart_template_id._install_template(company, code_digits=self.code_digits, transfer_account_id=self.transfer_account_id)

			  # write values of default taxes for product as super user
			  if self.sale_tax_id and taxes_ref:
			      ir_values_obj.sudo().set_default('product.template', "taxes_id", [taxes_ref[self.sale_tax_id.id]], for_all_users=True, company_id=company.id)
			  if self.purchase_tax_id and taxes_ref:
			      ir_values_obj.sudo().set_default('product.template', "supplier_taxes_id", [taxes_ref[self.purchase_tax_id.id]], for_all_users=True, company_id=company.id)

			  # Create Bank journals
			  self._create_bank_journals_from_o2m(company, acc_template_ref)

			  # Create the current year earning account if it wasn't present in the CoA
			  account_obj = self.env['account.account']
			  unaffected_earnings_xml = self.env.ref("account.data_unaffected_earnings")
			  if unaffected_earnings_xml and not account_obj.search([('company_id', '=', company.id), ('user_type_id', '=', unaffected_earnings_xml.id)]):
			      account_obj.create({
			          'code': '999999',
			          'name': _('Undistributed Profits/Losses'),
			          'user_type_id': unaffected_earnings_xml.id,
			          'company_id': company.id,})
			  return {}

			*/
			return true
		})

	pool.WizardMultiChartsAccounts().Methods().CreateBankJournalsFromO2m().DeclareMethod(
		`CreateBankJournalsFromO2m`,
		func(rs pool.WizardMultiChartsAccountsSet, company pool.CompanySet, accTemplateRef map[int64]int64) {
			//@api.multi
			/*def _create_bank_journals_from_o2m(self, company, acc_template_ref):
			  '''
			  This function creates bank journals and its accounts for each line encoded in the field bank_account_ids of the
			  wizard (which is currently only used to create a default bank and cash journal when the CoA is installed).

			  :param company: the company for which the wizard is running.
			  :param acc_template_ref: the dictionary containing the mapping between the ids of account templates and the ids
			      of the accounts that have been generated from them.
			  '''
			  self.ensure_one()
			  # Create the journals that will trigger the account.account creation
			  for acc in self.bank_account_ids:
			      self.env['account.journal'].create({
			          'name': acc.acc_name,
			          'type': acc.account_type,
			          'company_id': company.id,
			          'currency_id': acc.currency_id.id,
			          'sequence': 10
			      })


			*/
		})

	pool.AccountBankAccountsWizard().DeclareTransientModel()

	pool.AccountBankAccountsWizard().AddFields(map[string]models.FieldDefinition{
		"AccName": models.CharField{String: "Account Name", Required: true},
		"BankAccount": models.Many2OneField{RelationModel: pool.WizardMultiChartsAccounts(),
			Required: true, OnDelete: models.Cascade},
		"Currency": models.Many2OneField{String: "Account Currency", RelationModel: pool.Currency(),
			Help: "Forces all moves for this account to have this secondary currency."},
		"AccountType": models.SelectionField{Selection: types.Selection{
			"cash": "Cash",
			"bank": "Bank",
		}},
	})

	pool.AccountReconcileModelTemplate().DeclareModel()

	pool.AccountReconcileModelTemplate().AddFields(map[string]models.FieldDefinition{
		"Name":          models.CharField{String: "Button Label", Required: true},
		"Sequence":      models.IntegerField{String: "Sequence", Required: true, Default: models.DefaultValue(10)},
		"HasSecondLine": models.BooleanField{String: "Add a second line", Default: models.DefaultValue(false)},
		"Account": models.Many2OneField{String: "Account", RelationModel: pool.AccountAccountTemplate(),
			OnDelete: models.Cascade},
		"Label": models.CharField{String: "Journal Item Label"},
		"AmountType": models.SelectionField{Selection: types.Selection{
			"fixed":      "Fixed",
			"percentage": "Percentage of balance",
		}},
		"Amount": models.FloatField{Required: true, Default: models.DefaultValue(100.0),
			Help: "Fixed amount will count as a debit if it is negative, as a credit if it is positive."},
		"Tax": models.Many2OneField{RelationModel: pool.AccountTaxTemplate(),
			OnDelete: models.Restrict, Filter: pool.AccountTaxTemplate().TypeTaxUse().Equals("purchase")},
		"SecondAccount": models.Many2OneField{RelationModel: pool.AccountAccountTemplate(),
			OnDelete: models.Cascade},
		"SecondLabel": models.CharField{String: "Second Journal Item Label"},
		"SecondAmountType": models.SelectionField{Selection: types.Selection{
			"fixed":      "Fixed",
			"percentage": "Percentage of amount",
		}},
		"SecondAmount": models.FloatField{Required: true, Default: models.DefaultValue(100.0),
			Help: "Fixed amount will count as a debit if it is negative, as a credit if it is positive."},
		"SecondTax": models.Many2OneField{String: "Second Tax", RelationModel: pool.AccountTaxTemplate(),
			OnDelete: models.Restrict, Filter: pool.AccountTaxTemplate().TypeTaxUse().Equals("purchase")},
	})

}
