package neuralnet

import (
	"encoding/json"

	"github.com/unixpickle/autofunc"
)

// RescaleLayer is a Layer which adds a bias to its
// input and scales the translated input.
// It is useful for ensuring that input samples have
// a mean of 0 and a standard deviation of 1.
type RescaleLayer struct {
	Bias  float64
	Scale float64
}

func DeserializeRescaleLayer(d []byte) (*RescaleLayer, error) {
	var res RescaleLayer
	if err := json.Unmarshal(d, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *RescaleLayer) Apply(in autofunc.Result) autofunc.Result {
	return autofunc.Scale(autofunc.AddScaler(in, r.Bias), r.Scale)
}

func (r *RescaleLayer) ApplyR(v autofunc.RVector, in autofunc.RResult) autofunc.RResult {
	return autofunc.ScaleR(autofunc.AddScalerR(in, r.Bias), r.Scale)
}

func (r *RescaleLayer) Serialize() ([]byte, error) {
	return json.Marshal(r)
}

func (r *RescaleLayer) SerializerType() string {
	return serializerTypeRescaleLayer
}
