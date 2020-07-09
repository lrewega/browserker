package browserk

// ActionType defines the action type for a browser action
type ActionType int8

// revive:disable:var-naming
const (
	ActLoadURL ActionType = iota + 1
	ActExecuteJS
	ActLeftClick
	ActLeftClickDown
	ActLeftClickUp
	ActRightClick
	ActDoubleClick
	ActScroll
	ActSendKeys
	ActKeyUp
	ActKeyDown
	ActHover
	ActFocus
	ActBlur
	ActMouseOverAndOut
	ActMouseWheel

	ActFillForm
	ActWait

	// ActionTypes that occured automatically
	ActRedirect
	ActSubRequest
)

// ActionTypeMap to display the action
var ActionTypeMap = map[ActionType]string{
	ActLoadURL:         "ActLoadURL",
	ActExecuteJS:       "ActExecuteJS",
	ActLeftClick:       "ActLeftClick",
	ActLeftClickDown:   "ActLeftClickDown",
	ActLeftClickUp:     "ActLeftClickUp",
	ActRightClick:      "ActRightClick",
	ActDoubleClick:     "ActDoubleClick",
	ActScroll:          "ActScroll",
	ActSendKeys:        "ActSendKeys",
	ActKeyUp:           "ActKeyUp",
	ActKeyDown:         "ActKeyDown",
	ActHover:           "ActHover",
	ActFocus:           "ActFocus",
	ActBlur:            "ActBlur",
	ActWait:            "ActWait",
	ActMouseOverAndOut: "ActMouseOverAndOut",
	ActMouseWheel:      "ActMouseWheel",
	ActFillForm:        "ActFillForm",

	// ActionTypes that occured automatically
	ActRedirect:   "ActRedirect",
	ActSubRequest: "ActSubRequest",
}

// Action runs a browser action, may or may not create a result
type Action struct {
	browser Browser
	Type    ActionType       `graph:"type"`
	Input   []byte           `graph:"input"`
	Element *HTMLElement     `graph:"element"`
	Form    *HTMLFormElement `graph:"form"`
	Result  []byte           `graph:"result"`
}

func NewLoadURLAction(url string) *Action {
	return &Action{
		Type:  ActLoadURL,
		Input: []byte(url),
	}
}

func (a *Action) String() string {
	ret := ""
	switch a.Type {
	case ActLoadURL:
		ret += "[" + string(a.Input) + "]"
	case ActLeftClick:
		ret += "[" + HTMLTypeToStrMap[a.Element.Type] + " "
		for k, v := range a.Element.Attributes {
			if len(v) > 25 {
				v = v[0:24] + "..."
			}
			ret += k + "=" + v
		}
		ret += "]"
	case ActFillForm:
		ret += "[ FORM "
		for k, v := range a.Form.Attributes {
			ret += k + "=" + v
		}
		ret += "]"
	}
	return ret
}
