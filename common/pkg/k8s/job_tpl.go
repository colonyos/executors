package k8s

const jobTemplate = `
apiVersion: batch/v1
kind: Job 
metadata:
  name: {{ (.JobName) }} 
  labels:
    app: kubeexecutor 
spec:
  completions: 3
  parallelism: 3
  template:
    spec:
      containers:
      - name: {{ .JobContainerName}}   
        image: {{ .JobContainerImage }}
        command: ["sh", "-c", "echo Hello, World!"]
      restartPolicy: Never
  backoffLimit: 4
`
