network:
	docker network create papersnetwork
stop:
	docker-compose stop \
	&& docker-compose rm 
start:
	docker-compose build --no-cache \
	&& docker-compose up -d
buildauthpb: 
	protoc --proto_path=pkg/grpc/proto/authService --go_out=pkg/grpc/pb/authService --go-grpc_out=pkg/grpc/pb/authService pkg/grpc/proto/authService/*.proto
buildpaperspb: 
	protoc --proto_path=pkg/grpc/proto/papersService --go_out=pkg/grpc/pb/papersService --go-grpc_out=pkg/grpc/pb/papersService pkg/grpc/proto/papersService/*.proto
buildbalancepb: 
	protoc --proto_path=pkg/grpc/proto/balanceService --go_out=pkg/grpc/pb/balanceService --go-grpc_out=pkg/grpc/pb/balanceService pkg/grpc/proto/balanceService/*.proto
runmarket:
	docker build -t market ./services/market
	docker run -d --network papersnetwork --name market market
runbalance:
	docker build -t balance ./services/balance
	docker run -d --network papersnetwork --name balance balance
runpapers:
	docker build -t papers ./services/papers
	docker run -d --network papersnetwork --name papers papers
runaus:
	docker build -t aus ./services/aus
	docker run -d --network papersnetwork --name aus aus
rungateway:
	docker build -t gateway ./services/gateway
	docker run -d --network papersnetwork -p 8080:8080 --name gateway gateway
runall:
	make runmarket
	make runbalance
	make runpapers
	make runaus
	make rungateway
killall:
	docker stop market
	docker rm market
	docker stop balance
	docker rm balance
	docker stop papers
	docker rm papers
	docker stop aus
	docker rm aus
	docker stop gateway
	docker rm gateway