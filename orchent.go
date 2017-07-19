package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/dghubble/sling"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const OrchentVersion string = "1.0.4"

var (
	app     = kingpin.New("orchent", "The orchestrator client. Please store your access token in the 'ORCHENT_TOKEN' environment variable: 'export ORCHENT_TOKEN=<your access token>'. If you need to specify the file containing the trusted root CAs use the 'ORCHENT_CAFILE' environment variable: 'export ORCHENT_CAFILE=<path to file containing trusted CAs>'.").Version(OrchentVersion)
	hostUrl = app.Flag("url", "the base url of the orchestrator rest interface. Alternative the environment variable 'ORCHENT_URL' can be used: 'export ORCHENT_URL=<the_url>'").Short('u').String()

	lsDep       = app.Command("depls", "list deployments")
	lsDepFilter = lsDep.Flag("created_by", "the subject@issuer of user to filter the deployments for, 'me' is shorthand for the current user").Short('c').String()

	showDep     = app.Command("depshow", "show a specific deployment")
	showDepUuid = showDep.Arg("uuid", "the uuid of the deployment to display").Required().String()

	createDep          = app.Command("depcreate", "create a new deployment")
	createDepCallback  = createDep.Flag("callback", "the callback url").Default("").String()
	createDepTemplate  = createDep.Arg("template", "the tosca template file").Required().File()
	createDepParameter = createDep.Arg("parameter", "the parameter to set (json object)").Required().String()

	updateDep          = app.Command("depupdate", "update the given deployment")
	updateDepCallback  = updateDep.Flag("callback", "the callback url").Default("").String()
	updateDepUuid      = updateDep.Arg("uuid", "the uuid of the deployment to update").Required().String()
	updateDepTemplate  = updateDep.Arg("template", "the tosca template file").Required().File()
	updateDepParameter = updateDep.Arg("parameter", "the parameter to set (json object)").Required().String()

	depTemplate     = app.Command("deptemplate", "show the template of the given deployment")
	templateDepUuid = depTemplate.Arg("uuid", "the uuid of the deployment to get the template").Required().String()

	delDep     = app.Command("depdel", "delete a given deployment")
	delDepUuid = delDep.Arg("uuid", "the uuid of the deployment to delete").Required().String()

	lsRes        = app.Command("resls", "list the resources of a given deployment")
	lsResDepUuid = lsRes.Arg("depployment uuid", "the uuid of the deployment").Required().String()

	showRes        = app.Command("resshow", "show a specific resource of a given deployment")
	showResDepUuid = showRes.Arg("deployment uuid", "the uuid of the deployment").Required().String()
	showResResUuid = showRes.Arg("resource uuid", "the uuid of the resource to show").Required().String()

	testUrl = app.Command("test", "test if the given url is pointing to an orchestrator, please use this to ensure there is no typo in the url.")
)

type OrchentError struct {
	Code     int    `json:"code"`
	Title1   string `json:"title"`
	Title2   string `json:"error"`
	Message1 string `json:"message"`
	Message2 string `json:"error_description"`
}

func (e OrchentError) Error() string {
	if e.Title1 != "" || e.Message1 != "" {
		return fmt.Sprintf("Error '%s' [%d]: %s", e.Title1, e.Code, e.Message1)
	} else if e.Title2 != "" || e.Message2 != "" {
		return fmt.Sprintf("Error '%s': %s", e.Title2, e.Message2)
	} else {
		return ""
	}
}

func is_error(e *OrchentError) bool {
	return e.Error() != ""
}

type OrchentInfo struct {
	Version   string `json:"projectVersion"`
	Hostname  string `json:"serverHostname"`
	Revision  string `json:"projectRevision"`
	Timestamp string `json:"projectTimestamp"`
}

type OrchentLink struct {
	Rel  string `json:"rel"`
	HRef string `json:"href"`
}

func get_link(key string, links []OrchentLink) *OrchentLink {
	for _, link := range links {
		if link.Rel == key {
			return &link
		}
	}
	return nil
}

type OrchentPage struct {
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Number        int `json:"number"`
}

type OrchentDeployment struct {
	Uuid              string                 `json:"uuid"`
	CreationTime      string                 `json:"creationTime"`
	UpdateTime        string                 `json:"updateTime"`
	Status            string                 `json:"status"`
	StatusReason      string                 `json:"statusReason"`
	Task              string                 `json:"task"`
	CloudProviderName string                 `json:"CloudProviderName"`
	Callback          string                 `json:"callback"`
	Outputs           map[string]interface{} `json:"outputs"`
	Links             []OrchentLink          `json:"links"`
}

