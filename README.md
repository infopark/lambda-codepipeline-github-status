# lambda-codepipeline-github-status

This Lambda function updates the GitHub status of a pull request via CodePipeline events.

## Configuration

Add your GitHub token to the Lambda's env `GITHUB_TOKEN`.

Configure a CloudWatch event rule:

```json
{
  "source": [
    "aws.codepipeline"
  ],
  "detail-type": [
    "CodePipeline Stage Execution State Change"
  ],
  "detail": {
    "pipeline": [
      "my-pipeline"
    ]
  }
}
```

and connect it to the Lambda function. Set input to "Matched event".

Modify Lambda's policy to allow `codepipeline:GetPipelineExecution`.

## Testing

No tests yet

## Building

```rake build```

This command compiles a Linux binary `./handler`.

## Deploying

```rake deploy```

This command deploys the code to the already existing Go 1.x Lambda function
`codepipeline-github-status`. The handler's name is `handler`.
