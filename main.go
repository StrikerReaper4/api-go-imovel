package main

import (
	"apiGo/config"
	"apiGo/controller"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

func main() {
	config.Connect()
	defer config.DB.Close()

	http.HandleFunc("/", controller.Handler)
	http.HandleFunc("/criar/usuario", controller.Create)
	http.HandleFunc("/login/usuario", controller.Login)
	http.HandleFunc("/criar/imovel", controller.CreateImovel)
	http.HandleFunc("/filtrar/imoveis", controller.FilterImovel)
	http.HandleFunc("/deletar/imovel", controller.DeleteImovel)
	http.HandleFunc("/atualizar/imovel", controller.UpdateImovel)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://frontend-imovel-bretas.onrender.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(http.DefaultServeMux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor rodando na porta %s...", port)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
