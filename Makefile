buildbuilder: # call in /Papers
	docker build -t "nekkkkitch/docker" -f .\Dockerfile .
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