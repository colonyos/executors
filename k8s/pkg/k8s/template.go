package k8s

const deploymentTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ (.DeploymentName) }} 
  labels:
    app: executor 
spec:
  replicas: {{ (.NumberOfPods) }} 
  selector:
    matchLabels:
      app: executor 
  template:
    metadata:
      labels:
        app: executor
    spec:
      containers:
        {{- $coloniesTLS := (.ColoniesTLS) }}
        {{- $coloniesServerHost := (.ColoniesServerHost) }}
        {{- $coloniesServerPort := (.ColoniesServerPort) }}
        {{- $coloniesColonyID := (.ColoniesColonyID) }}
        {{- $coloniesColonyPrvKey := (.ColoniesColonyPrvKey ) }}
        {{- $coloniesExecutorID := (.ColoniesExecutorID) }}
        {{- $coloniesExecutorPrvKey := (.ColoniesExecutorPrvKey ) }}
        {{- $test := (.TestMode) }}
        {{- $enableRamdisk := (.EnableRamdisk) }}
        {{- $ramdiskSize := (.RamdiskSize) }}
        {{- $dockerImage := (.DockerImage ) }}
		{{- range $val := Iterate .ExecutorsPerPod }}
      - name: executor-{{$val}}
        image: {{ $dockerImage }}
		{{- if $test }}
        command:
            - "sleep"
            - "1000"
	    {{- end}}
		{{- if $enableRamdisk }}
        volumeMounts:
        - mountPath: /ramdisk
          name: ramdisk
	    {{- end}}
        env:
        - name: COLONIES_SERVER_HOST
          value: "{{ $coloniesServerHost }}"
        - name: COLONIES_SERVER_PORT
          value: "{{ $coloniesServerPort }}"
        - name: COLONIES_COLONY_ID
          value: "{{ $coloniesColonyID }}"
        - name: COLONIES_COLONY_PRVKEY
          value: "{{ $coloniesColonyPrvKey }}"
        - name: COLONIES_EXECUTOR_ID
          value: "{{ $coloniesExecutorID }}"
        - name: COLONIES_EXECUTOR_PRVKEY
          value: "{{ $coloniesExecutorPrvKey }}"
        - name: COLONIES_TLS
          value: "{{ $coloniesTLS }}"
        - name: PODNAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
      {{- end }}
      imagePullSecrets:
      - name: prvdockerreg
	  {{- if $enableRamdisk }}
      volumes:
      - name: ramdisk
        emptyDir:
          medium: Memory
          sizeLimit: {{ $ramdiskSize }} 
      {{- end }}
`
