package front

import "html/template"

func FinanceExplorer() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/finance_explorer.html")...))
}
