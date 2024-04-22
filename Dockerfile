FROM golang:1.22.2 as build

WORKDIR /go/workspace
COPY . .

ARG GIT_COMMIT
ENV GOPATH=/go/workspace
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN cd src/docker_rdma_sriov && go install -ldflags="-s -w -X main.GitCommitId=$GIT_COMMIT -extldflags "-static"" -v docker_rdma_sriov

FROM debian:bookworm-slim
RUN mkdir /container_tools
COPY --from=build /go/workspace/bin/docker_rdma_sriov /container_tools/docker_rdma_sriov

CMD cp -f /container_tools/* /tmp/
