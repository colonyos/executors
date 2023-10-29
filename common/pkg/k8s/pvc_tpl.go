package k8s

const pvcTemplate = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ .PVCName }} 
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: {{ .StorageClass }}
  resources:
    requests:
      storage: {{ .DiskSize }}
`
