build:
				protoc -I. --go_out=plugins=micro:/Users/tattoor/source/consignment/consignment-service \
					proto/consignment/consignment.proto
				go get github.com/gregory-vc/vessel-service
				go mod vendor
				docker build -t consignment-service .
run: 
				docker run -p 50051:50051 \
        -e MICRO_SERVER_ADDRESS=:50051 \
        -e MICRO_REGISTRY=mdns consignment-service