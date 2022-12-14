/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	"bytes"
	"fmt"
	"net/url"
	"text/template"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
)

type Notification struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"                json:"id,omitempty"`
	CodehostID   int                 `bson:"codehost_id"                  json:"codehost_id"`
	Tasks        []*NotificationTask `bson:"tasks,omitempty"              json:"tasks,omitempty"`
	PrID         int                 `bson:"pr_id"                        json:"pr_id"`
	CommentID    string              `bson:"comment_id"                   json:"comment_id"`
	ProjectID    string              `bson:"project_id"                   json:"project_id"`
	Created      int64               `bson:"create"                       json:"create"`
	BaseURI      string              `bson:"base_uri"                     json:"base_uri"`
	IsPipeline   bool                `bson:"is_pipeline"                  json:"is_pipeline"`
	IsTest       bool                `bson:"is_test"                      json:"is_test"`
	IsScanning   bool                `bson:"is_scanning"                  json:"is_scanning"`
	IsWorkflowV4 bool                `bson:"is_workflowv4"                json:"is_workflowv4"`
	ErrInfo      string              `bson:"err_info"                     json:"err_info"`
	PrTask       *PrTaskInfo         `bson:"pr_task_info,omitempty"       json:"pr_task_info,omitempty"`
	Label        string              `bson:"label"                        json:"label"  `
	Revision     string              `bson:"revision"                     json:"revision"`
	RepoOwner    string              `bson:"repo_owner"                   json:"repo_owner"`
	RepoName     string              `bson:"repo_name"                    json:"repo_name"`
}

type PrTaskInfo struct {
	EnvStatus        string `bson:"env_status,omitempty"                json:"env_status,omitempty"`
	EnvName          string `bson:"env_name,omitempty"                  json:"env_name,omitempty"`
	EnvRecyclePolicy string `bson:"env_recycle_policy,omitempty"        json:"env_recycle_policy,omitempty"`
	ProductName      string `bson:"product_name,omitempty"              json:"product_name,omitempty"`
}

type NotificationTask struct {
	ProductName         string            `bson:"product_name"            json:"product_name"`
	WorkflowName        string            `bson:"workflow_name"           json:"workflow_name"`
	WorkflowDisplayName string            `bson:"workflow_display_name"   json:"workflow_display_name"`
	EncodedDisplayName  string            `bson:"encoded_display_name"    json:"encoded_display_name"`
	PipelineName        string            `bson:"pipeline_name"           json:"pipeline_name"`
	ScanningName        string            `bson:"scanning_name"           json:"scanningName"`
	ScanningID          string            `bson:"scanning_id"             json:"scanning_id"`
	TestName            string            `bson:"test_name"               json:"test_name"`
	ID                  int64             `bson:"id"                      json:"id"`
	Status              config.TaskStatus `bson:"status"                  json:"status"`
	TestReports         []*TestSuite      `bson:"test_reports,omitempty"  json:"test_reports,omitempty"`

	FirstCommented bool `json:"first_commented,omitempty" bson:"first_commented,omitempty"`
}

func (t NotificationTask) StatusVerbose() string {
	switch t.Status {
	case config.TaskStatusReady:
		return "?????????"
	case config.TaskStatusRunning:
		return "?????????"
	case config.TaskStatusCompleted:
		return "??????"
	case config.TaskStatusFailed:
		return "??????"
	case config.TaskStatusTimeout:
		return "??????"
	case config.TaskStatusCancelled:
		return "?????????"
	case config.TaskStatusPass:
		return "??????"
	default:
		return "??????"
	}
}

func (Notification) TableName() string {
	return "scm_notify"
}

func (n *Notification) ToString() string {
	return fmt.Sprintf("notification: %s #%d", n.ProjectID, n.PrID)
}

