FROM armory-docker-local.jfrog.io/armory-cloud/go-app
COPY ./build/dist/linux_amd64/armory /opt/go-application/goapp
