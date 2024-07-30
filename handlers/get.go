package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

func (h *Handler) GetPickBasedMappingRec(w http.ResponseWriter, r *http.Request) {
	p := new(Payload)
	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p.Records == nil {
		h.l.Error("custom object records are not found in payload.", zap.Error(err))
		http.Error(w, "error reading records from payload", http.StatusBadRequest)
		return
	}

	customRecMap := make(CustomRecordsMap)
	for _, v := range p.Records {
		if strings.Contains(v.MeasureNameNew, "Recommendation") {
			customRecMap["Recommendation"] = append(customRecMap["Recommendation"], v)
		} else {
			customRecMap["Direct Install"] = append(customRecMap["Direct Install"], v)
		}
	}

	h.l.Info("\n", zap.Any("", customRecMap))

	var eq EquipmentRecords
	var rec RecommendationRecords

	if len(customRecMap["Recommendation"]) > 0 && len(customRecMap["Direct Install"]) > 0 {
		processMapping := func(recordType string, fieldName string, targetRecords interface{}) error {
			for _, v := range customRecMap[recordType] {
				p.RecTypeID = v.RecTypeId
				p.FieldName = fieldName
				pickURL := fmt.Sprintf("%s%s/picklist-values/%s/%s", h.uiapiURL, p.SObject, p.RecTypeID, p.FieldName)
				h.l.Info(pickURL)

				resp, err := h.handleNewRequest(http.MethodGet, pickURL, nil)
				if err != nil {
					h.l.Error("error with request", zap.Error(err))
				}
				defer resp.Body.Close()

				pickResponse := new(PicklistQueryResponse)
				err = FromJSON(pickResponse, resp.Body)
				if err != nil {
					h.l.Error("error unmarshalling response:", zap.Error(err))
				}

				h.l.Info("\n", zap.Any("", pickResponse))
				for _, val := range pickResponse.PicklistValues {
					switch records := targetRecords.(type) {
					case *RecommendationRecords:
						*records = append(*records, RecommendationRecord{
							PicklistVal: html.UnescapeString(val.PickValues),
							MeasureName: v.MeasureNameNew,
							ProgName:    v.ProgRec.Name,
						})
					case *EquipmentRecords:
						*records = append(*records, EquipmentRecord{
							PicklistVal: html.UnescapeString(val.PickValues),
							MeasureName: v.MeasureNameNew,
							ProgName:    v.ProgRec.Name,
						})
					}
				}
			}
			return nil
		}

		if err := processMapping("Recommendation", "Recommendation__c", &rec); err != nil {
			h.l.Error("error processing recommendations", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := processMapping("Direct Install", "Equipment_Type__c", &eq); err != nil {
			h.l.Error("error processing equipment", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res := PicklistMappedResp{
			Eqs:  eq,
			Recs: rec,
		}
		err = ToJSON(res, w)
		if err != nil {
			h.l.Error("error marshalling", zap.Error(err))
		}
		h.l.Info("\n", zap.Any("recommendation map", rec))
		h.l.Info("\n", zap.Any("equipment map", eq))
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) QueryRecords(w http.ResponseWriter, r *http.Request) {
	p := new(Payload)
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		h.l.Error("error decoding body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qURL := fmt.Sprintf("%s%s", h.queryURL, url.QueryEscape(p.Query))

	resp, err := h.handleNewRequest(http.MethodGet, qURL, nil)
	if err != nil {
		h.l.Error("error sending request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var qResp QueryResponse
	err = FromJSON(&qResp, resp.Body)
	if err != nil {
		h.l.Error("error unmarshalling response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ToJSON(&qResp, w)
	if err != nil {
		h.l.Error("error writing result", zap.Error(err))
	}
	w.WriteHeader(http.StatusOK)
}
