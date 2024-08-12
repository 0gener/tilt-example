load('ext://helm_remote', 'helm_remote')

include('./tilt-extensions/observability/Tiltfile')

k8s_yaml([blob("""
apiVersion: v1
kind: Namespace
metadata:
    name: postgres
""")])

k8s_yaml([blob("""
apiVersion: v1
kind: Namespace
metadata:
    name: localstack
""")])

postgres_release_name= 'postgres'
postgres_namespace= 'postgres'
postgres_username = 'apiservice'
postgres_password = 'password'
postgres_database = 'apiservice'
postgres_service_name = 'postgres-postgresql'
postgres_host = '{}.{}.svc.cluster.local'.format(postgres_service_name, postgres_namespace)
postgres_port = '5432'

database_connection_string = 'postgres://{}:{}@{}:{}/{}?sslmode=disable'.format(
    postgres_username, postgres_password, postgres_host, postgres_port, postgres_database
)

helm_remote(
    'postgresql',
    repo_name='bitnami',
    repo_url='https://charts.bitnami.com/bitnami',
    release_name=postgres_release_name,
    namespace=postgres_namespace,
    set=[
        'auth.username={}'.format(postgres_username),
        'auth.password={}'.format(postgres_password),
        'auth.database={}'.format(postgres_database),
    ]
)

# localstack
helm_remote(
    'localstack',
    repo_name='localstack',
    release_name='localstack',
    repo_url='https://helm.localstack.cloud',
    namespace='localstack',
    set=[
        'startServices=sns\\,sqs',
        'service.type=ClusterIP',
        'service.ports={}'.format(4566),
    ]
)

localstack_pod = local('kubectl get pods -n localstack -l "app.kubernetes.io/name=localstack" -o jsonpath="{.items[0].metadata.name}"')

local('kubectl exec -n localstack {} -- awslocal sns create-topic --name events-topic'.format(localstack_pod))
local('kubectl exec -n localstack {} -- awslocal sqs create-queue --queue-name eventconsumerservice-queue'.format(localstack_pod))
local('kubectl exec -n localstack {} -- awslocal sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:events-topic --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:eventconsumerservice-queue'.format(localstack_pod))

# apiservice
apiservice_image_name = 'apiservice'
apiservice_helm_chart_dir = './charts/apiservice'
apiservice_namespace = 'apiservice'

docker_build(apiservice_image_name, '.', dockerfile='./docker/apiservice/Dockerfile')

apiservice_helm_release = helm(
    apiservice_helm_chart_dir,
    name=apiservice_image_name,
    namespace=apiservice_namespace,
    set=[
        'image.repository=' + apiservice_image_name,
        'image.tag=latest',
        'database.connection_string=' + database_connection_string,
        'messaging.aws_endpoint=http://localstack.localstack.svc.cluster.local:4566',
        'messaging.events_topic_arn=arn:aws:sns:us-east-1:000000000000:events-topic',
    ]
)

k8s_yaml(apiservice_helm_release)

# eventconsumerservice
eventconsumerservice_image_name = 'eventconsumerservice'
eventconsumerservice_helm_chart_dir = './charts/eventconsumerservice'
eventconsumerservice_namespace = 'eventconsumerservice'

docker_build(eventconsumerservice_image_name, '.', dockerfile='./docker/eventconsumerservice/Dockerfile')

eventconsumerservice_helm_release = helm(
    eventconsumerservice_helm_chart_dir,
    name=eventconsumerservice_image_name,
    namespace=eventconsumerservice_namespace,
    set=[
        'image.repository=' + eventconsumerservice_image_name,
        'image.tag=latest',
        'messaging.aws_endpoint=http://localstack.localstack.svc.cluster.local:4566',
        'messaging.events_queue_url=http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/eventconsumerservice-queue',
    ]
)

k8s_yaml(eventconsumerservice_helm_release)