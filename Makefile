grpc: 
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	proto/btc.proto

build:
	go build -o gateway/_output/gateway gateway/gateway.go
	go build -o server/_output/server server/server.go
	cd js && yarn run build

cluster:
	kind create cluster --config kind.yaml

weave: 
	kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"

docker-build:
	docker build -f gateway/Dockerfile.prod -t gateway gateway
	docker build -f server/Dockerfile -t server server
	docker build -f js/Dockerfile.prod -t js js
	docker build -f redis/Dockerfile -t redis redis
	
docker-tag:
	docker tag gateway manedurphy/grpc-gateway
	docker tag server manedurphy/grpc-server
	docker tag js manedurphy/grpc-js
	docker tag js manedurphy/grpc-redis

docker-push: docker-build docker-tag
	docker push manedurphy/grpc-gateway
	docker push manedurphy/grpc-server
	docker push manedurphy/grpc-js
	docker push manedurphy/grpc-redis

compose-dev: build
	docker-compose up --build

compose-prod: build
	docker-compose -f ../docker-compose.prod.yaml up --build

teardown:
	docker-compose down && docker image prune -a

gateway-image:
	go build -o gateway/gateway gateway/gateway.go && \
	docker build -t k8s/gateway:latest -f gateway/Dockerfile.prod gateway/

server-image:
	go build -o server/server server/server.go && \
	docker build -t k8s/server:latest -f server/Dockerfile server/

js-image:
	cd js && yarn run build && \
	docker build -t k8s/js:latest -f Dockerfile.prod .

redis-image:
	cd redis && \
	docker build -t k8s/redis:latest .

load: gateway-image server-image js-image redis-image
	kind load docker-image k8s/gateway:latest && \
	kind load docker-image k8s/server:latest && \
	kind load docker-image k8s/js:latest && \
	kind load docker-image k8s/redis:latest

deploy:
	# kubectl apply -f k8s/spc.yaml
	kubectl apply -f k8s/configmaps.yaml
	kubectl apply -f k8s/services.yaml
	kubectl apply -f k8s/deployments.yaml
	kubectl apply -f k8s/hpas.yaml
	kubectl apply -f k8s/networkpolicies.yaml

destroy:
	# kubectl delete -f k8s/spc.yaml
	kubectl delete -f k8s/configmaps.yaml
	kubectl delete -f k8s/services.yaml
	kubectl delete -f k8s/deployments.yaml
	kubectl delete -f k8s/hpas.yaml
	kubectl delete -f k8s/networkpolicies.yaml

forward:
	kubectl port-forward service/js 3000:80

metrics:
	kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

del-metrics:
	kubectl delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

password:
	kubectl create secret generic redis-credentials --from-literal redis-password=password

vault-deploy:
	kubectl create namespace vault
	helm install vault --namespace=vault hashicorp/vault \
		--set "server.dev.enabled=true" \
		--set "injector.enabled=false" \
		--set "csi.enabled=true"
	kubectl wait --namespace=vault --for=condition=Ready --timeout=30s pod/vault-0
	kubectl wait --namespace=vault --for=condition=Ready --timeout=30s pod -l app.kubernetes.io/name=vault-csi-provider

vault-secret:
	kubectl exec --namespace=vault vault-0 -- vault kv put secret/redis redis-password="password"
	kubectl exec --namespace=vault vault-0 -- vault kv get secret/redis

vault-auth:
	kubectl exec --namespace=vault vault-0 -- vault auth enable kubernetes
	kubectl exec --namespace=vault vault-0 -- sh -c 'vault write auth/kubernetes/config \
		token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
		kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
		kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
		issuer="https://kubernetes.default.svc.cluster.local"'

vault-policy:
	cat vault/redis-policy.hcl | kubectl exec -i --namespace=vault vault-0 -- vault policy write redis-policy -

vault-role:
	kubectl exec --namespace=vault vault-0 -- vault write auth/kubernetes/role/redis-role \
		bound_service_account_names=redis-sa \
		bound_service_account_namespaces=default \
		policies=redis-policy \
		ttl=20m	

vault: vault-deploy vault-secret vault-auth vault-policy vault-role

csi-driver:
	helm install csi --namespace=vault secrets-store-csi-driver/secrets-store-csi-driver

secrets-store: vault csi-driver

ss-destroy:
	helm uninstall --namespace=vault vault
	helm uninstall --namespace=vault csi
	kubectl delete namespace vault