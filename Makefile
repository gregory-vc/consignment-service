build:
				protoc -I. --go_out=plugins=micro:. proto/consignment/consignment.proto
				go get github.com/gregory-vc/vessel-service
				go mod vendor
				git add --all
				git diff-index --quiet HEAD || git commit -a -m 'fix'
				git push origin master
run: 
				docker run -p 50051:50051 \
        -e MICRO_SERVER_ADDRESS=:50051 \
        -e MICRO_REGISTRY=mdns consignment-service