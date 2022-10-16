version_settings(constraint='>=0.30.9')

load('ext://restart_process', 'docker_build_with_restart')
load('ext://secret', 'secret_from_dict')
load('ext://configmap', 'configmap_from_dict')
load('ext://configmap', 'configmap_create')

ports = {
    'picsum-photos': (8080, 2345),
    'image-service': (8081, 2346),
    'minio': (9000, 9001),
}

# picsum-photos

docker_build_with_restart(
    'dmarby/picsum-photos',
    context='.',
    dockerfile='./containers/picsum-photos/Dockerfile.dev',
    entrypoint='''
        dlv debug \\
        --accept-multiclient \\
        --continue \\
        --headless \\
        --listen=:%s \\
        --api-version=2 \\
        --log \\
        ./cmd/picsum-photos
    ''' % ports['picsum-photos'][1],
    live_update=[
        fall_back_on('./go.mod'),
        fall_back_on([
            './web',
            './package.json',
            './package-lock.json',
        ]),
        sync('.', '/app/'),
    ],
)

k8s_yaml(secret_from_dict('picsum-hmac', inputs = {
    'hmac_key': 'foo',
}))
k8s_resource(
    new_name='picsum-hmac',
    objects=['picsum-hmac:secret'],
    labels=['picsum-photos', 'image-service'],
)

k8s_yaml([
    'kubernetes/picsum.yaml',
])
# Create images manifest
k8s_yaml(configmap_from_dict('picsum-images', inputs={
  'picsum-images.json': read_file('./test/fixtures/file/metadata.json'),
}))
k8s_resource(
    new_name='picsum-images',
    objects=['picsum-images:configmap'],
    labels=['picsum-photos'],
)
k8s_resource(
    'picsum',
    port_forwards=[
        port_forward(ports['picsum-photos'][0], ports['picsum-photos'][0], name='picsum'),
        port_forward(ports['picsum-photos'][1], ports['picsum-photos'][1], name='debugger'),
    ],
    labels=['picsum-photos'],
)


# image-service

# Minio to simulate DigitalOcean spaces

minio_root_username = 'username'
minio_root_password = 'password'

minio_access_key = 'access_key'
minio_secret_key = 'secret_key'

# Load file into configmap for minio-setup to upload
configmap_create('minio-files', from_file='1.jpg=./test/fixtures/file/1.jpg')
k8s_resource(
    new_name='minio-files',
    objects=['minio-files:configmap'],
    labels=['minio'],
)


minio = '''
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  labels:
    component: minio
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      component: minio
  template:
    metadata:
      labels:
        component: minio
    spec:
      volumes:
      - name: storage
        emptyDir: {}
      - name: config
        emptyDir: {}
      containers:
      - name: minio
        image: minio/minio:latest
        imagePullPolicy: IfNotPresent
        args:
        - server
        - /storage
        - --config-dir=/config
        env:
        - name: MINIO_CONSOLE_ADDRESS
          value: ":%s"
        - name: MINIO_ROOT_USER
          value: "%s"
        - name: MINIO_ROOT_PASSWORD
          value: "%s"
        ports:
        - containerPort: 9000
        volumeMounts:
        - name: storage
          mountPath: "/storage"
        - name: config
          mountPath: "/config"

---
apiVersion: v1
kind: Service
metadata:
  name: minio
  labels:
    component: minio
spec:
  clusterIP: None
  selector:
    component: minio
  ports:
  - port: 9000
    name: minio

---
apiVersion: batch/v1
kind: Job
metadata:
  name: minio-setup
  labels:
    component: minio
spec:
  template:
    metadata:
      name: minio-setup
    spec:
      restartPolicy: OnFailure
      volumes:
      - name: minio-files
        configMap:
          name: minio-files
      containers:
      - name: mc
        image: minio/mc:latest
        imagePullPolicy: IfNotPresent
        entrypoint: ["/bin/sh","-c"]
        volumeMounts:
          - name: minio-files
            mountPath: "/etc/minio-files"
            readOnly: true
        command:
        - /bin/sh
        - -c
        - "sleep 10 && mc alias set minio http://minio:%s %s %s && mc mb -p minio/picsum-photos && mc admin user add minio %s %s && mc admin policy set minio readwrite user=%s && mc cp /etc/minio-files/1.jpg minio/picsum-photos/1.jpg"
''' % (ports['minio'][1], minio_root_username, minio_root_password, ports['minio'][0], minio_root_username, minio_root_password, minio_access_key, minio_secret_key, minio_access_key)

k8s_yaml([
    blob(minio),
])
k8s_resource(
    'minio',
    port_forwards=[
        port_forward(ports['minio'][0], ports['minio'][0], name='minio-api'),
        port_forward(ports['minio'][1], ports['minio'][1], name='minio-console'),
    ],
    labels=['minio'],
)
k8s_resource(
  'minio-setup',
  labels=['minio'],
)

# image-service


k8s_yaml(secret_from_dict('picsum-spaces', inputs = {
    'access_key': minio_access_key,
    'secret_key': minio_secret_key,
    'endpoint': 'http://minio:%s' % ports['minio'][0],
    'space': 'picsum-photos',
}))
k8s_resource(
    new_name='picsum-spaces',
    objects=['picsum-spaces:secret'],
    labels=['image-service'],
)

docker_build_with_restart(
    'dmarby/image-service',
    context='.',
    dockerfile='./containers/image-service/Dockerfile.dev',
    entrypoint='''
        dlv debug \\
        --accept-multiclient \\
        --continue \\
        --headless \\
        --listen=:%s \\
        --api-version=2 \\
        --log \\
        ./cmd/image-service
    ''' % ports['image-service'][1],
    live_update=[
        fall_back_on('./go.mod'),
        sync('.', '/app/'),
    ],
)
k8s_yaml([
    'kubernetes/redis.yaml',
    'kubernetes/image-service.yaml',
])
k8s_resource(
    'image-service',
    port_forwards=[
        port_forward(ports['image-service'][0], ports['image-service'][0], name='image-service'),
        port_forward(ports['image-service'][1], ports['image-service'][1], name='debugger'),
    ],
    labels=['image-service'],
)
k8s_resource(
    'redis',
    port_forwards=[
        port_forward(6379, 6379, name='redis'),
    ],
    labels=['image-service'],
)
