package main

import (
	"apiGo/config"
	"apiGo/controller"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

// gzipResponseWriter encaminha a escrita para o compressor em vez do socket direto.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// gzipMiddleware comprime as respostas quando o cliente anuncia suporte a gzip.
// O JSON da API carrega imagens em base64 — texto altamente repetitivo, que
// costuma encolher bastante. Clientes que não pedem gzip seguem sem compressão.
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Vary avisa caches (nginx, CDN) que a resposta muda conforme o header.
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		// O tamanho original deixa de valer depois de comprimir.
		w.Header().Del("Content-Length")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		next.ServeHTTP(gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
	})
}

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
		AllowedOrigins: []string{
			"http://localhost:5173",
			"https://bretasimoveis.cloud",
			"https://www.bretasimoveis.cloud",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// CORS por fora, gzip por dentro: o preflight OPTIONS responde sem passar
	// pelo compressor.
	handler := c.Handler(gzipMiddleware(http.DefaultServeMux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor rodando na porta %s...", port)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
