task :default => :build

desc "Compile the code"
task :build do
  sh %!go mod vendor!
  sh %!go mod tidy!
  sh %!GO111MODULE=on GOOS=linux go build -ldflags="-s -w" -mod=vendor -o handler ./...!
end

desc "Deploy on lambda"
task :deploy => :build do
  rm_f "handler.zip"
  sh %!zip handler.zip handler!
  sh %!aws lambda --profile ip-saas-dev --region eu-west-1 \
    update-function-code \
    --function-name "codepipeline-github-status" \
    --zip-file "fileb://handler.zip" \
  !
end
