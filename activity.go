package scentroidsl50

import (
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

type Activity struct {
	settings       *Settings // Defind in metadata.go in this package
	EnvironmentURL string
	PollutantURL   string
	SerialNumber   string
	Mappings       map[string]map[string]interface{}
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
	for i, v := range result {
		//a.Mappings[i] = make(map[string]interface{})
		fmt.Println("i[", i, "]=", v)
		fmt.Printf("i (type) %T\n", i)
		fmt.Printf("v (type) %T\n", v)

		// mm[i] = make(map[string]interface{})
		// fmt.Println(i, v)
		// se := map[string]interface{}{}
		// //f, found := result[i]["field"]
		// if !found {
		// 	continue
		// }
		// se["field"] = f
		// m, mFound := s.Mappings[i]["multiplier"]
		// if !mFound {
		// 	se["multiplier"] = 1.0
		// } else {
		// 	se["multiplier"] = m
		// }

		// fmt.Println(se)
		// fmt.Println(m)
		// mm[i] = se
	}
	fmt.Println(mm)

	act := &Activity{
		EnvironmentURL: commonURL + environmentalEndpoint,
		PollutantURL:   commonURL + pollutantEndpoint,
		SerialNumber:   s.SerialNumber,
		Mappings:       mm,
	}

	logger.Info("scentroidsl50:New exit")
	return act, nil
}

// Eval evaluates the activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	logger.Info("scentroidsl50:Eval enter")
	logger.Info("scentroidsl50:Eval exit")
	return true, nil
}
