package k8s

const jobTemplate = `
apiVersion: batch/v1
kind: Job 
metadata:
  name: {{ .JobName }} 
  labels:
    app: kubeexecutor 
spec:
  completions: {{ .Parallelism }}
  parallelism: {{ .Parallelism }}
  template:
    spec:
      containers:
        {{- $jobContainerName := (.JobContainerName) }}
        {{- $jobContainerImage := (.JobContainerImage) }}
        {{- $execCmd := (.ExecCmd) }}
        {{- $argsStr := (.ArgsStr) }}
		{{- range $val := Iterate .ContainersPerPod }}
      - name: {{ $jobContainerName }}-{{$val}}   
        image: {{ $jobContainerImage }}
        command: ["sh", "-c", "{{ $execCmd }} {{ $argsStr }}"]
        {{- end }}
      restartPolicy: Never
  backoffLimit: 4
`
