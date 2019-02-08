// consignment-service/main.go
package main

import (

	// Import the generated protobuf code
	"errors"
	"fmt"
	"log"
	"os"

	pb "github.com/gregory-vc/consignment-service/proto/consignment"
	"golang.org/x/net/context"

	userService "github.com/gregory-vc/user-service/proto/user"
	vesselProto "github.com/gregory-vc/vessel-service/proto/vessel"

	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/metadata"
	k8s "github.com/micro/kubernetes/go/micro"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"

	bkr "github.com/micro/go-plugins/broker/grpc"
	cli "github.com/micro/go-plugins/client/grpc"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	_ "github.com/micro/go-plugins/selector/static"
	srv "github.com/micro/go-plugins/server/grpc"
	"github.com/micro/micro/api"
)

// AuthWrapper is a high-order function which takes a HandlerFunc
// and returns a function, which takes a context, request and response interface.
// The token is extracted from the context set in our consignment-cli, that
// token is then sent over to the user service to be validated.
// If valid, the call is passed along to the handler. If not,
// an error is returned.
func AuthWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, resp interface{}) error {
		meta, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("no auth meta-data found in request")
		}

		// Note this is now uppercase (not entirely sure why this is...)
		token := meta["token"]

		if token == "" {
			return errors.New("token is empty")
		}

		log.Println("Authenticating with token: ", token)

		// Auth here
		authClient := userService.NewUserServiceClient("user", client.DefaultClient)
		_, err := authClient.ValidateToken(context.Background(), &userService.Token{
			Token: token,
		})
		if err != nil {
			return err
		}
		err = fn(ctx, req, resp)
		return err
	}
}

func main() {

	// disable namespace
	api.Namespace = ""

	// set values for registry/selector
	os.Setenv("MICRO_REGISTRY", "kubernetes")
	os.Setenv("MICRO_SELECTOR", "static")

	// setup broker/client/server
	broker.DefaultBroker = bkr.NewBroker()
	client.DefaultClient = cli.NewClient()
	server.DefaultServer = srv.NewServer()

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
	srv := k8s.NewService(

		// This name must match the package name given in your protobuf definition
		micro.Name("consignment"),
		micro.Version("latest"),
		// Our auth middleware
		micro.WrapHandler(AuthWrapper),
	)

	vesselClient := vesselProto.NewVesselServiceClient("vessel", srv.Client())

	// Init will parse the command line flags.
	srv.Init()

	// Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{session, vesselClient})

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
