grpc: 
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	proto/btc.proto

build:
	go build -o gateway/_output/gateway gateway/gateway.go
	go build -o server/_output/server server/server.go
	
react-build:
	cd js && yarn run build

cluster:
	kind create cluster --config kind.yaml

cli:
	cd kubeconfig && docker build -t linode .
	docker run --rm -it -v $(shell pwd):/work -w /work --entrypoint /bin/bash linode

weave: 
	kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(shell kubectl version | base64 | tr -d '\n')"

ingress-controller:
	kubectl create namespace ingress-nginx
	helm install --namespace=ingress-nginx ingress-nginx ingress-nginx/ingress-nginx
	kubectl wait --namespace=ingress-nginx --for=condition=Ready --timeout=5m pod -l app.kubernetes.io/name=ingress-nginx

ingress-destroy:
	kubectl delete namespace ingress-nginx --force --grace-period=0

docker-build: build react-build
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

compose: build
	docker-compose up --build

teardown:
	docker-compose down && docker image prune -a

gateway-image:
	docker build -t k8s/gateway:latest -f gateway/Dockerfile.prod gateway/

server-image:
	docker build -t k8s/server:latest -f server/Dockerfile server/

js-image:
	docker build -t k8s/js:latest -f js/Dockerfile.prod js/

redis-image:
	docker build -t k8s/redis:latest -f redis/Dockerfile redis/

load: build react-build gateway-image server-image js-image redis-image
	kind load docker-image k8s/gateway:latest
	kind load docker-image k8s/server:latest
	kind load docker-image k8s/js:latest
	kind load docker-image k8s/redis:latest

deploy:
	kubectl create namespace crypto-charts
	kubectl create --namespace=crypto-charts secret generic crypto-token --from-literal CRYPTO_API_KEY=$$CRYPTO_API_KEY
	kubectl --namespace=crypto-charts apply -f k8s/spc.yaml
	kubectl apply -f k8s/metrics-server.yaml
	kubectl --namespace=crypto-charts create secret generic redis-credentials --from-literal redis-password=password
	kubectl --namespace=crypto-charts apply -f k8s/configmaps.yaml
	kubectl --namespace=crypto-charts apply -f k8s/services.yaml
	kubectl --namespace=crypto-charts apply -f k8s/deployments.yaml
	kubectl --namespace=crypto-charts apply -f k8s/ingress.yaml
	kubectl --namespace=crypto-charts apply -f k8s/hpas.yaml
	kubectl --namespace=crypto-charts apply -f k8s/statefulsets/redis-statefulset.yaml
	kubectl wait --namespace=crypto-charts --for=condition=Ready --timeout=5m pod -l statefulset.kubernetes.io/pod-name=redis-0
	kubectl wait --namespace=crypto-charts --for=condition=Ready --timeout=5m pod -l statefulset.kubernetes.io/pod-name=redis-1
	kubectl wait --namespace=crypto-charts --for=condition=Ready --timeout=5m pod -l statefulset.kubernetes.io/pod-name=redis-2
	kubectl --namespace=crypto-charts apply -f k8s/statefulsets/sentinel-statefulset.yaml
	# kubectl --namespace=crypto-charts apply -f k8s/networkpolicies.yaml

destroy:
	kubectl delete namespace crypto-charts --force --grace-period=0

forward:
	kubectl config set-context --current --namespace=ingress-nginx
	kubectl port-forward service/ingress-nginx-controller 3000:80

metrics:
	kubectl apply -f k8s/metrics-server.yaml

del-metrics:
	kubectl delete -f k8s/metrics-server.yaml

tls:
	cd certs && ./gen.sh
	cp certs/server.key server/tls
	cp certs/server.crt server/tls
	cp certs/ca.crt gateway/tls

vault-deploy:
	kubectl create namespace vault
	helm install vault --namespace=vault hashicorp/vault \
		--set "server.dev.enabled=true" \
		--set "injector.enabled=false" \
		--set "csi.enabled=true"
	kubectl wait --namespace=vault --for=condition=Ready --timeout=5m pod/vault-0
	kubectl wait --namespace=vault --for=condition=Ready --timeout=5m pod -l app.kubernetes.io/name=vault-csi-provider

vault-secret:
	kubectl exec --namespace=vault vault-0 -- vault kv put secret/redis redis-password="password"
	kubectl exec --namespace=vault vault-0 -- vault kv get secret/redis

vault-auth:
	kubectl exec --namespace=vault vault-0 -- vault auth enable kubernetes
	kubectl exec --namespace=vault vault-0 -- sh -c 'vault write auth/kubernetes/config \
		token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
		kubernetes_host="https://$$KUBERNETES_PORT_443_TCP_ADDR:443" \
		kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
		issuer="https://kubernetes.default.svc.cluster.local"'

vault-policy:
	cat vault/redis-policy.hcl | kubectl exec -i --namespace=vault vault-0 -- vault policy write redis-policy -

vault-role:
	kubectl exec --namespace=vault vault-0 -- vault write auth/kubernetes/role/redis-role \
		bound_service_account_names=redis-sa \
		bound_service_account_namespaces=crypto-charts \
		policies=redis-policy \
		ttl=20m	

vault: vault-deploy vault-secret vault-auth vault-policy vault-role

csi-driver:
	helm install csi --namespace=vault secrets-store-csi-driver/secrets-store-csi-driver
	kubectl wait --namespace=vault --for=condition=Ready --timeout=5m pod -l app.kubernetes.io/name=secrets-store-csi-driver

secrets-store: vault csi-driver

ss-destroy:
	kubectl delete namespace vault --force --grace-period=0

crypto-charts: tls cluster secrets-store ingress-controller load deploy forward

kill-all: destroy ss-destroy ingress-destroy
	kubectl config set-context --current --namespace=default

kill-cluster:
	kind delete cluster