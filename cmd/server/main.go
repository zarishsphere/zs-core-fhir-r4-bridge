// ZarishSphere FHIR R4↔R5 Bridge — HTTP server.
// Translates between FHIR R4 (4.0.1) and FHIR R5 (5.0.0) on demand.
// Used by partner integrations: DHIS2, OpenMRS, legacy HIS systems.
// ADR-0001: Go 1.26.1 | Governance: BLUEPRINT.md Layer 2 (Interoperability)
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zarishsphere/zs-core-fhir-r4-bridge/internal/translator"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Str("service", "zs-core-fhir-r4-bridge").Logger()

	addr := getEnv("SERVER_ADDR", ":8085")

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// Health
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(
			`{"status":"ok","service":"zs-core-fhir-r4-bridge","versions":["4.0.1","5.0.0"]}`,
		))
	})

	// Single resource translation
	r.Post("/translate/r4-to-r5", translateHandler(translator.R4ToR5))
	r.Post("/translate/r5-to-r4", translateHandler(translator.R5ToR4))

	// Batch (Bundle) translation
	r.Post("/translate/batch/r4-to-r5", batchHandler(translator.R4ToR5))
	r.Post("/translate/batch/r5-to-r4", batchHandler(translator.R5ToR4))

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		q := make(chan os.Signal, 1)
		signal.Notify(q, syscall.SIGTERM, syscall.SIGINT)
		<-q
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		close(done)
	}()

	log.Info().Str("addr", addr).Msg("r4-bridge: listening")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("server error")
	}
	<-done
	log.Info().Msg("r4-bridge: stopped")
}

type translateFn func(map[string]any) (*translator.TranslationResult, error)

func translateHandler(fn translateFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resource map[string]any
		if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
			fhirError(w, http.StatusBadRequest, "structure", "Invalid JSON: "+err.Error())
			return
		}
		result, err := fn(resource)
		if err != nil {
			fhirError(w, http.StatusUnprocessableEntity, "processing", err.Error())
			return
		}
		writeFHIR(w, result)
	}
}

func batchHandler(fn translateFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var bundle map[string]any
		if err := json.NewDecoder(r.Body).Decode(&bundle); err != nil {
			fhirError(w, http.StatusBadRequest, "structure", "Invalid JSON")
			return
		}
		if rt, _ := bundle["resourceType"].(string); rt != "Bundle" {
			fhirError(w, http.StatusBadRequest, "structure", "Expected Bundle resourceType for batch translation")
			return
		}
		result, err := fn(bundle)
		if err != nil {
			fhirError(w, http.StatusUnprocessableEntity, "processing", err.Error())
			return
		}
		writeFHIR(w, result)
	}
}

func writeFHIR(w http.ResponseWriter, result *translator.TranslationResult) {
	w.Header().Set("Content-Type", "application/fhir+json; charset=utf-8")
	if len(result.LossyFields) > 0 {
		w.Header().Set("X-ZS-Lossy-Fields", strings.Join(result.LossyFields, ","))
		w.Header().Set("X-ZS-Translation-Warnings", "true")
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result.Resource)
}

func fhirError(w http.ResponseWriter, status int, code, diag string) {
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(
		`{"resourceType":"OperationOutcome","issue":[{"severity":"error","code":"` + code +
			`","diagnostics":"` + diag + `"}]}`,
	))
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
