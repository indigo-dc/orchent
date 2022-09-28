package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/dghubble/sling"
	"github.com/indigo-dc/liboidcagent-go/liboidcagent"
	"github.com/zpatrick/go-config"
	"gopkg.in/alecthomas/kingpin.v2"
)

const OrchentVersion string = "1.2.8"

var (
	app     = kingpin.New("orchent", "The orchestrator client. \n \nPlease either store your access token in 'ORCHENT_TOKEN' or set the account to use with oidc-agent in the 'ORCHENT_AGENT_ACCOUNT' and the socket of the oidc-agent in the 'OIDC_SOCK' environment variable: \n export ORCHENT_TOKEN=<your access token> \n         OR \n export OIDC_SOCK=<path to the oidc-agent socket> (usually this is already exported) \n export ORCHENT_AGENT_ACCOUNT=<account to use> \nIf you need to specify the file containing the trusted root CAs use the 'ORCHENT_CAFILE' environment variable: \n export ORCHENT_CAFILE=<path to file containing trusted CAs>\n \n").Version(OrchentVersion)
	hostUrl = app.Flag("url", "the base url of the orchestrator rest interface. Alternative the environment variable 'ORCHENT_URL' can be used: 'export ORCHENT_URL=<the_url>'").Short('u').String()

	lsDep       = app.Command("depls", "list deployments")
	lsDepUser   = lsDep.Flag("created_by", "the subject@issuer of user to filter the deployments for, 'me' is shorthand for the current user").Short('c').String()
	lsDepGroup   = lsDep.Flag("user_group", "the user group to filter the deployments for").Short('g').String()
	lsDepBefore = lsDep.Flag("before", "filter the deployments, they must be created before the given date/time, the format is YYYYMMDDHHMM").Short('b').String()
	lsDepAfter  = lsDep.Flag("after", "filter the deployments, they must be created after the given date/time, the format is YYYYMMDDHHMM").Short('a').String()

	showDep        = app.Command("depshow", "show a specific deployment")
	showDepUuid    = showDep.Arg("uuid", "the uuid of the deployment to display").Required().String()
	showDepVerbose = showDep.Flag("verbose", "enable verbose output").Default("false").Bool()
	showDepJson    = showDep.Flag("ojson", "json formatted output").Default("false").Bool()

	createDep                  = app.Command("depcreate", "create a new deployment")
	createDepJson              = createDep.Flag("ojson", "json formatted output").Default("false").Bool()
	createDepCallback          = createDep.Flag("callback", "the callback url").Default("").String()
	createDepMaxProvidersRetry = createDep.Flag("maxProvidersRetry", "Maximum number of cloud providers to be used in case of failure (Default: UNBOUNDED).").Uint8()
	createDepKeepLastAttempt   = createDep.Flag("keepLastAttempt", "In case of failure, keep the resources allocated in the last try (Default: true).").Default("true").Enum("true", "false")
	createDepUserGroup         = createDep.Flag("user_group", "the user group").Short('g').String()
	createDepTemplate          = createDep.Arg("template", "the tosca template file").Required().File()
	createDepParameter         = createDep.Arg("parameter", "the parameter to set (json object)").Required().String()

	updateDep                  = app.Command("depupdate", "update the given deployment")
	updateDepCallback          = updateDep.Flag("callback", "the callback url").Default("").String()
	updateDepMaxProvidersRetry = updateDep.Flag("maxProvidersRetry", "Maximum number of cloud providers to be used in case of failure (Default: UNBOUNDED).").Uint8()
	updateDepKeepLastAttempt   = updateDep.Flag("keepLastAttempt", "In case of failure, keep the resources allocated in the last try (Default: true).").Default("true").Enum("true", "false")
	updateDepUuid              = updateDep.Arg("uuid", "the uuid of the deployment to update").Required().String()
	updateDepTemplate          = updateDep.Arg("template", "the tosca template file").Required().File()
	updateDepParameter         = updateDep.Arg("parameter", "the parameter to set (json object)").Required().String()

	depTemplate     = app.Command("deptemplate", "show the template of the given deployment")
	templateDepUuid = depTemplate.Arg("uuid", "the uuid of the deployment to get the template").Required().String()

	delDep     = app.Command("depdel", "delete a given deployment")
	delDepUuid = delDep.Arg("uuid", "the uuid of the deployment to delete").Required().String()

	resetDep     = app.Command("depreset", "reset the state of a given deployment")
	resetDepUuid = resetDep.Arg("uuid", "the uuid of the deployment to reset").Required().String()
	resetDepStatus = resetDep.Flag("status", "the state of the deployment to be set forcefully (allowed values: DELETE_FAILED)").Default("DELETE_FAILED").Enum("DELETE_FAILED")

	logDep     = app.Command("deplog", "get the log for given deployment")
	logDepUuid = logDep.Arg("uuid", "the uuid of the deployment").Required().String()

	lsRes        = app.Command("resls", "list the resources of a given deployment")
	lsResDepUuid = lsRes.Arg("depployment uuid", "the uuid of the deployment").Required().String()

	showRes        = app.Command("resshow", "show a specific resource of a given deployment")
	showResDepUuid = showRes.Arg("deployment uuid", "the uuid of the deployment").Required().String()
	showResResUuid = showRes.Arg("resource uuid", "the uuid of the resource to show").Required().String()

	testUrl = app.Command("test", "test if the given url is pointing to an orchestrator, please use this to ensure there is no typo in the url.")
	getConfig = app.Command("showconf", "list the endpoints used by the current orchestrator.")
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

func deployment_time_to_number(time string) int {
	y := time[0:4]
	m := time[5:7]
	d := time[8:10]
	h := time[11:13]
	min := time[14:16]
	yi, _ := strconv.Atoi(y)
	mi, _ := strconv.Atoi(m)
	di, _ := strconv.Atoi(d)
	hi, _ := strconv.Atoi(h)
	mini, _ := strconv.Atoi(min)
	value := mini + 100*hi + 10000*di + 1000000*mi + 100000000*yi
	return value
}

type OrchentCreatedBy struct {
      Issuer    string    `json:"issuer"`
      Subject   string    `json:"subject"`
}

type OrchentDeployment struct {
	Uuid                   string                 `json:"uuid"`
	CreationTime           string                 `json:"creationTime"`
	UpdateTime             string                 `json:"updateTime"`
	CreatedBy              OrchentCreatedBy       `json:"createdBy"`
	UserGroup              string                 `json:"userGroup"`
	PhysicalId             string                 `json:"physicalId"`
	Status                 string                 `json:"status"`
	StatusReason           string                 `json:"statusReason"`
	Task                   string                 `json:"task"`
	CloudProviderName      string                 `json:"cloudProviderName"`
	CloudProviderEndpoint  map[string]interface{} `json:"cloudProviderEndpoint"`
	Callback               string                 `json:"callback"`
	Outputs                map[string]interface{} `json:"outputs"`
	Links                  []OrchentLink          `json:"links"`
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

func filter_deployments_by_time(depList OrchentDeploymentList, before int, after int) OrchentDeploymentList {
	newDeps := make([]OrchentDeployment, 0)
	for _, dep := range depList.Deployments {
		created := deployment_time_to_number(dep.CreationTime)
		isBefore := (created <= before) || (before < 0)
		isAfter := (created >= after) || (after < 0)
		if isBefore && isAfter {
			tempDeps := make([]OrchentDeployment, len(newDeps)+1)
			copy(tempDeps, newDeps)
			newDeps = tempDeps
			newDeps[len(newDeps)-1] = dep
		}
	}
	depList.Deployments = newDeps
	return depList
}

type OrchentResourceList struct {
	Resources []OrchentResource `json:"content"`
	Links     []OrchentLink     `json:"links"`
	Page      OrchentPage       `json:"page"`
}

type OrchentCreateRequest struct {
	Template          string                 `json:"template"`
	Parameters        map[string]interface{} `json:"parameters"`
	Callback          string                 `json:"callback,omitempty"`
	MaxProvidersRetry uint8                  `json:"maxProvidersRetry,omitempty"`
	KeepLastAttempt   string                 `json:"keepLastAttempt,omitempty"`
	UserGroup         string                 `json:"userGroup,omitempty"`
}

func (depList OrchentDeploymentList) String() string {
	output := ""
	output = output + fmt.Sprintf("  page: %s\n", depList.Page)
	output = output + fmt.Sprintln("  links:")
	for _, link := range depList.Links {
		output = output + fmt.Sprintf("    %s\n", link)
	}
	output = output + fmt.Sprintf("\n")
	for _, dep := range depList.Deployments {
		output = output + deployment_to_string(dep, 0)
	}
	return output
}

func (dep OrchentDeployment) String() string {
	output := deployment_to_string(dep, 1)
	return output
}

func (createdby OrchentCreatedBy) String() string {
        output := ""
	output = output + fmt.Sprintf("  { issuer: %s;", createdby.Issuer)
        output = output + fmt.Sprintf(" subject: %s }", createdby.Subject)
        return output
}


func deployment_to_string(dep OrchentDeployment, verboseLevel int) string {
	output := ""
	outputs, _ := json.MarshalIndent(dep.Outputs, "  ", "    ")
        lines := []string{"Deployment [" + dep.Uuid + "]:",
                "  status: " + dep.Status,
                "  creation time: " + dep.CreationTime,
                "  update time: " + dep.UpdateTime,
	}
	switch verboseLevel {
	case 0:
	case 1:
		lines = append(lines, []string { "  outputs: \n  " + fmt.Sprintf("%s", outputs) }...)
	case 2:
		endpoint, _ := json.MarshalIndent(dep.CloudProviderEndpoint, "  ", "    ")
		more_lines := []string{
			"  outputs: \n  " + fmt.Sprintf("%s", outputs),
			"  physical id: " + dep.PhysicalId,
			"  created by: " +  fmt.Sprintf("%s", dep.CreatedBy),
			"  user group: " + dep.UserGroup,
			"  status reason: " + dep.StatusReason,
			"  task: " + dep.Task,
			"  callback: " + dep.Callback,
			"  CloudProviderName: " + dep.CloudProviderName,
			"  CloudProviderEndpoint: " + fmt.Sprintf("%s", endpoint),
			"  links:"}
		lines = append(lines, more_lines...)
		for _, link := range dep.Links {
			lines = append(lines, []string { output + fmt.Sprintf("    %s\n", link) }...)
		}
	}
	for _, line := range lines {
		output = output + fmt.Sprintf("%s\n", line)
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

func time_string_to_int(time string) int {
	if len(time) == 0 {
		return -1
	}
	if len(time) != 12 {
		fmt.Println("the before/after parameter must be 12 digits long: YYYYMMDDHHMM")
		return 0
	}
	value, err := strconv.Atoi(time)
	if err != nil {
		fmt.Println("could not convert before/after to an integer")
		return 0
	}
	if value < 100000000000 {
		fmt.Println("the before/after value is too low")
		return 0
	}
	return value
}

func deployments_list(base *sling.Sling, user string, group string, before string, after string) {
	path := "./deployments"
	query_params := []string {}
	if user != "" {
		query_params = append(query_params, "createdBy=" + user)
	}
	if group != "" {
		query_params = append(query_params, "userGroup=" + group)
	}
	if len(query_params) > 0 {
	   path += "?" + strings.Join(query_params[:], "&")
	}
	base = base.Get(path)
	fmt.Println("retrieving deployment list:")
	before_int := time_string_to_int(before)
	after_int := time_string_to_int(after)
	if before_int == 0 || after_int == 0 {
		return
	}
	receive_and_print_deploymentlist(base, before_int, after_int)
}

func receive_and_print_deploymentlist(complete *sling.Sling, before int, after int) {
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
		filteredDeploymentList := filter_deployments_by_time(*deploymentList, before, after)
		if len(filteredDeploymentList.Deployments) > 0 {
			fmt.Printf("%s\n", filteredDeploymentList)
		}
		if curPage != nil && nextPage != nil && lastPage != nil &&
			curPage.HRef != lastPage.HRef {
			receive_and_print_deploymentlist(base_connection(nextPage.HRef), before, after)
		}

	}
}

func deployment_create_update(templateFile *os.File, parameter string, callback string, maxProvidersRetry uint8, keepLastAttempt string, depUuid *string, userGroup string, jsonFormat bool, base *sling.Sling) {

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
		Template:          template,
		Parameters:        parameterMap,
		Callback:          callback,
		MaxProvidersRetry: maxProvidersRetry,
		KeepLastAttempt:   keepLastAttempt,
		UserGroup:         userGroup,
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
		return
	}

	if depUuid == nil {
		if jsonFormat {
			data, _ := json.Marshal(deployment)
			fmt.Printf("%s\n", data)
		} else {
			fmt.Printf("%s\n", deployment)
		}
	} else {
		fmt.Printf("update of deployment %s successfully triggered\n", *depUuid)
	}

}

func get_deployment_extra_info(uuid string) {
	baseUrl := get_base_url()
	base := base_connection(baseUrl)
	orchentError := new(OrchentError)
	req, err := base.Get("./deployments/" + uuid + "/extrainfo").Request()
	if err != nil {
		fmt.Printf("error getting extra info of %s:\n  %s\n", uuid, err)
		return
	}
	// unable to use sling here as the return is plain text and not json
	cl := client()
	resp, err := cl.Do(req)
	if err != nil {
		fmt.Printf("error requesting extra info of %s:\n  %s\n", uuid, err)
		return
	}
	defer resp.Body.Close()
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		var bodyBytes []byte
		var err error
		
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
	    
	    if len(bodyBytes) > 0 {
			var prettyJSON bytes.Buffer
			if err = json.Indent(&prettyJSON, bodyBytes, "  ", "  "); err != nil {
				fmt.Printf("JSON parse error: %v", err)
				return
			}
			fmt.Println("  ====== Deployment extra information: ======\n  " + string(prettyJSON.Bytes()))
		
		}
	    
	} else {
		json.NewDecoder(resp.Body).Decode(orchentError)
		fmt.Printf("error processing extra info of %s:\n  %d\n", uuid, resp.StatusCode)
	}
}

func deployment_show(uuid string, verbose bool, jsonFormat bool, base *sling.Sling) {
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
		return
	}

	if jsonFormat {
		data, _ := json.Marshal(deployment)
		fmt.Printf("%s\n", data)
		return
	}
	if verbose {
		fmt.Printf("%s\n", deployment_to_string(*deployment, 2))
		get_deployment_extra_info(uuid)

	} else {
		fmt.Printf("%s\n", deployment_to_string(*deployment, 1))
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

type StatusReset struct {
    Status     string   `json:"status,omitempty"`
}	

func deployment_reset(uuid string, status string, base *sling.Sling) {
	orchentError := new(OrchentError)

	body := &StatusReset {
		Status: status,
	}
	base = base.BodyJSON(body).Patch("./deployments/" + uuid)
	_, err := base.Receive(nil, orchentError)
	if err != nil {
		fmt.Printf("error resetting state for deployment %s:\n %s\n", uuid, err)
		return
	}
	if is_error(orchentError) {
		fmt.Printf("error resetting state for deployment %s:\n %s\n", uuid, orchentError)
	} else {
		fmt.Printf("reset of deployment %s successfully triggered\n", uuid)
	}	

}	

func deployment_log(uuid string, base *sling.Sling) {
	orchentError := new(OrchentError)
	req, err := base.Get("./deployments/" + uuid + "/log").Request()
	if err != nil {
		fmt.Printf("error requesting log of %s:\n  %s\n", uuid, err)
		return
	}
	// unable to use sling here as the return is plain text and not json
	cl := client()
	resp, err := cl.Do(req)
	if err != nil {
		fmt.Printf("error requesting log of %s:\n  %s\n", uuid, err)
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
		fmt.Printf("error requesting log of %s:\n  %s\n", uuid, orchentError)
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

func get_conf(base *sling.Sling) {
	fmt.Println("retrieving orchestrator configuration:")
	config := make(map[string]string)
	orchentError := new(OrchentError)
	base = base.Get("./configuration")
	_, err := base.Receive(&config, orchentError)
	if err != nil {
		fmt.Printf("error requesting orchestrator configuration: %s\n", err)
                return
        }
        if is_error(orchentError) {
                fmt.Printf("error requesting orchestrator configuration: %s\n", orchentError)
        } else {
		for key, value := range config {
			fmt.Printf("    %s: %s\n", key, value)
		}
        }
}

func settings() map[string]string {
	emptyset := make(map[string]string)
	user, err := user.Current()
	if err != nil {
		return emptyset
	}
	dir := user.HomeDir + "/.config/orchent/orchent.conf"
	confFile := config.NewINIFile(dir)
	c := config.NewConfig([]config.Provider{confFile})
	settings, err := c.Settings()
	if err != nil {
		return emptyset
	}

	return settings
}

func aliases(settings map[string]string) map[string]string {
	aliases := make(map[string]string)
	for key, val := range settings {
		if key[0:6] == "alias." {
			alias := key[6:len(key)]
			aliases[alias] = val
			// aliases[val] = alias
		}
	}
	return aliases
}

func try_alias_uuid(alias string, aliases map[string]string) string {
	value, found := aliases[alias]
	if found {
		return value
	}
	return alias
}

func get_account() (issuerSet bool, agentIssuer string) {
	agentAccount, accountSet := os.LookupEnv("ORCHENT_AGENT_ACCOUNT")
	// issuerValue, issuerSet = os.LookupEnv("ORCHENT_ISSUER")
	// if !agentSet && issuerSet {
	// 	agentIssuer = issuerValue
	// }
	return accountSet, agentAccount
}

func user_info(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func try_agent_token(account string) (tokenSet bool, tokenValue string) {
	token, err := liboidcagent.GetAccessToken(account, 120, "", "wattson")
	if err != nil {
		fmt.Println("*** ERROR: Could not get token from oidc-agent and $ORCHENT_TOKEN not set ***")
		fmt.Printf("agent error: %s\n", err)
		return false, tokenValue
	}
	return true, token
}

func try_token(accountSet bool, account string) (tokenSet bool, token string) {
	tokenValue, tokenSet := os.LookupEnv("ORCHENT_TOKEN")
	if !tokenSet && accountSet {
		return try_agent_token(account)
	}
	return tokenSet, tokenValue
}

func base_connection(urlBase string) *sling.Sling {
	client := client()
	accountSet, account := get_account()
	tokenSet, tokenValue := try_token(accountSet, account)
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
	settings := settings()
	aliases := aliases(settings)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case lsDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployments_list(base, *lsDepUser, *lsDepGroup, *lsDepBefore, *lsDepAfter)

	case showDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*showDepUuid, aliases)
		deployment_show(uuid, *showDepVerbose, *showDepJson, base)

	case createDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		deployment_create_update(*createDepTemplate, *createDepParameter, *createDepCallback, *createDepMaxProvidersRetry, *createDepKeepLastAttempt, nil, *createDepUserGroup, *createDepJson, base)

	case updateDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*updateDepUuid, aliases)
		deployment_create_update(*updateDepTemplate, *updateDepParameter, *updateDepCallback, *updateDepMaxProvidersRetry, *updateDepKeepLastAttempt, &uuid, "", *createDepJson, base)

	case depTemplate.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*templateDepUuid, aliases)
		deployment_get_template(uuid, base)

	case delDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*delDepUuid, aliases)
		deployment_delete(uuid, base)

	case resetDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*resetDepUuid, aliases)
		deployment_reset(uuid, *resetDepStatus, base)	

	case logDep.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*logDepUuid, aliases)
		deployment_log(uuid, base)		

	case lsRes.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*lsResDepUuid, aliases)
		resources_list(uuid, base)

	case showRes.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		uuid := try_alias_uuid(*showResDepUuid, aliases)
		resUuid := try_alias_uuid(*showResResUuid, aliases)
		resource_show(uuid, resUuid, base)

	case testUrl.FullCommand():
		baseUrl := get_base_url()
		base := base_connection(baseUrl)
		test_url(base)

	case getConfig.FullCommand():
                baseUrl := get_base_url()
                base := base_connection(baseUrl)
                get_conf(base)
        }
}
