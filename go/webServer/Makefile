docker:
	sudo docker build -t introspection-web-server .
	sudo docker tag introspection-web-server mfeinb02/introspection-web-server:latest
	sudo docker push mfeinb02/introspection-web-server:latest

delete:
	kubectl delete -f pod.yaml
apply:
	kubectl apply -f pod.yaml
logs:
	kubectl --namespace introspection-resources logs introspection-tool
