FROM armory-docker-local.jfrog.io/armory-cloud/go-app
COPY ./build/dist/linux_amd64/armory /opt/go-application/goapp
# for 'traditional' users - make the look & feel of the container like it used to be - cli is named armory and is available via PATH
ENV PATH=$PATH:/home/goapp
RUN ln -s /opt/go-application/goapp /home/goapp/armory
