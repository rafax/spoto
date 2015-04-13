GOOS=linux GOARCH=amd64 go build
zip spoto.zip spoto
scp -i /Users/rafal/dev/aws/rafal.pem -P 2225 /Users/rafal/dev/go/src/github.com/rafax/spoto/spoto.zip ec2-user@52.1.82.203:/app/spoto
ssh -i /Users/rafal/dev/aws/rafal.pem ec2-user@52.1.82.203 -p 2225 'rm /app/spoto/spoto ; unzip /app/spoto/spoto.zip -d /app/spoto ; killall spoto ; SPOTO_DB_HOST=postgres.c0rawtfsbsk2.us-east-1.rds.amazonaws.com nohup /app/spoto/spoto > /dev/null 2>&1 &'