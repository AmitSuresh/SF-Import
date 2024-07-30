package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) CreateMappedRecords(w http.ResponseWriter, r *http.Request) {
	p := new(Payload)
	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bi := &BulkInsert{
		AllOrNone:  false,
		Operations: []Operation{},
	}

	if len(p.PicklistMapToInsert.Eqs) > 0 {
		bi.Operations = append(bi.Operations, Operation{
			Type:    "CREATE",
			Records: convertToRecords("Measure_Equipment_Type__c", p.PicklistMapToInsert.Eqs),
		})
	}

	if len(p.PicklistMapToInsert.Recs) > 0 {
		bi.Operations = append(bi.Operations, Operation{
			Type:    "CREATE",
			Records: convertToRecords("Measure_Recommendation__c", p.PicklistMapToInsert.Recs),
		})
	}

	j, err := json.Marshal(bi)
	if err != nil {
		h.l.Error("error marshalling BulkInsert request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := h.handleNewRequest(http.MethodPost, h.uiapibatchURL, bytes.NewReader(j))
	if err != nil {
		h.l.Error("error sending request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.l.Error("error response from Salesforce", zap.Int("status_code", resp.StatusCode))
		http.Error(w, "Error from Salesforce", resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	insRes := BulkInsertResult{}

	err = FromJSON(insRes, r.Body)
	if err != nil {
		h.l.Error("error decoding body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ToJSON(insRes, w); err != nil {
		h.l.Error("error writing result", zap.Error(err))
		http.Error(w, "Error from Salesforce", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateBulkMappedRecords(w http.ResponseWriter, r *http.Request) {
	h.l.Info("")
	h.l.Info("CreateBulkMappedRecords")
	h.l.Info("")
	p := new(Payload)
	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobID, err := h.createJob(p.TargetSObject)
	if err != nil {
		h.l.Error("Error creating job:", zap.Error(err))
		return
	}

	var data [][]string
	//h.l.Info("[INFO]", zap.Any("rec_to_insert: ", p.RecordsToInsert))
	switch p.TargetSObject {
	case "Measure_Recommendation__c":

		data = [][]string{
			{"Program_Name__c", "Measure_Description__c", "Recommendation__c"},
		}

		var recommRecords RecommendationRecords

		for _, record := range p.RecordsToInsert {
			rec := &RecommendationRecord{
				ProgName:    record["Program_Name__c"].(string),
				MeasureName: record["Measure_Description__c"].(string),
				PicklistVal: record["Recommendation__c"].(string),
			}
			recommRecords = append(recommRecords, *rec)
		}

		for _, v := range recommRecords {
			data = append(data, []string{v.ProgName, v.MeasureName, v.PicklistVal})
		}
		//h.l.Info("[INFO]", zap.Any("eq rec", data))

	case "Measure_Equipment_Type__c":

		data = [][]string{
			{"Program_Name__c", "Measure_Description__c", "Equipment_Type__c"},
		}

		var eqRecords EquipmentRecords

		for _, record := range p.RecordsToInsert {
			rec := &EquipmentRecord{
				ProgName:    record["Program_Name__c"].(string),
				MeasureName: record["Measure_Description__c"].(string),
				PicklistVal: record["Equipment_Type__c"].(string),
			}
			eqRecords = append(eqRecords, *rec)
		}

		for _, v := range eqRecords {
			data = append(data, []string{v.ProgName, v.MeasureName, v.PicklistVal})
		}

		//h.l.Info("[INFO]", zap.Any("recomm rec", data))

	}
	//h.l.Info("[INFO]", zap.Any("data is", data))

	err = h.uploadBatch(jobID, data)
	if err != nil {
		h.l.Error("Error uploading batch:", zap.Error(err))
		return
	}

	err = h.closeJob(jobID)
	if err != nil {
		h.l.Error("Error closing job:", zap.Error(err))
		return
	}
}

func convertToRecords(objectAPIName string, records interface{}) []Record {
	var result []Record

	switch v := records.(type) {
	case EquipmentRecords:
		for _, record := range v {
			result = append(result, Record{
				APIName: objectAPIName,
				Fields:  record,
			})
		}
	case RecommendationRecords:
		for _, record := range v {
			result = append(result, Record{
				APIName: objectAPIName,
				Fields:  record,
			})
		}
	}
	return result
}

func (h *Handler) createJob(object string) (string, error) {
	job := map[string]string{
		"object":      object,
		"operation":   "insert",
		"contentType": "CSV",
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		h.l.Error("error marshalling create job", zap.Error(err))
		return "", err
	}
	h.l.Info("createJob: marshalled job data", zap.Any("jobData", string(jobData)))

	req, err := http.NewRequest("POST", h.ingestURL, bytes.NewBuffer(jobData))
	if err != nil {
		h.l.Error("error with creating POST request", zap.Error(err))
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+h.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		h.l.Error("error with request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	var res BulkCreateJobResult
	err = FromJSON(&res, resp.Body)
	if err != nil {
		h.l.Error("error parsing err result", zap.Error(err))
		return "", err
	}
	h.l.Info("createJob: job created", zap.String("jobID", res.ID))

	return res.ID, nil
}

func (h *Handler) uploadBatch(jobID string, data [][]string) error {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	//h.l.Info("uploadBatch: CSV data", zap.String("csvData", buffer.String()))

	url := fmt.Sprintf("%s/%s/batches", h.ingestURL, jobID)
	h.l.Info(url)
	req, err := http.NewRequest("PUT", url, &buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+h.accessToken)
	req.Header.Set("Content-Type", "text/csv")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (h *Handler) closeJob(jobID string) error {
	job := map[string]string{
		"state": "UploadComplete",
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}
	h.l.Info("createJob: marshalled job data", zap.Any("jobData", string(jobData)))
	url := fmt.Sprintf("%s/%s", h.ingestURL, jobID)
	h.l.Info(url)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jobData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+h.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
