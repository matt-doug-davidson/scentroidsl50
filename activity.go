package scentroidsl50

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/matt-doug-davidson/timestamps"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

type Activity struct {
	settings       *Settings // Defind in metadata.go in this package
	EnvironmentURL string
	PollutantURL   string
	SerialNumber   string
	Mappings       map[string]map[string]interface{}
	Entity         string
}

const (
	environmentalEndpoint string = "/do/api/v1.rpi_get_pollutant"
	pollutantEndpoint     string = "/do/api/v1.rpi_get_samples"
	valueIndex            int    = 0
	msgTimeIndex          int    = 1
	measurementIDIndex    int    = 2
	sensorIndex           int    = 3
)

// Metadata returns the activity's metadata
// Common function
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// The init function is executed after the package is imported. This function
// runs before any other in the package.
func init() {
	//_ = activity.Register(&Activity{})
	_ = activity.Register(&Activity{}, New)
}

// Used when the init function is called. The settings, Input and Output
// structures are optional depends application. These structures are
// defined in the metadata.go file in this package.
var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// New Looks to be used when the Activity structure contains fields that need to be
// configured using the InitContext information.
// New does this
func New(ctx activity.InitContext) (activity.Activity, error) {
	logger := ctx.Logger()
	logger.Info("scentroidsl50:New enter")
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		logger.Error("Failed to convert settings")
		return nil, err
	}
	commonURL := "http://" + s.Host + ":" + s.Port

	// Declared an empty map interface
	var result map[string]interface{}
	json.Unmarshal([]byte(s.Mappings), &result)

	mm := map[string]map[string]interface{}{}
	for _, mapper := range result {
		mapper1 := mapper.(map[string]interface{})
		for sensor, sensorInfo := range mapper1 {
			si := sensorInfo.(map[string]interface{})
			mm[sensor] = make(map[string]interface{})
			se := map[string]interface{}{}
			f, foundF := si["field"]
			if !foundF {
				continue
			}
			se["field"] = f
			m, foundM := si["multiplier"]
			if foundM {
				se["multiplier"] = m
			} else {
				se["multiplier"] = 1.0
			}
			mm[sensor] = se
		}
	}

	act := &Activity{
		EnvironmentURL: commonURL + environmentalEndpoint,
		PollutantURL:   commonURL + pollutantEndpoint,
		SerialNumber:   s.SerialNumber,
		Mappings:       mm,
		Entity:         s.Entity,
	}

	logger.Info("scentroidsl50:New exit")
	return act, nil
}

// Eval evaluates the activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	logger.Info("scentroidsl50:Eval enter")

	var timestamp string
	values := []map[string]interface{}{}
	values, timestamp = a.getPollutantData(values)
	values, _ = a.getEnvironmentalData(values)

	// Convert time from message to UTCZ time-string
	msts := timestamps.MillisecondTimestamp{}
	datetime := msts.ConvertUTCZ(timestamp)

	message := map[string]interface{}{}
	message["values"] = values
	message["datetime"] = datetime

	output := map[string]interface{}{}
	output["data"] = message
	output["entity"] = a.Entity

	err = ctx.SetOutput("connectorMsg", output)
	if err != nil {
		logger.Error("Failed to set output oject ", err.Error())
		return false, err
	}

	logger.Info("scentroidsl50:Eval exit")
	return true, nil
}

func (a *Activity) getData(url string) ([]byte, error) {

	client := http.Client{Timeout: 5 * time.Second}

	reqGet, _ := http.NewRequest("GET", url, nil)

	// Build the query
	q := reqGet.URL.Query()
	q.Add("sn", a.SerialNumber)
	q.Add("latest", "true")
	reqGet.URL.RawQuery = q.Encode()

	resGet, err := client.Do(reqGet)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resGet.Body)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return nil, err
	}

	return body, nil
}

func (a *Activity) getEnvironmentalData(values []map[string]interface{}) ([]map[string]interface{}, string) {
	body, err := a.getData(a.EnvironmentURL)
	if err != nil {
		return nil, ""
	}

	r := map[string]interface{}{}
	json.Unmarshal(body, &r)
	rl := r["items"].([]interface{})
	var msgTime string
	// The environmental measurements are wrapped in an extra slice that
	// is compensated for here.
	for _, v := range rl {
		value := map[string]interface{}{}
		vv := v.([]interface{})
		amount := vv[valueIndex].(float64)
		sensor := vv[sensorIndex].(string)
		msgTime = vv[msgTimeIndex].(string)

		// Go says this is a float so we have to then convert to int
		msrmenttID := vv[measurementIDIndex].(float64)
		measurementID := int(msrmenttID)
		if measurementID == 4 || measurementID == 5 {
			sensor += "(internal)"
		}
		if measurementID == 6 || measurementID == 7 {
			sensor += "(external)"
		}

		value["amount"] = amount * a.Mappings[sensor]["multiplier"].(float64)
		value["field"] = a.Mappings[sensor]["field"]
		values = append(values, value)
	}
	return values, msgTime
}

func (a *Activity) getPollutantData(values []map[string]interface{}) ([]map[string]interface{}, string) {
	body, err := a.getData(a.PollutantURL)
	if err != nil {
		return nil, ""
	}

	r := map[string]interface{}{}
	json.Unmarshal(body, &r)
	rl := r["items"].([]interface{})
	msgTime := ""
	for _, v := range rl {
		vu := v.([]interface{})
		for _, viv := range vu {
			value := map[string]interface{}{}
			vv := viv.([]interface{})
			amount := vv[valueIndex].(float64)
			sensor := vv[sensorIndex].(string)
			// Convert sensor to field
			msgTime = vv[msgTimeIndex].(string)
			// Measurement ID not used so it was removed

			value["amount"] = amount * a.Mappings[sensor]["multiplier"].(float64)
			value["field"] = a.Mappings[sensor]["field"]
			values = append(values, value)
		}
	}
	return values, msgTime
}
