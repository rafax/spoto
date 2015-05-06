GOOS=linux GOARCH=amd64 go build
tar -cjvf spoto.tar.bz2 spoto
scp -i /Users/rafal/dev/aws/rafal.pem -P 2225 /Users/rafal/dev/go/src/github.com/rafax/spoto/spoto.tar.bz2 ubuntu@52.1.82.203:/app/spoto
ssh -i /Users/rafal/dev/aws/rafal.pem ubuntu@52.1.82.203 -p 2225 'cd /app/spoto; spoto ; tar -xvf /app/spoto/spoto.tar.bz2 ; sudo supervisorctl restart spoto'