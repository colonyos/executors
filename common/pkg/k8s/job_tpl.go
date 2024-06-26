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
  ttlSecondsAfterFinished: 10 
  template:
    spec:
      activeDeadlineSeconds: {{ .Walltime }}
      {{- if and .UseGPU (ne .GPUName "") }} 
      nodeSelector:
        accelerator: {{ .GPUName }}
      {{- end }}  
      containers:
        {{- $envMap := (.EnvMap) }}
        {{- $processID := (.ProcessID) }}
        {{- $jobContainerName := (.JobContainerName) }}
        {{- $jobContainerImage := (.JobContainerImage) }}
        {{- $execCmd := (.ExecCmd) }}
        {{- $argsStr := (.ArgsStr) }}
        {{- $cpu := (.CPU) }}
        {{- $memory := (.Memory) }}
        {{- $useGPU := (.UseGPU) }}
        {{- $gpuCount := (.GPUCount) }}
        {{- $pvcName := (.PVCName) }}
        {{- $mountPath := (.MountPath) }}
		{{- range $val := Iterate .ContainersPerPod }}
      - name: {{ $jobContainerName }}-{{$val}}   
        image: {{ $jobContainerImage }}
        {{- if $mountPath }}
        volumeMounts:
          - name: kube-executor-volume
            mountPath: {{ $mountPath }}
        {{- end }}
        command: ["sh", "-c", "{{ $execCmd }} {{ $argsStr }}"]
        env:
        {{- range $key, $value := $envMap }} 
        - name: {{ $key }}
          value: "{{ $value }}"
        {{- end}}
        - name: COLONIES_PROCESS_ID
          value: "{{ $processID }}"
        resources:
          requests:
            memory: "{{ $memory }}"
          limits:
            memory: "{{ $memory }}"
            cpu: "{{ $cpu }}"
            {{- if $useGPU }}
            nvidia.com/gpu: "{{ $gpuCount }}"
            {{- end }}
        {{- end }}
      restartPolicy: Never
      {{- if $mountPath }}
      volumes:
        - name: kube-executor-volume 
          persistentVolumeClaim:
            claimName: {{ $pvcName }}
      {{- end }}
  backoffLimit: 4
`
