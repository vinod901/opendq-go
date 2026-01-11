package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vinod901/opendq-go/internal/alerting"
	"github.com/vinod901/opendq-go/internal/check"
	"github.com/vinod901/opendq-go/internal/datasource"
	"github.com/vinod901/opendq-go/internal/scheduler"
	"github.com/vinod901/opendq-go/internal/view"
)

// DataQualityHandler handles data quality related HTTP endpoints
type DataQualityHandler struct {
	datasourceManager *datasource.Manager
	checkManager      *check.Manager
	schedulerManager  *scheduler.Manager
	alertManager      *alerting.Manager
	viewManager       *view.Manager
}

// NewDataQualityHandler creates a new data quality handler
func NewDataQualityHandler(
	datasourceManager *datasource.Manager,
	checkManager *check.Manager,
	schedulerManager *scheduler.Manager,
	alertManager *alerting.Manager,
	viewManager *view.Manager,
) *DataQualityHandler {
	return &DataQualityHandler{
		datasourceManager: datasourceManager,
		checkManager:      checkManager,
		schedulerManager:  schedulerManager,
		alertManager:      alertManager,
		viewManager:       viewManager,
	}
}

// RegisterRoutes registers data quality routes
func (h *DataQualityHandler) RegisterRoutes(mux *http.ServeMux) {
	// Datasource routes
	mux.HandleFunc("/api/v1/datasources", h.handleDatasources)
	mux.HandleFunc("/api/v1/datasources/", h.handleDatasource)
	mux.HandleFunc("/api/v1/datasources/test", h.testDatasourceConnection)

	// Check routes
	mux.HandleFunc("/api/v1/checks", h.handleChecks)
	mux.HandleFunc("/api/v1/checks/", h.handleCheck)

	// Schedule routes
	mux.HandleFunc("/api/v1/schedules", h.handleSchedules)
	mux.HandleFunc("/api/v1/schedules/", h.handleSchedule)

	// Alert channel routes
	mux.HandleFunc("/api/v1/alerts/channels", h.handleAlertChannels)
	mux.HandleFunc("/api/v1/alerts/channels/", h.handleAlertChannel)
	mux.HandleFunc("/api/v1/alerts/history", h.getAlertHistory)

	// View routes
	mux.HandleFunc("/api/v1/views", h.handleViews)
	mux.HandleFunc("/api/v1/views/", h.handleView)
}

