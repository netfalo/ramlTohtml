package main

import "os"
import "html/template"
import "strings"
//import "fmt"
import "io/ioutil"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isStandardType() (bool) {
	return true
}

func readTemplate(fileName string, templateName string) (*template.Template) {
	funcs := template.FuncMap{"join": strings.Join, "isStandardType": isStandardType}
 	m, err := ioutil.ReadFile(fileName)
	check(err)
	tmpl, err := template.New(templateName).Funcs(funcs).Parse(string(m))
 	check(err)
	return tmpl
}

func readTemplates() (*template.Template) {
	templates := [4]string{"index.html", "item.html", "examples.html", "resource.html"}
	master := template.New("master")
	for _, tmpl := range templates {
		tmplName := strings.Replace(tmpl, ".html", "", -1)
		t := readTemplate(tmpl, tmplName)
		master.AddParseTree(tmplName, t.Tree)
	}
	return master
}

func main() {
	t := readTemplates()
	err := t.Execute(os.Stdout, map[string]interface{}{
		"title": "banán",
		"version": "1.0.0",
		"baseUri": "www.example.com",
		"baseUriParameters": map[string]interface{}{
			"zone": map[string]interface{}{
				"enum": []string{"us-east", "us-west", "emea", "apac"}}}})
	check(err)
}

