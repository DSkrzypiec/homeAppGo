package front

import "html/template"

func Books() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/books.html")...))
}

func BooksNewForm() *template.Template {
	return template.Must(template.ParseFiles(withCommonTemplates("html/books_form.html")...))
}
