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
      {{- if and .UseGPU (ne .GPUName "") }} 
      nodeSelector:
        accelerator: {{ .GPUName }}
      {{- end }}  
      containers:
        {{- $jobContainerName := (.JobContainerName) }}
        {{- $jobContainerImage := (.JobContainerImage) }}
        {{- $execCmd := (.ExecCmd) }}
        {{- $argsStr := (.ArgsStr) }}
        {{- $cpu := (.CPU) }}
        {{- $memory := (.Memory) }}
        {{- $useGPU := (.UseGPU) }}
        {{- $gpuCount := (.GPUCount) }}
		{{- range $val := Iterate .ContainersPerPod }}
      - name: {{ $jobContainerName }}-{{$val}}   
        image: {{ $jobContainerImage }}
        command: ["sh", "-c", "{{ $execCmd }} {{ $argsStr }}"]
        {{- end }}
        resources:
          requests:
            memory: "{{ $memory }}"
          limits:
            memory: "{{ $memory }}"
            cpu: "{{ $cpu }}"
            {{- if $useGPU }}
              nvidia.com/gpu: "{{ $gpuCount }}"
            {{- end }}
      restartPolicy: Never
  backoffLimit: 4
`