type OrchentResource struct {
	Uuid          string        `json:"uuid"`
	CreationTime  string        `json:"creationTime"`
	State         string        `json:"state"`
	ToscaNodeType string        `json:"toscaNodeType"`
	ToscaNodeName string        `json:"toscaNodeName"`
	RequiredBy    []string      `json:"requiredBy"`
	Links         []OrchentLink `json:"links"`
}

type OrchentDeploymentList struct {
	Deployments []OrchentDeployment `json:"content"`
	Links       []OrchentLink       `json:"links"`
	Page        OrchentPage         `json:"page"`
}

type OrchentResourceList struct {
	Resources []OrchentResource `json:"content"`
	Links     []OrchentLink     `json:"links"`
	Page      OrchentPage       `json:"page"`
}

type OrchentCreateRequest struct {
	Template   string                 `json:"template"`
	Parameters map[string]interface{} `json:"parameters"`
	Callback   string                 `json:"callback,omitempty"`
}

func (depList OrchentDeploymentList) String() string {
	output := ""
	output = output + fmt.Sprintf("  page: %s\n", depList.Page)
	output = output + fmt.Sprintln("  links:")
	for _, link := range depList.Links {
		output = output + fmt.Sprintf("    %s\n", link)
	}
	output = output + fmt.Sprintln("\n")
	for _, dep := range depList.Deployments {
		output = output + deployment_to_string(dep, true)
	}
	return output
}

func (dep OrchentDeployment) String() string {
	output := deployment_to_string(dep, false)
	return output
}

func deployment_to_string(dep OrchentDeployment, short bool) string {
	output := ""
	lines := []string{"Deployment [" + dep.Uuid + "]:",
		"  status: " + dep.Status,
		"  creation time: " + dep.CreationTime,
		"  update time: " + dep.UpdateTime,
		"  callback: " + dep.Callback,
	}
	if !short {
		outputs, _ := json.MarshalIndent(dep.Outputs, "  ", "    ")
		more_lines := []string{
			"  status reason: " + dep.StatusReason,
			"  task: " + dep.Task,
			"  CloudProviderName: " + dep.CloudProviderName,
			"  outputs: \n  " + fmt.Sprintf("%s", outputs),
			"  links:"}
		lines = append(lines, more_lines...)
	}
	for _, line := range lines {
		output = output + fmt.Sprintf("%s\n", line)
	}
	if !short {
		for _, link := range dep.Links {
			output = output + fmt.Sprintf("    %s\n", link)
		}
	}
	return output

}

func (resList OrchentResourceList) String() string {
	output := ""
	output = output + fmt.Sprintf("  page: %s\n", resList.Page)
	output = output + fmt.Sprintln("  links:")
	for _, link := range resList.Links {
		output = output + fmt.Sprintf("    %s\n", link)
	}
	for _, res := range resList.Resources {
		output = output + fmt.Sprintln(res)
	}
	return output
}

func (res OrchentResource) String() string {
	lines := []string{"Resource [" + res.Uuid + "]:",
		"  creation time: " + res.CreationTime,
		"  state: " + res.State,
		"  toscaNodeType: " + res.ToscaNodeType,
		"  toscaNodeName: " + res.ToscaNodeName,
		"  requiredBy:"}
	output := ""
	for _, line := range lines {
		output = output + fmt.Sprintf("%s\n", line)
	}
	for _, req := range res.RequiredBy {
		output = output + fmt.Sprintf("    %s\n", req)
	}
	output = output + "  links:\n"
	for _, link := range res.Links {
		output = output + fmt.Sprintf("    %s\n", link)
	}
	return output
}

func (link OrchentLink) String() string {
	return fmt.Sprintf("%s [%s]", link.Rel, link.HRef)
}

func (page OrchentPage) String() string {
	return fmt.Sprintf("%d/%d [ #Elements: %d, size: %d ]", page.Number, page.TotalPages, page.TotalElements, page.Size)
}

func client() *http.Client {
	ca_file, use_other_ca := os.LookupEnv("ORCHENT_CAFILE")

	if use_other_ca {
		rootCAs := x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(read_ca_file(ca_file))
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: rootCAs},
		}
		return &http.Client{Transport: tr}
	}
	return http.DefaultClient
}

