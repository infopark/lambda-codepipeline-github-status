package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

type event struct {
	ExecutionID string `json:"execution-id"`
	GithubToken string `json:"github-token"`
	Pipeline    string `json:"pipeline"`
}

type ghReqPayload struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// HandleLambdaEvent is triggered by a CloudWatch event rule.
func HandleLambdaEvent(ev event) error {
	if ev.ExecutionID == "" {
		return errors.New("missing event param execution-id")
	}
	if ev.GithubToken == "" {
		return errors.New("missing event param github-token")
	}
	if ev.Pipeline == "" {
		return errors.New("missing event param pipeline")
	}

	sess := session.Must(session.NewSession())
	cpSvc := codepipeline.New(sess)
	res, err := cpSvc.GetPipelineExecution(&codepipeline.GetPipelineExecutionInput{
		PipelineExecutionId: aws.String(ev.ExecutionID),
		PipelineName:        aws.String(ev.Pipeline),
	})
	if err != nil {
		return err
	}

	var sourceArti *codepipeline.ArtifactRevision
	for _, a := range res.PipelineExecution.ArtifactRevisions {
		if aws.StringValue(a.Name) == "SourceArtifact" {
			sourceArti = a
			break
		}
	}
	if sourceArti == nil {
		return errors.New("missing SourceArtifact")
	}

	rev := aws.StringValue(sourceArti.RevisionId)
	url, err := url.Parse(aws.StringValue(sourceArti.RevisionUrl))
	if err != nil {
		return err
	}
	status := aws.StringValue(res.PipelineExecution.Status)
	var ghStatus string
	switch status {
	case "InProgress":
		ghStatus = "pending"
	case "Succeeded":
		ghStatus = "success"
	default:
		ghStatus = "failure"
	}

	pathComponents := strings.Split(url.Path, "/")
	owner := pathComponents[1]
	repo := pathComponents[2]

	deepLink := fmt.Sprintf(
		"https://%s.console.aws.amazon.com/codesuite/codepipeline/pipelines/%s/executions/%s",
		"eu-west-1", ev.Pipeline, ev.ExecutionID)
	ghURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/statuses/%s", owner, repo, rev)

	log.Printf("Setting status for repo=%s/%s, commit=%s to %s\n", owner, repo, rev, ghStatus)

	var b bytes.Buffer
	err = json.NewEncoder(&b).Encode(ghReqPayload{
		State:     ghStatus,
		TargetURL: deepLink,
		Context:   "continuous-integration/codepipeline",
	})
	if err != nil {
		return err
	}

	ghReq, err := http.NewRequest("POST", ghURL, &b)
	if err != nil {
		return err
	}
	ghReq.Header.Set("Accept", "application/json")
	ghReq.Header.Set("Authorization", "token "+ev.GithubToken)
	ghReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	ghRes, err := client.Do(ghReq)
	if err != nil {
		return err
	}
	defer ghRes.Body.Close()
	if ghRes.StatusCode != 201 {
		resBody, _ := ioutil.ReadAll(ghRes.Body)
		return fmt.Errorf("unexpected response from GitHub: %d body: %s",
			ghRes.StatusCode, string(resBody))
	}

	return nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
