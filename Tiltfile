include('./tilt-extensions/observability/Tiltfile')

image_name = 'tiltexample'
helm_chart_dir = './charts/tiltexample'
namespace = 'tiltexample'

docker_build(image_name, '.', dockerfile='Dockerfile')

# Define Helm deployment using helm function
helm_release = helm(
    helm_chart_dir,
    name=image_name,
    namespace=namespace,
    set=[
        'image.repository=' + image_name,
        'image.tag=latest'
    ]
)

# Integrate the Helm release into Tilt's resource graph
k8s_yaml(helm_release)