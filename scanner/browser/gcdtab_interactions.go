package browser

import "github.com/wirepair/gcd/v2/gcdapi"

// Click the x, y coords one time
func (t *Tab) Click(x, y float64) error {
	return t.click(x, y, 1)
}

func (t *Tab) click(x, y float64, clickCount int) error {
	// "mousePressed", "mouseReleased", "mouseMoved"
	// enum": ["none", "left", "mIDdle", "right"]

	mousePressedParams := &gcdapi.InputDispatchMouseEventParams{TheType: "mousePressed",
		X:          x,
		Y:          y,
		Button:     "left",
		ClickCount: clickCount,
	}

	if _, err := t.t.Input.DispatchMouseEventWithParams(t.ctx.Ctx, mousePressedParams); err != nil {
		return err
	}

	mouseReleasedParams := &gcdapi.InputDispatchMouseEventParams{TheType: "mouseReleased",
		X:          x,
		Y:          y,
		Button:     "left",
		ClickCount: clickCount,
	}

	if _, err := t.t.Input.DispatchMouseEventWithParams(t.ctx.Ctx, mouseReleasedParams); err != nil {
		return err
	}
	return nil
}

// DoubleClick issues a double click on the x, y coords provided.
func (t *Tab) DoubleClick(x, y float64) error {
	return t.click(x, y, 2)
}

// MoveMouse to the x, y coords provided.
func (t *Tab) MoveMouse(x, y float64) error {
	mouseMovedParams := &gcdapi.InputDispatchMouseEventParams{TheType: "mouseMoved",
		X: x,
		Y: y,
	}

	_, err := t.t.Input.DispatchMouseEventWithParams(t.ctx.Ctx, mouseMovedParams)
	return err
}

// SendKeys to whatever is focused, best called from Element.SendKeys which will
// try to focus on the element first. Use \n for Enter, \b for backspace or \t for Tab.
func (t *Tab) SendKeys(text string) error {
	inputParams := &gcdapi.InputDispatchKeyEventParams{TheType: "char"}

	// loop over input, looking for system keys and handling them
	for _, inputchar := range text {
		input := string(inputchar)

		// check system keys
		switch input {
		case "\r", "\n", "\t", "\b":
			if err := t.pressSystemKey(input); err != nil {
				return err
			}
			continue
		}
		inputParams.Text = input
		_, err := t.t.Input.DispatchKeyEventWithParams(t.ctx.Ctx, inputParams)
		if err != nil {
			return err
		}
	}
	return nil
}

// Super ghetto, i know.
func (t *Tab) pressSystemKey(systemKey string) error {
	inputParams := &gcdapi.InputDispatchKeyEventParams{TheType: "rawKeyDown"}

	switch systemKey {
	case "\b":
		inputParams.UnmodifiedText = "\b"
		inputParams.Text = "\b"
		inputParams.WindowsVirtualKeyCode = 8
		inputParams.NativeVirtualKeyCode = 8
	case "\t":
		inputParams.UnmodifiedText = "\t"
		inputParams.Text = "\t"
		inputParams.WindowsVirtualKeyCode = 9
		inputParams.NativeVirtualKeyCode = 9
	case "\r", "\n":
		inputParams.UnmodifiedText = "\r"
		inputParams.Text = "\r"
		inputParams.WindowsVirtualKeyCode = 13
		inputParams.NativeVirtualKeyCode = 13
	}

	if _, err := t.t.Input.DispatchKeyEventWithParams(t.ctx.Ctx, inputParams); err != nil {
		return err
	}

	inputParams.TheType = "char"
	if _, err := t.t.Input.DispatchKeyEventWithParams(t.ctx.Ctx, inputParams); err != nil {
		return err
	}

	inputParams.TheType = "keyUp"
	if _, err := t.t.Input.DispatchKeyEventWithParams(t.ctx.Ctx, inputParams); err != nil {
		return err
	}
	return nil
}