func (n *Notification) CreateCommentBody() (comment string, err error) {
	hasTest := false
	for _, task := range n.Tasks {
		task.EncodedDisplayName = url.QueryEscape(task.WorkflowDisplayName)
		if len(task.TestReports) != 0 {
			hasTest = true
			break
		}
	}

	tmplSource := ""
	if n.IsPipeline {
		if len(n.Tasks) == 0 {
			tmplSource = "??????????????????????????????????????????"
		} else if !hasTest {
			tmplSource =
				"|??????????????????|??????| \n |---|---| \n {{range .Tasks}}|[{{.PipelineName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/pipelines/single/{{.PipelineName}}/{{.ID}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | \n {{end}}"
		} else {
			tmplSource =
				"|??????????????????|??????|????????????????????????/??????????????????| \n |---|---|---| \n {{range .Tasks}}|[{{.PipelineName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/pipelines/single/{{.PipelineName}}/{{.ID}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | {{range .TestReports}}{{.Name}}: {{.Successes}}/{{.Tests}} <br> {{end}} | \n {{end}}"
		}
	} else if n.IsTest {
		if len(n.Tasks) == 0 {
			tmplSource = "???????????????????????????????????????"
		} else if !hasTest {
			tmplSource =
				"|???????????????|??????| \n |---|---| \n {{range .Tasks}}|[{{.TestName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/test/detail/function/{{.TestName}}/{{.ID}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | \n {{end}}"
		} else {
			tmplSource =
				"|???????????????|??????|????????????????????????/??????????????????| \n |---|---|---| \n {{range .Tasks}}|[{{.TestName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/test/detail/function/{{.TestName}}/{{.ID}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | {{range .TestReports}}{{.Name}}: {{.Successes}}/{{.Tests}} <br> {{end}} | \n {{end}}"
		}
	} else if n.IsScanning {
		if len(n.Tasks) == 0 {
			tmplSource = "?????????????????????????????????????????????"
		} else {
			tmplSource =
				"|?????????????????????|??????| \n |---|---| \n {{range .Tasks}}|[{{.ScanningName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/scanner/detail/{{.ScanningName}}/task/{{.ID}}?id={{.ScanningID}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | \n {{end}}"

		}
	} else if n.IsWorkflowV4 {
		if len(n.Tasks) == 0 {
			tmplSource = "??????????????????????????????????????????"
		} else {
			tmplSource =
				"|??????????????????|??????| \n |---|---| \n {{range .Tasks}}|[{{.WorkflowDisplayName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/pipelines/custom/{{.WorkflowName}}/{{.ID}}?display_name={{.EncodedDisplayName}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | \n {{end}}"
		}
	} else {
		if len(n.Tasks) == 0 {
			tmplSource = "??????????????????????????????????????????"
		} else if !hasTest {
			tmplSource =
				"|??????????????????|??????| \n |---|---| \n {{range .Tasks}}|[{{.WorkflowDisplayName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/pipelines/multi/{{.WorkflowName}}/{{.ID}}?display_name={{.EncodedDisplayName}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | \n {{end}}"
		} else {
			tmplSource =
				"|??????????????????|??????|????????????????????????/??????????????????| \n |---|---|---| \n {{range .Tasks}}|[{{.WorkflowDisplayName}}#{{.ID}}]({{$.BaseURI}}/v1/projects/detail/{{.ProductName}}/pipelines/multi/{{.WorkflowName}}/{{.ID}}?display_name={{.EncodedDisplayName}}) | {{if eq .StatusVerbose $.Success}} {+ {{.StatusVerbose}} +}{{else}}{- {{.StatusVerbose}} -}{{end}} | {{range .TestReports}}{{.Name}}: {{.Successes}}/{{.Tests}} <br> {{end}} | \n {{end}}"
		}
	}

	if n.PrTask != nil {
		if n.PrTask.EnvName != "" {
			content := fmt.Sprintf("?????????????????????[%s]({{$.BaseURI}}/v1/projects/detail/%s/envs/detail?envName=%s) ?????????%s \n\n", n.PrTask.EnvName, n.PrTask.ProductName, n.PrTask.EnvName, n.PrTask.EnvStatus)
			tmplSource = fmt.Sprintf("%s%s", content, tmplSource)
		}

		if n.PrTask.EnvRecyclePolicy != "" {
			policyName := getEnvRecyclePolicy(n.PrTask.EnvRecyclePolicy)
			content := fmt.Sprintf("???????????????????????????[%s]({{$.BaseURI}}/v1/projects/detail/%s/envs/detail?envName=%s) ???????????????%s \n\n", n.PrTask.EnvName, n.PrTask.ProductName, n.PrTask.EnvName, policyName)
			tmplSource = fmt.Sprintf("%s%s", tmplSource, content)
		}
	}

	tmpl := template.Must(template.New("comment").Parse(tmplSource))
	buffer := bytes.NewBufferString("")

	if err = tmpl.Execute(buffer, struct {
		Tasks   []*NotificationTask
		BaseURI string
		NoTask  bool
		Success string
	}{
		n.Tasks,
		n.BaseURI,
		len(n.Tasks) == 0,
		"??????",
	}); err != nil {
		return
	}

	return buffer.String(), nil
}

func getEnvRecyclePolicy(policy string) string {
	switch policy {
	case config.EnvRecyclePolicyAlways:
		return "????????????"
	case config.EnvRecyclePolicyTaskStatus:
		return "???????????????????????????"
	case config.EnvRecyclePolicyNever:
		return "????????????"
	default:
		return "????????????"
	}
}
