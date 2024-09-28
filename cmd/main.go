package main

import (
	"fmt"
	"log"

	"github.com/groshiniprasad/uploady/cmd/api"
	"github.com/groshiniprasad/uploady/configs"
)

func main() {

	server := api.NewAPIServer(fmt.Sprintf(":%s", configs.Envs.Port))
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
