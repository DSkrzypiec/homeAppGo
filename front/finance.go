package front

import "html/template"

func Finance() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/finance.html")...))
}

func FinanceNewForm() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/finance_form.html")...))
}
