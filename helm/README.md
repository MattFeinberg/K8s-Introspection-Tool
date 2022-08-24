helm chart lives here

after updating helm chart, run these commands in helm directory and push to repo:

helm package [chart]
helm repo index --url https://MattFeinberg.github.io/K8s-Introspection-Tool/helm .

then run helm repo update wherever you want to use the helm chart
