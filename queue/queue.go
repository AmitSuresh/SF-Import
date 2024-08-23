package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	rbmq "github.com/AmitSuresh/sfdataapp/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	config   *rbmq.Config
	log      *zap.Logger
	fileLock sync.Mutex
)

func init() {

	log, _ = zap.NewProduction()
	c, err := rbmq.LoadConfig(log)
	if err != nil {
		log.Fatal("failed to load configuration", zap.Error(err))
	}
	config = c
}

func main() {

	ch, close, err := rbmq.ConnectAmqp(config, log)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}
	defer func() {
		ch.Close()
		close()
	}()

	go listen(ch)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Info("Shutdown signal received, shutting down...")

	log.Sync()
}

func listen(ch *amqp091.Channel) {
	q, err := ch.QueueDeclare(rbmq.PicklistQueryEvent, true, false, false, false, nil)
	if err != nil {
		log.Error("error queue", zap.Error(err))
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Error("error consuming messages", zap.Error(err))
	}

	forever := make(chan struct{})
	go func() {
		for d := range msgs {

			o := &rbmq.PicklistQueueRequest{}
			if err := json.Unmarshal(d.Body, o); err != nil {
				d.Nack(false, false)
				log.Error("error unmarshalling", zap.Error(err))
				continue
			}

			req, err := http.NewRequest(o.Method, o.Url, o.Body)
			if err != nil {
				log.Error("error creating request", zap.Error(err))
				continue
			}
			req.Header.Add("Authorization", "Bearer "+o.AccessToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Error("error sending request", zap.Error(err))
				continue
			}

			pickResponse := new(rbmq.PicklistQueryResponse)
			err = FromJSON(pickResponse, resp.Body)
			if err != nil {
				log.Error("error unmarshalling response", zap.Error(err))
			}
			log.Info("", zap.Any("amit ", pickResponse))

			for _, val := range pickResponse.PicklistValues {
				switch o.RecordType {
				case "Recommendation":
					updateRecPicksJSON(&rbmq.RecommendationRecord{
						PicklistVal: html.UnescapeString(val.PickValues),
						MeasureName: o.CustomObj.MeasureNameNew,
						ProgName:    o.CustomObj.ProgRec.Name,
					}, o.CustomObj.ProgRec.Name, log)
				case "Direct Install":
					updateEqPicksJSON(&rbmq.EquipmentRecord{
						PicklistVal: html.UnescapeString(val.PickValues),
						MeasureName: o.CustomObj.MeasureNameNew,
						ProgName:    o.CustomObj.ProgRec.Name,
					}, o.CustomObj.ProgRec.Name, log)
				}
			}
		}
	}()

	<-forever
}

func createJSONFile(dir, fileName string) (*os.File, error) {
	fName := fmt.Sprintf("%s.json", fileName)
	newFilename := filepath.Join(dir, fName)
	file, err := os.OpenFile(newFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func updateRecPicksJSON(targetRecords *rbmq.RecommendationRecord, programName string, log *zap.Logger) error {

	picklistMappedValues := new(rbmq.PicklistMappedResp)

	if err := os.MkdirAll(config.JsonDirPath, os.ModePerm); err != nil {
		log.Error("error creating directory", zap.Error(err))
		return err
	}

	fileLock.Lock()
	defer fileLock.Unlock()
	file, err := createJSONFile(config.JsonDirPath, programName)
	if err != nil {
		log.Error("error creating file in rec method", zap.Error(err))
		return err
	}
	defer file.Close()

	fName := fmt.Sprintf("%s/%s.json", config.JsonDirPath, programName)

	dataa, err := os.ReadFile(fName)
	if err != nil {
		log.Error("error reading le", zap.Error(err))
		return err
	}
	if len(dataa) > 0 {
		if err := json.Unmarshal(dataa, picklistMappedValues); err != nil {
			log.Error("error unmarshalling existing JSON file", zap.Error(err))
			return err
		}
	}
	picklistMappedValues.Recs = append(picklistMappedValues.Recs, rbmq.RecommendationRecord{
		PicklistVal: targetRecords.PicklistVal,
		MeasureName: targetRecords.MeasureName,
		ProgName:    targetRecords.ProgName,
	})

	updatedData, err := json.Marshal(picklistMappedValues)
	if err != nil {
		log.Error("error marshalling updated picklist values", zap.Error(err))
		return err
	}

	if _, err := file.Write(updatedData); err != nil {
		log.Error("error writing to picks.json", zap.Error(err))
		return err
	}

	log.Info("Successfully updated json file from Rec")
	return nil
}

func updateEqPicksJSON(targetRecords *rbmq.EquipmentRecord, programName string, log *zap.Logger) error {

	picklistMappedValues := new(rbmq.PicklistMappedResp)

	if err := os.MkdirAll(config.JsonDirPath, os.ModePerm); err != nil {
		log.Error("error creating directory", zap.Error(err))
		return err
	}

	fileLock.Lock()
	defer fileLock.Unlock()
	file, err := createJSONFile(config.JsonDirPath, programName)
	if err != nil {
		log.Error("error creating file in eq method", zap.Error(err))
		return err
	}
	defer file.Close()

	fName := fmt.Sprintf("%s/%s.json", config.JsonDirPath, programName)

	dataa, err := os.ReadFile(fName)
	if err != nil {
		log.Error("error reading le", zap.Error(err))
		return err
	}
	if len(dataa) > 0 {
		if err := json.Unmarshal(dataa, picklistMappedValues); err != nil {
			log.Error("error unmarshalling existing JSON file", zap.Error(err))
			return err
		}
	}
	picklistMappedValues.Eqs = append(picklistMappedValues.Eqs, rbmq.EquipmentRecord{
		PicklistVal: targetRecords.PicklistVal,
		MeasureName: targetRecords.MeasureName,
		ProgName:    targetRecords.ProgName,
	})

	updatedData, err := json.Marshal(picklistMappedValues)
	if err != nil {
		log.Error("error marshalling updated picklist values", zap.Error(err))
		return err
	}

	if _, err := file.Write(updatedData); err != nil {
		log.Error("error writing to picks.json", zap.Error(err))
		return err
	}

	log.Info("Successfully updated json file from Eq")
	return nil
}

func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func FromJSON(i interface{}, r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}
