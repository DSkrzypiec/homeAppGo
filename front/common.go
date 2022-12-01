package front

var TemplateCommonFiles []string = []string{
	"html/common/css.html",
	"html/common/header.html",
	"html/common/logo.html",
	"html/common/menu.html",
	"html/common/session_timer.html",
	"html/common/js.js",
}

func withCommonTemplates(paths ...string) []string {
	return append(paths, TemplateCommonFiles...)
}
