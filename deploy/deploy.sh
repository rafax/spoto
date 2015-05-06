GOOS=linux GOARCH=amd64 go build
zip spoto.zip spoto
scp -i /Users/rafal/dev/aws/rafal.pem -P 2225 /Users/rafal/dev/go/src/github.com/rafax/spoto/spoto.zip ubuntu@52.1.82.203:/app/spoto
ssh -i /Users/rafal/dev/aws/rafal.pem ubuntu@52.1.82.203 -p 2225 'rm /app/spoto/spoto ; unzip /app/spoto/spoto.zip -d /app/spoto ; sudo supervisorctl restart spoto'