package api

import (
	"net/http"

	"bulkmailer/internal/utils"
	"bulkmailer/internal/worker"
)

type API struct {
	Worker *worker.Worker
}

func NewAPI(w *worker.Worker) *API {
	return &API{Worker: w}
}

func (a *API) HandleStart(w http.ResponseWriter, r *http.Request) {
	if err := a.Worker.StartWorker(); err != nil {
		utils.HTTPError(w, err, 500)
		return
	}

	csvPath := r.URL.Query().Get("csv")
	if csvPath == "" {
		csvPath = a.Worker.GetCSVPath()
	}

	count, err := a.Worker.ParseCSVAndEnqueue(csvPath)
	if err != nil {
		utils.HTTPError(w, err, 500)
		return
	}

	utils.WriteJSON(w, map[string]interface{}{"enqueued": count})
}

func (a *API) HandleStop(w http.ResponseWriter, r *http.Request) {
	if err := a.Worker.StopWorker(); err != nil {
		utils.HTTPError(w, err, 500)
		return
	}
	utils.WriteJSON(w, map[string]string{"status": "stopped"})
}

func (a *API) HandlePause(w http.ResponseWriter, r *http.Request) {
	if err := a.Worker.PauseProcessing(r.Context()); err != nil {
		utils.HTTPError(w, err, 500)
		return
	}
	utils.WriteJSON(w, map[string]string{"status": "paused"})
}

func (a *API) HandleResume(w http.ResponseWriter, r *http.Request) {
	if err := a.Worker.ResumeProcessing(r.Context()); err != nil {
		utils.HTTPError(w, err, 500)
		return
	}
	utils.WriteJSON(w, map[string]string{"status": "resumed"})
}

func (a *API) HandleMonitor(w http.ResponseWriter, r *http.Request) {
	info := a.Worker.GetMonitorInfo(r.Context())
	utils.WriteJSON(w, info)
}