func read_ca_file(caFileName string) []byte {
	data := make([]byte, 0)
	caFile, openErr := os.Open(caFileName)
	if openErr != nil {
		fmt.Printf("error opening ca-file: %s\n", openErr)
		return data[:0]
	}
	info, infoErr := caFile.Stat()
	if infoErr != nil {
		fmt.Printf("error getting ca-file size: %s\n", infoErr)
		return data[:0]
	}
	size := info.Size()
	data = make([]byte, size)
	count, readErr := caFile.Read(data)
	if readErr != nil || int64(count) < size {
		fmt.Printf("error reading the ca-file: %s\n  (read %d/%d)\n", readErr, count, size)
		return data[:0]
	}
	return data[:count]
}

func deployments_list(base *sling.Sling, filter string) {
	append := "./deployments"
	if filter != "" {
		append += ("?createdBy=" + filter)
	}
	base = base.Get(append)
	fmt.Println("retrieving deployment list:")
	receive_and_print_deploymentlist(base)
}

func receive_and_print_deploymentlist(complete *sling.Sling) {
	deploymentList := new(OrchentDeploymentList)
	orchentError := new(OrchentError)
	_, err := complete.Receive(deploymentList, orchentError)
	if err != nil {
		fmt.Printf("error requesting list of providers:\n %s\n", err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error requesting list of deployments:\n %s\n", orchentError)
	} else {
		links := deploymentList.Links
		curPage := get_link("self", links)
		nextPage := get_link("next", links)
		lastPage := get_link("last", links)
		fmt.Printf("%s\n", deploymentList)
		if curPage != nil && nextPage != nil && lastPage != nil &&
			curPage.HRef != lastPage.HRef {
			receive_and_print_deploymentlist(base_connection(nextPage.HRef))
		}

	}
}

func deployment_create_update(templateFile *os.File, parameter string, callback string, depUuid *string, base *sling.Sling) {
	var parameterMap map[string]interface{}
	paramErr := json.Unmarshal([]byte(parameter), &parameterMap)
	if paramErr != nil {
		fmt.Printf("error parsing the parameter: %s\n", paramErr)
		return
	}

	info, infoErr := templateFile.Stat()
	if infoErr != nil {
		fmt.Printf("error getting file size: %s\n", infoErr)
		return
	}
	size := info.Size()
	data := make([]byte, size)
	count, readErr := templateFile.Read(data)
	if readErr != nil || int64(count) < size {
		fmt.Printf("error reading the file: %s\n  (read %d/%d)\n", readErr, count, size)
		return
	}
	template := string(data[:count])
	body := &OrchentCreateRequest{
		Template:   template,
		Parameters: parameterMap,
		Callback:   callback,
	}
	deployment := new(OrchentDeployment)
	orchentError := new(OrchentError)
	action := ""
	if depUuid == nil {
		action = "creating"
		base = base.Post("./deployments")
	} else {
		action = "updating"
		base = base.Put("./deployments/" + *depUuid)
	}
	_, err := base.BodyJSON(body).Receive(deployment, orchentError)
	if err != nil {
		fmt.Printf("error %s deployment:\n %s\n", action, err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error %s deployment:\n %s\n", action, orchentError)
	} else {
		if depUuid == nil {
			fmt.Printf("%s\n", deployment)
		} else {
			fmt.Println("update of deployment %s successfully triggered\n", depUuid)
		}
	}
}

func deployment_show(uuid string, base *sling.Sling) {
	deployment := new(OrchentDeployment)
	orchentError := new(OrchentError)
	base = base.Get("./deployments/" + uuid)
	_, err := base.Receive(deployment, orchentError)
	if err != nil {
		fmt.Printf("error requesting provider %s:\n %s\n", uuid, err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error requesting deployment %s:\n %s\n", uuid, orchentError)
	} else {
		fmt.Printf("%s\n", deployment)
	}
}

func deployment_get_template(uuid string, base *sling.Sling) {
	orchentError := new(OrchentError)
	req, err := base.Get("./deployments/" + uuid + "/template").Request()
	if err != nil {
		fmt.Printf("error requesting template of %s:\n  %s\n", uuid, err)
		return
	}
	// unable to use sling here as the return is plain text and not json
	cl := client()
	resp, err := cl.Do(req)
	if err != nil {
		fmt.Printf("error requesting template of %s:\n  %s\n", uuid, err)
		return
	}
	defer resp.Body.Close()
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			fmt.Print(scanner.Text())
		}
	} else {
		json.NewDecoder(resp.Body).Decode(orchentError)
		fmt.Printf("error requesting template of %s:\n  %s\n", uuid, orchentError)
	}
}

func deployment_delete(uuid string, base *sling.Sling) {
	orchentError := new(OrchentError)
	_, err := base.Delete("./deployments/"+uuid).Receive(nil, orchentError)
	if err != nil {
		fmt.Printf("error deleting deployment %s:\n  %s\n", uuid, err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error deleting deployment %s:\n %s\n", uuid, orchentError)
	} else {
		fmt.Printf("deletion of deployment %s successfully triggered\n", uuid)
	}
}

func resources_list(depUuid string, base *sling.Sling) {
	base = base.Get("./deployments/" + depUuid + "/resources")
	fmt.Println("retrieving resource list:")
	receive_and_print_resourcelist(depUuid, base)
}

func receive_and_print_resourcelist(depUuid string, complete *sling.Sling) {
	resourceList := new(OrchentResourceList)
	orchentError := new(OrchentError)
	_, err := complete.Receive(resourceList, orchentError)
	if err != nil {
		fmt.Printf("error requesting list of resources for %s:\n %s\n", depUuid, err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error requesting resource list for %s:\n %s\n", depUuid, orchentError)
	} else {
		links := resourceList.Links
		curPage := get_link("self", links)
		nextPage := get_link("next", links)
		lastPage := get_link("last", links)
		fmt.Printf("%s\n", resourceList)
		if curPage != nil && nextPage != nil && lastPage != nil &&
			curPage.HRef != lastPage.HRef {
			receive_and_print_resourcelist(depUuid, base_connection(nextPage.HRef))
		}
	}
}

func resource_show(depUuid string, resUuid string, base *sling.Sling) {
	resource := new(OrchentResource)
	orchentError := new(OrchentError)
	base = base.Get("./deployments/" + depUuid + "/resources/" + resUuid)
	_, err := base.Receive(resource, orchentError)
	if err != nil {
		fmt.Printf("error requesting resources %s for %s:\n %s\n", resUuid, depUuid, err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error requesting resource %s for %s:\n %s\n", resUuid, depUuid, orchentError)
	} else {
		fmt.Printf("%s\n", resource)
	}
}

func test_url(base *sling.Sling) {
	info := new(OrchentInfo)
	orchentError := new(OrchentError)
	base = base.Get("./info")
	_, err := base.Receive(info, orchentError)
	if err != nil {
		fmt.Println("error checking orchent url, it seems like the url is not correct")
		return
	}
	if is_error(orchentError) {
		fmt.Println("error checking orchent url, it seems like the url is not correct")
	} else {
		fmt.Println("looks like the orchent url is valid")
	}
}

func base_connection(urlBase string) *sling.Sling {
	client := client()
	tokenValue, tokenSet := os.LookupEnv("ORCHENT_TOKEN")
	base := sling.New().Client(client).Base(urlBase)
	base = base.Set("User-Agent", "Orchent")
	base = base.Set("Accept", "application/json")
	if tokenSet {
		token := "Bearer " + tokenValue
		return base.Set("Authorization", token)
	} else {
		fmt.Println(" ")
		fmt.Println("*** WARNING: no access token has been specified ***")
		return base
	}
}

func base_url(rawUrl string) string {
	if !strings.HasSuffix(rawUrl, "/") {
		rawUrl = rawUrl + "/"
	}
	u, _ := url.Parse(rawUrl)
	urlBase := u.Scheme + "://" + u.Host + u.Path
	return urlBase
}

func get_base_url() string {
	urlValue, urlSet := os.LookupEnv("ORCHENT_URL")
	baseUrl := ""
	if *hostUrl != "" {
		baseUrl = base_url(*hostUrl)
	} else if urlSet {
		baseUrl = base_url(urlValue)
	} else {
		fmt.Println("*** ERROR: No url given! Either set the environment varible 'ORCHENT_URL' or use the --url flag")
		os.Exit(1)
	}
	return baseUrl
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case lsDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployments_list(base, *lsDepFilter)

	case showDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployment_show(*showDepUuid, base)

	case createDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployment_create_update(*createDepTemplate, *createDepParameter, *createDepCallback, nil, base)

	case updateDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployment_create_update(*updateDepTemplate, *updateDepParameter, *updateDepCallback, updateDepUuid, base)

	case depTemplate.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployment_get_template(*templateDepUuid, base)

	case delDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployment_delete(*delDepUuid, base)

	case lsRes.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		resources_list(*lsResDepUuid, base)

	case showRes.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		resource_show(*showResDepUuid, *showResResUuid, base)

	case testUrl.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		test_url(base)
	}
}
