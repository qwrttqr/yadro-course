FROM golang:1.27

COPY . /tests

WORKDIR /tests

ENTRYPOINT [ "go", "test", "-race", "-v", "./..." ]
