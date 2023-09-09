package executor

const SifTemplate = `Bootstrap: docker
From: docker.io/{{.DockerImage}}

%post
  apt update
`
