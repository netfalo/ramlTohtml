package main

import "bufio"
import "text/template"
import "regexp"
import "strings"
import "os"
import "flag"
import "github.com/Jumpscale/go-raml/raml"
import "gopkg.in/russross/blackfriday.v2"
import "github.com/GeertJohan/go.rice"
import "log"

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func isStandardType() bool {
	return true
}

func readTemplate(fileName string, templateName string) *template.Template {
	funcs := template.FuncMap{"join": strings.Join, "isStandardType": isStandardType}
	templateBox, err := rice.FindBox("templates")
	check(err)
	m, err := templateBox.String(fileName)
	check(err)
	tmpl, err := template.New(templateName).Funcs(funcs).Parse(string(m))
	check(err)
	return tmpl
}

func readTemplates() *template.Template {
	master := readTemplate("index.html", "index")
	templates := [3]string{"examples.html", "item.html", "resource.html"}
	for _, tmpl := range templates {
		tmplName := strings.Replace(tmpl, ".html", "", -1)
		t := readTemplate(tmpl, tmplName)
		master.AddParseTree(tmplName, t.Tree)
	}
	return master
}

func populateTemplate(data map[string]interface{}, outputFile string) {
	f, err := os.Create(outputFile)
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)

	t := readTemplates()
	err = t.Execute(w, data)
	check(err)
}

func makeUniqueId(input string) string {
	re, err := regexp.Compile("\\W")
	check(err)
	stringWithSpacesReplaced := re.ReplaceAllLiteralString(input, "_")
	stringWithLeadingUnderscoreRemoved := strings.TrimLeft(stringWithSpacesReplaced, "_")
	return strings.ToLower(stringWithLeadingUnderscoreRemoved)
}

func convertToMarkdown(description string) string {
	return string(blackfriday.Run([]byte(description)))
}

func convertNamedParameters(namedParameters map[string]raml.NamedParameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for index := range namedParameters {
		param := make(map[string]interface{})
		np := namedParameters[index]
		param["key"] = index
		param["type"] = np.Type
		param["description"] = convertToMarkdown(np.Description)
		param["example"] = np.Example
		param["minLength"] = np.MinLength
		param["maxLength"] = np.MaxLength
		param["minimum"] = np.Minimum
		param["maximum"] = np.Maximum
		result = append(result, param)
	}
	return result
}

func convertSecuredBys(definitionOfChoices []raml.DefinitionChoice) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for index := range definitionOfChoices {
		res := make(map[string]interface{})
		res["schemeName"] = definitionOfChoices[index].Name
		result = append(result, res)
	}
	return result
}

func convertResponses(responses map[raml.HTTPCode]raml.Response) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for index := range responses {
		res := make(map[string]interface{})
		res["code"] = index
		res["body"] = convertBodies(responses[index].Bodies)
		result = append(result, res)
	}
	return result
}

func convertBodies(bodies raml.Bodies) map[string]interface{} {
	result := make(map[string]interface{})
	result["description"] = convertToMarkdown(bodies.Description)
	return result
}

func convertMethod(method *raml.Method, uniqueId string, parentUrl string, relativeUri string, uriParameters []map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["method"] = strings.ToLower((*method).Name)
	res["displayName"] = (*method).DisplayName
	res["description"] = convertToMarkdown((*method).Description)
	res["queryParameters"] = convertNamedParameters((*method).QueryParameters)
	res["securedBy"] = convertSecuredBys((*method).SecuredBy)
	res["annotations"] = ""
	res["headers"] = ""
	res["queryString"] = convertNamedParameters((*method).QueryString)
	res["responses"] = convertResponses((*method).Responses)
	res["body"] = ""

	res["allUriParameters"] = uriParameters
	res["uniqueId"] = uniqueId
	res["parentUrl"] = parentUrl
	res["relativeUri"] = relativeUri
	return res
}

func convertMethods(methods []*raml.Method, uniqueId string, parentUrl string, relativeUri string, uriParameters []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for index := range methods {
		res := convertMethod(methods[index], uniqueId, parentUrl, relativeUri, uriParameters)
		result = append(result, res)
	}
	return result
}

func convertResource(r raml.Resource) map[string]interface{} {
	res := make(map[string]interface{})
	res["relativeUri"] = r.URI
	res["displayName"] = r.DisplayName
	res["description"] = convertToMarkdown(r.Description)
	res["resources"] = convertResourcesPtr(r.Nested)
	if r.Parent == nil {
		res["parentUrl"] = ""
	} else {
		res["parentUrl"] = r.Parent.URI
	}
	uniqueId := makeUniqueId(res["parentUrl"].(string) + res["relativeUri"].(string))
	res["uniqueId"] = uniqueId
	uriParameters := convertNamedParameters(r.URIParameters)
	res["methods"] = convertMethods(r.Methods, uniqueId, res["parentUrl"].(string), res["relativeUri"].(string), uriParameters)
	return res
}

func convertResourcesPtr(resources map[string]*raml.Resource) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for resource := range resources {
		r := *resources[resource]
		res := convertResource(r)
		result = append(result, res)
	}
	return result
}

func convertResources(resources map[string]raml.Resource) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for resource := range resources {
		r := resources[resource]
		res := convertResource(r)
		result = append(result, res)
	}
	return result
}

func convertDocumentation(documentation []raml.Documentation) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 0)
	for index := range documentation {
		res := make(map[string]interface{})
		res["title"] = documentation[index].Title
		res["content"] = convertToMarkdown(documentation[index].Content)
		res["uniqueId"] = makeUniqueId(documentation[index].Title)
		result = append(result, res)
	}
	return result
}

func convertRamlToMap(root *raml.APIDefinition) map[string]interface{} {
	return map[string]interface{}{
		"title":             root.Title,
		"version":           root.Version,
		"baseUri":           root.BaseURI,
		"baseUriParameters": convertNamedParameters(root.BaseURIParameters),
		"documentation":     convertDocumentation(root.Documentation),
		"resources":         convertResources(root.Resources)}
}

func main() {
	inputFile := flag.String("in", "api.raml", "path")
	outputFile := flag.String("out", "result.html", "path")
	flag.Parse()
	root := new(raml.APIDefinition)
	err := raml.ParseFile(*inputFile, root)
	check(err)
	data := convertRamlToMap(root)
	populateTemplate(data, *outputFile)
}
