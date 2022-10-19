version_settings(constraint='>=0.30.9')

load('ext://restart_process', 'docker_build_with_restart')
load('ext://secret', 'secret_from_dict')

def helmfile(file, environment):
    return local("helmfile -f %s --environment %s template" % (file, environment))

def dlv_live_reload(service):
    docker_build_with_restart(
        'picsum-registry/%s' % service,
        context='.',
        dockerfile='./containers/Dockerfile.dev',
        entrypoint="""
            dlv debug \\
            --accept-multiclient \\
            --continue \\
            --headless \\
            --listen=:%s \\
            --api-version=2 \\
            --log \\
            --build-flags="-gcflags='all=-N -l'" \\
            ./cmd/%s
        """ % (ports[service][1], service),
        live_update=[
            sync('.', '/app/'),
        ],
    )

ports = {
    'picsum-photos': (8080, 2345),
    'image-service': (8081, 2346),
    'minio': (9000, 9001),
}

k8s_yaml(helmfile('kubernetes/helmfile.yaml', 'local'))

# picsum-photos

dlv_live_reload('picsum-photos')

k8s_yaml(secret_from_dict('picsum-hmac', inputs = {
    'hmac_key': 'foo',
}))
k8s_resource(
    new_name='picsum-hmac',
    objects=['picsum-hmac:secret'],
    labels=['picsum-photos', 'image-service'],
)

k8s_resource(
    new_name='image-manifest',
    objects=['image-manifest:configmap'],
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

# minio for a local replacement for Spaces
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
k8s_resource(
    new_name='minio-files',
    objects=['minio-files:configmap'],
    labels=['minio'],
)

k8s_yaml(secret_from_dict('picsum-spaces', inputs = {
    'access_key': 'username',
    'secret_key': 'password',
    'endpoint': 'http://minio:%s' % ports['minio'][0],
    'space': 'picsum-photos',
}))
k8s_resource(
    new_name='picsum-spaces',
    objects=['picsum-spaces:secret'],
    labels=['image-service'],
)

dlv_live_reload('image-service')
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
