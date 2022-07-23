package service

import (
	usecases2 "github.com/avalido/mpc-controller/tests/mocks/mpc_server/usecases"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v4emb"
	"log"
	"net/http"
)

func ListenAndServe(port string, participants, threshold int) {
	// Set group info
	usecases2.Participants = participants
	usecases2.Threshold = threshold

	s := web.DefaultService()

	// Init API documentation schema.
	s.OpenAPI.Info.Title = "MPC mock server."
	s.OpenAPI.Info.WithDescription("This app showcases a naive mpc-server REST API.")
	s.OpenAPI.Info.Version = "v1.0.0"

	// Add use case handler to router.
	s.Post("/keygen", usecases2.Keygen())
	s.Post("/sign", usecases2.Sign())
	s.Post("/result/{reqId}", usecases2.Result())

	// Swagger UI endpoint at /docs.
	s.Docs("/docs", swgui.New)

	// Start server.
	log.Println("http://localhost:" + port + "/docs")
	if err := http.ListenAndServe(":"+port, s); err != nil {
		log.Fatal(err)
	}
}
