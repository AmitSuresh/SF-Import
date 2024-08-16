package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	rbmq "github.com/AmitSuresh/sfdataapp/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
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

	if len(customRecMap["Recommendation"]) > 0 && len(customRecMap["Direct Install"]) > 0 {
		processMapping := func(recordType string, fieldName string) error {
			for _, v := range customRecMap[recordType] {
				p.RecTypeID = v.RecTypeId
				p.FieldName = fieldName
				pickURL := fmt.Sprintf("%s%s/picklist-values/%s/%s", h.uiapiURL, p.SObject, p.RecTypeID, p.FieldName)
				h.l.Info(pickURL)

				q, err := h.amqpCh.QueueDeclare(rbmq.PicklistQueryEvent, true, false, false, false, nil)
				if err != nil {
					h.l.Error("error declaring queue", zap.Error(err))
					return err
				}

				request := &rbmq.PicklistQueueRequest{
					Method:      http.MethodGet,
					Url:         pickURL,
					Body:        nil,
					AccessToken: h.accessToken,
					CustomObj: rbmq.CustomRecords{
						Id:             v.Id,
						MeasureNameNew: v.MeasureNameNew,
						RecTypeName:    v.RecTypeName,
						RecTypeId:      v.RecTypeId,
						ProgRec:        rbmq.ProgramRecord(v.ProgRec),
					},
					RecordType: recordType,
				}

				marshalledReq, err := json.Marshal(request)
				if err != nil {
					h.l.Error("error marshalling", zap.Error(err))
				}

				err = h.amqpCh.PublishWithContext(r.Context(), "", q.Name, false, false, amqp091.Publishing{
					ContentType: "application/json",
					Body:        marshalledReq,
				})
				if err != nil {
					h.l.Error("error publishing request", zap.Error(err))
					return err
				}
				h.l.Info("Published to queue", zap.String("queue", q.Name))
			}
			return nil
		}

		if err := processMapping("Recommendation", "Recommendation__c"); err != nil {
			h.l.Error("error processing recommendations", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := processMapping("Direct Install", "Equipment_Type__c"); err != nil {
			h.l.Error("error processing equipment", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Request is placed successfully."))
	if err != nil {
		h.l.Error("error writing response", zap.Error(err))
	}
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
