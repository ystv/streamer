package tx

type FunctionTX string

var (
	TransmissionOn FunctionTX = "transmission_on"
	AllOff         FunctionTX = "rehearsal_transmission_off"
	RehearsalOn    FunctionTX = "rehearsal_on"
)

func (f FunctionTX) String() string {
	return string(f)
}
