package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

// Template はHTMLテンプレートを利用するためのRenderer Interfaceです。
type Template struct {
}

// Render はHTMLテンプレートにデータを埋め込んだ結果をWriterに書き込みます。
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if t, ok := templates[name]; ok {
		return t.ExecuteTemplate(w, "layout.html", data)
	}
	c.Echo().Logger.Debugf("Template[%s] Not Found.", name)
	return templates["error"].ExecuteTemplate(w, "layout.html", "Internal Server Error")
}

// HTMLテンプレートの読み込み
func loadTemplates() {
	var baseTemplate = "../client/templates/layout.html"
	templates = make(map[string]*template.Template)
	// 各HTMLテンプレートに共通レイアウトを適用した結果をmapに保存する
	templates["index"] = template.Must(
		template.ParseFiles(baseTemplate, "../client/templates/index.html"))
	templates["error"] = template.Must(
		template.ParseFiles(baseTemplate, "../client/templates/error.html"))
	templates["user"] = template.Must(
		template.ParseFiles(baseTemplate, "../client/templates/user.html"))
	templates["login"] = template.Must(
		template.ParseFiles(baseTemplate, "../client/templates/login.html"))
	templates["admin"] = template.Must(
		template.ParseFiles(baseTemplate, "../client/templates/admin.html"))
	templates["admin_users"] = template.Must(
		template.ParseFiles(baseTemplate, "../client/templates/admin_users.html"))
}