// Helper to extract ID from path
func extractIDFromPath(path, prefix string) string {
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// Datasource handlers

func (h *DataQualityHandler) handleDatasources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listDatasources(w, r)
	case http.MethodPost:
		h.createDatasource(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) handleDatasource(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v1/datasources")

	// Check for sub-resources
	if strings.Contains(r.URL.Path, "/checks") {
		h.listDatasourceChecks(w, r, id)
		return
	}
	if strings.Contains(r.URL.Path, "/tables") {
		h.listDatasourceTables(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getDatasource(w, r, id)
	case http.MethodPut:
		h.updateDatasource(w, r, id)
	case http.MethodDelete:
		h.deleteDatasource(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) listDatasources(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")

	datasources, err := h.datasourceManager.ListDatasources(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(datasources)
}

func (h *DataQualityHandler) createDatasource(w http.ResponseWriter, r *http.Request) {
	var ds datasource.Datasource
	if err := json.NewDecoder(r.Body).Decode(&ds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.datasourceManager.CreateDatasource(r.Context(), &ds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ds)
}

func (h *DataQualityHandler) getDatasource(w http.ResponseWriter, r *http.Request, id string) {
	ds, err := h.datasourceManager.GetDatasource(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ds)
}

func (h *DataQualityHandler) updateDatasource(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.datasourceManager.UpdateDatasource(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ds, _ := h.datasourceManager.GetDatasource(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ds)
}

func (h *DataQualityHandler) deleteDatasource(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.datasourceManager.DeleteDatasource(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DataQualityHandler) testDatasourceConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ds datasource.Datasource
	if err := json.NewDecoder(r.Body).Decode(&ds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.datasourceManager.TestConnection(r.Context(), &ds); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

func (h *DataQualityHandler) listDatasourceTables(w http.ResponseWriter, r *http.Request, id string) {
	connector, err := h.datasourceManager.GetConnector(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tables, err := connector.GetTables(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)
}

func (h *DataQualityHandler) listDatasourceChecks(w http.ResponseWriter, r *http.Request, id string) {
	checks, err := h.checkManager.ListChecks(r.Context(), "", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checks)
}

// Check handlers

func (h *DataQualityHandler) handleChecks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listChecks(w, r)
	case http.MethodPost:
		h.createCheck(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v1/checks")

	// Check for sub-resources
	if strings.Contains(r.URL.Path, "/run") {
		h.runCheck(w, r, id)
		return
	}
	if strings.Contains(r.URL.Path, "/results") {
		h.getCheckResults(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getCheck(w, r, id)
	case http.MethodPut:
		h.updateCheck(w, r, id)
	case http.MethodDelete:
		h.deleteCheck(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) listChecks(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	datasourceID := r.URL.Query().Get("datasource_id")

	checks, err := h.checkManager.ListChecks(r.Context(), tenantID, datasourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checks)
}

func (h *DataQualityHandler) createCheck(w http.ResponseWriter, r *http.Request) {
	var chk check.Check
	if err := json.NewDecoder(r.Body).Decode(&chk); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.checkManager.CreateCheck(r.Context(), &chk); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chk)
}

func (h *DataQualityHandler) getCheck(w http.ResponseWriter, r *http.Request, id string) {
	chk, err := h.checkManager.GetCheck(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chk)
}

func (h *DataQualityHandler) updateCheck(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.checkManager.UpdateCheck(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chk, _ := h.checkManager.GetCheck(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chk)
}

func (h *DataQualityHandler) deleteCheck(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.checkManager.DeleteCheck(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DataQualityHandler) runCheck(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.checkManager.RunCheck(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *DataQualityHandler) getCheckResults(w http.ResponseWriter, r *http.Request, id string) {
	results, err := h.checkManager.GetCheckResults(r.Context(), id, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Schedule handlers

func (h *DataQualityHandler) handleSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSchedules(w, r)
	case http.MethodPost:
		h.createSchedule(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) handleSchedule(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v1/schedules")

	// Check for sub-resources
	if strings.Contains(r.URL.Path, "/run") {
		h.runScheduleNow(w, r, id)
		return
	}
	if strings.Contains(r.URL.Path, "/executions") {
		h.getScheduleExecutions(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getSchedule(w, r, id)
	case http.MethodPut:
		h.updateSchedule(w, r, id)
	case http.MethodDelete:
		h.deleteSchedule(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) listSchedules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")

	schedules, err := h.schedulerManager.ListSchedules(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

func (h *DataQualityHandler) createSchedule(w http.ResponseWriter, r *http.Request) {
	var schedule scheduler.Schedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.schedulerManager.CreateSchedule(r.Context(), &schedule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

func (h *DataQualityHandler) getSchedule(w http.ResponseWriter, r *http.Request, id string) {
	schedule, err := h.schedulerManager.GetSchedule(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

func (h *DataQualityHandler) updateSchedule(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.schedulerManager.UpdateSchedule(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	schedule, _ := h.schedulerManager.GetSchedule(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

func (h *DataQualityHandler) deleteSchedule(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.schedulerManager.DeleteSchedule(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DataQualityHandler) runScheduleNow(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	execution, err := h.schedulerManager.RunScheduleNow(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

func (h *DataQualityHandler) getScheduleExecutions(w http.ResponseWriter, r *http.Request, id string) {
	executions, err := h.schedulerManager.GetScheduleExecutions(r.Context(), id, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

// Alert channel handlers

func (h *DataQualityHandler) handleAlertChannels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAlertChannels(w, r)
	case http.MethodPost:
		h.createAlertChannel(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) handleAlertChannel(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v1/alerts/channels")

	// Check for sub-resources
	if strings.Contains(r.URL.Path, "/test") {
		h.testAlertChannel(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getAlertChannel(w, r, id)
	case http.MethodPut:
		h.updateAlertChannel(w, r, id)
	case http.MethodDelete:
		h.deleteAlertChannel(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) listAlertChannels(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")

	channels, err := h.alertManager.ListChannels(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

func (h *DataQualityHandler) createAlertChannel(w http.ResponseWriter, r *http.Request) {
	var channel alerting.Channel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.alertManager.CreateChannel(r.Context(), &channel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

func (h *DataQualityHandler) getAlertChannel(w http.ResponseWriter, r *http.Request, id string) {
	channel, err := h.alertManager.GetChannel(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channel)
}

func (h *DataQualityHandler) updateAlertChannel(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.alertManager.UpdateChannel(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	channel, _ := h.alertManager.GetChannel(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channel)
}

func (h *DataQualityHandler) deleteAlertChannel(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.alertManager.DeleteChannel(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DataQualityHandler) testAlertChannel(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.alertManager.TestChannel(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test alert sent successfully",
	})
}

func (h *DataQualityHandler) getAlertHistory(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")

	history, err := h.alertManager.GetAlertHistory(r.Context(), channelID, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// View handlers

func (h *DataQualityHandler) handleViews(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listViews(w, r)
	case http.MethodPost:
		h.createView(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) handleView(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v1/views")

	// Check for sub-resources
	if strings.Contains(r.URL.Path, "/query") {
		h.queryView(w, r, id)
		return
	}
	if strings.Contains(r.URL.Path, "/validate") {
		h.validateView(w, r, id)
		return
	}
	if strings.Contains(r.URL.Path, "/sql") {
		h.getViewSQL(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getView(w, r, id)
	case http.MethodPut:
		h.updateView(w, r, id)
	case http.MethodDelete:
		h.deleteView(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *DataQualityHandler) listViews(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	datasourceID := r.URL.Query().Get("datasource_id")

	views, err := h.viewManager.ListViews(r.Context(), tenantID, datasourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

func (h *DataQualityHandler) createView(w http.ResponseWriter, r *http.Request) {
	var v view.View
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.viewManager.CreateView(r.Context(), &v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

func (h *DataQualityHandler) getView(w http.ResponseWriter, r *http.Request, id string) {
	v, err := h.viewManager.GetView(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *DataQualityHandler) updateView(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.viewManager.UpdateView(r.Context(), id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	v, _ := h.viewManager.GetView(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *DataQualityHandler) deleteView(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.viewManager.DeleteView(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DataQualityHandler) queryView(w http.ResponseWriter, r *http.Request, id string) {
	limit := 100 // Default limit

	result, err := h.viewManager.QueryView(r.Context(), id, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *DataQualityHandler) validateView(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.viewManager.ValidateView(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"message": "View is valid",
	})
}

func (h *DataQualityHandler) getViewSQL(w http.ResponseWriter, r *http.Request, id string) {
	sql, err := h.viewManager.GetViewSQL(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sql": sql,
	})
}
