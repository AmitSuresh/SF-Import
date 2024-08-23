package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

func (h *Handler) GetMeasureCalcs(w http.ResponseWriter, r *http.Request) {
	p := new(MeasureCalcsResponse)

	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding", zap.Error(err))
		return
	}

	filesCreated := make(map[string]*os.File)
	defer func() {
		for _, file := range filesCreated {
			file.Close()
		}
	}()

	for _, v := range p.MeasureCalcRecords {
		fileName := strings.ReplaceAll(v.Name, "/", "-")
		h.l.Info("", zap.Any("file Name is: ", fileName))

		file, err := h.CreateCSVFile(h.cfg.JsonDirPath, fileName)
		if err != nil {
			h.l.Error("error creating or opening a csv file", zap.Error(err))
			return
		}
		defer file.Close()

		fileName = fmt.Sprintf("%s/%s.csv", h.cfg.JsonDirPath, fileName)

		d, err := os.ReadFile(fileName)
		if err != nil {
			h.l.Error("error reading file", zap.Error(err))
		}

		var records [][]string

		r := csv.NewReader(strings.NewReader(string(d)))

		records, err = r.ReadAll()
		if err != nil {
			h.l.Error("error reading CSV data", zap.Error(err))
		}

		filesCreated[v.Name] = file
		h.l.Info("size of first records", zap.Any("", len(records)))

		switch len(records) {
		case 0:
			records = append(records, []string{"Id", "Name", "Program_Name__c", "CLR_CNI_Sequence__c", "CLI_CNR_Field_to_calculate__c", "CLR_CNI_Mesaure_Formula__c"})
			fallthrough
		default:
			records = append(records, []string{
				v.Id, v.Name, v.ProgramName, strconv.FormatFloat(float64(v.Sequence), 'f', -1, 32), v.FieldToCalc, v.Formula,
			})
		}

		c := csv.NewWriter(file)

		err = c.WriteAll(records)
		if err != nil {
			h.l.Error("error writing csv", zap.Error(err))
		}
		c.Flush()

		if err := c.Error(); err != nil {
			h.l.Error("error flushing CSV writer", zap.String("file", v.Name), zap.Error(err))
		}
	}
	w.Write([]byte("success"))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMCLIToSearch(w http.ResponseWriter, r *http.Request) {
	p := new(MeasureCalcsResponse)

	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding", zap.Error(err))
		return
	}

	filesCreated := make(map[string]*os.File)
	defer func() {
		for _, file := range filesCreated {
			file.Close()
		}
	}()

	for _, v := range p.MeasureCalcRecords {
		fileName := "MCLI To Search"
		h.l.Info("", zap.Any("file Name is: ", fileName))

		file, err := h.CreateCSVFile(h.cfg.JsonDirPath, fileName)
		if err != nil {
			h.l.Error("error creating or opening a csv file", zap.Error(err))
			return
		}
		defer file.Close()

		fileName = fmt.Sprintf("%s/%s.csv", h.cfg.JsonDirPath, fileName)

		d, err := os.ReadFile(fileName)
		if err != nil {
			h.l.Error("error reading file", zap.Error(err))
		}

		var records [][]string

		r := csv.NewReader(strings.NewReader(string(d)))

		records, err = r.ReadAll()
		if err != nil {
			h.l.Error("error reading CSV data", zap.Error(err))
		}

		filesCreated[v.Name] = file
		h.l.Info("size of first records", zap.Any("", len(records)))

		if strings.EqualFold(v.Formula, "Lookup") {
			switch len(records) {
			case 0:
				records = append(records, []string{"Id", "Name", "Program_Name__c", "CLR_CNI_Sequence__c", "CLI_CNR_Field_to_calculate__c", "CLR_CNI_Mesaure_Formula__c"})
				fallthrough
			default:
				records = append(records, []string{
					v.Id, v.Name, v.ProgramName, strconv.FormatFloat(float64(v.Sequence), 'f', -1, 32), v.FieldToCalc, v.Formula,
				})
			}
		}

		h.mcIdMap[v.Id] = v

		c := csv.NewWriter(file)

		err = c.WriteAll(records)
		if err != nil {
			h.l.Error("error writing csv", zap.Error(err))
		}
		c.Flush()

		if err := c.Error(); err != nil {
			h.l.Error("error flushing CSV writer", zap.String("file", v.Name), zap.Error(err))
		}
	}
	w.Write([]byte("success"))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Getmcliquery(w http.ResponseWriter, r *http.Request) {
	p := new(MeasureCalcsResponse)

	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding", zap.Error(err))
		return
	}

	idMap := make(map[string]struct{})
	for _, v := range p.MeasureCalcRecords {
		idMap[v.Id] = struct{}{}
	}

	var ids []string
	for id := range idMap {
		ids = append(ids, fmt.Sprintf("'%s'", id))
	}

	result := strings.Join(ids, ",")
	query := fmt.Sprintf("SELECT Condition__c,Id,Measure_Calculation__c,Measure_Formula__c FROM CLR_CNI_Measure_Calculations_Line_Item__c WHERE Measure_Calculation__c IN (%s)", result)
	encodedQuery := url.QueryEscape(query)
	w.Write([]byte(encodedQuery))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateCSVFile(dir, fileName string) (*os.File, error) {
	fName := fmt.Sprintf("%s.csv", fileName)
	newFilename := filepath.Join(dir, fName)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		h.l.Error("error creating directory", zap.Error(err))
		return nil, err
	}

	file, err := os.OpenFile(newFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (h *Handler) GetMCLI(w http.ResponseWriter, r *http.Request) {
	p := new(MCLIResponse)

	err := FromJSON(p, r.Body)
	if err != nil {
		h.l.Error("error decoding", zap.Error(err))
		return
	}

	filesCreated := make(map[string]*os.File)
	defer func() {
		for _, file := range filesCreated {
			file.Close()
		}
	}()

	for _, v := range p.MCLIRecords {
		fileName := "AllMCLI"
		h.l.Info("", zap.Any("file Name is: ", fileName))

		file, err := h.CreateCSVFile(h.cfg.JsonDirPath, fileName)
		if err != nil {
			h.l.Error("error creating or opening a csv file", zap.Error(err))
			return
		}
		defer file.Close()

		fileName = fmt.Sprintf("%s/%s.csv", h.cfg.JsonDirPath, fileName)

		d, err := os.ReadFile(fileName)
		if err != nil {
			h.l.Error("error reading file", zap.Error(err))
		}

		var records [][]string

		r := csv.NewReader(strings.NewReader(string(d)))

		records, err = r.ReadAll()
		if err != nil {
			h.l.Error("error reading CSV data", zap.Error(err))
		}

		filesCreated["AllMCLI"] = file
		h.l.Info("size of first records", zap.Any("", len(records)))

		switch len(records) {
		case 0:
			records = append(records, []string{"Condition__c", "Id", "Measure_Calculation__c", "Measure_Formula__c",
				"Measure_Calculation__r.CLI_CNR_Field_to_calculate__c", "Measure_Calculation__r.Name", "Measure_Calculation__r.Program_Name__c",
				"Measure_Calculation__r.CLR_CNI_Mesaure_Formula__c", "Measure_Calculation__r.Id",
			})
			fallthrough
		default:
			records = append(records, []string{
				v.Condition, v.Id, v.MeasureCalc, v.Formula, h.mcIdMap[v.MeasureCalc].FieldToCalc, h.mcIdMap[v.MeasureCalc].Name, h.mcIdMap[v.MeasureCalc].ProgramName,
				h.mcIdMap[v.MeasureCalc].Formula, h.mcIdMap[v.MeasureCalc].Id,
			})
		}

		c := csv.NewWriter(file)

		err = c.WriteAll(records)
		if err != nil {
			h.l.Error("error writing csv", zap.Error(err))
		}
		c.Flush()

		if err := c.Error(); err != nil {
			h.l.Error("error flushing CSV writer", zap.String("file", "AllMCLI"), zap.Error(err))
		}
	}
	w.Write([]byte("success"))
	w.WriteHeader(http.StatusOK)
}
