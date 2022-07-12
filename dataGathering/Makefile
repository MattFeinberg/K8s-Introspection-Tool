docker:
	sudo docker build -t introspection-data-gather .
	sudo docker tag introspection-data-gather mfeinb02/introspection-data-gather:latest
	sudo docker push mfeinb02/introspection-data-gather:latest

delete:
	kubectl delete -f pod.yaml
apply:
	kubectl apply -f pod.yaml
logs:
	kubectl --namespace introspection-resources logs introspection-tool
