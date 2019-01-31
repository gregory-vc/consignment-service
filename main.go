// consignment-service/main.go
package main

import (

	// Import the generated protobuf code
	"fmt"
	"log"
	"os"

	pb "github.com/gregory-vc/consignment-service/proto/consignment"

	vesselProto "github.com/gregory-vc/vessel-service/proto/vessel"

	micro "github.com/micro/go-micro"
)

func main() {

	// Database host from the environment variables
	host := os.Getenv("DB_HOST")

	session, err := CreateSession(host)

	// Mgo creates a 'master' session, we need to end that session
	// before the main function closes.
	defer session.Close()

	if err != nil {

		// We're wrapping the error returned from our CreateSession
		// here to add some context to the error.
		log.Panicf("Could not connect to datastore with host %s - %v", host, err)
	}

	// Create a new service. Optionally include some options here.
	srv := micro.NewService(

		// This name must match the package name given in your protobuf definition
		micro.Name("go.micro.srv.consignment"),
		micro.Version("latest"),
	)

	vesselClient := vesselProto.NewVesselServiceClient("go.micro.srv.vessel", srv.Client())

	// Init will parse the command line flags.
	srv.Init()

	// Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{session, vesselClient})

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
