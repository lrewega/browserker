package browser

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wirepair/gcd/v2/gcdapi"
	"gitlab.com/browserker/scanner/browser/keymap"
)

// ErrIncorrectElementType an attempt was made to execute a method for which the element type is incorrect
type ErrIncorrectElementType struct {
	NodeName     string
	ExpectedName string
}

func (e *ErrIncorrectElementType) Error() string {
	return "incorrect element type, expected " + e.ExpectedName + " but this type is " + e.NodeName
}

// ErrInvalidElement the element has been removed from the DOM
type ErrInvalidElement struct {
}

func (e *ErrInvalidElement) Error() string {
	return "this element has been invalidated"
}

// ErrElementHasNoChildren The element has no children
type ErrElementHasNoChildren struct {
}

func (e *ErrElementHasNoChildren) Error() string {
	return "this element has no child elements"
}

// ErrElementNotReady when we have an element that has not been populated
// with data yet.
type ErrElementNotReady struct {
}

func (e *ErrElementNotReady) Error() string {
	return "this element is not ready"
}

// ErrInvalidDimensions when the dimensions of an element are incorrect to calculate the centroid
type ErrInvalidDimensions struct {
	Message string
}

func (e *ErrInvalidDimensions) Error() string {
	return "invalid dimensions " + e.Message
}

// Element is an abstraction over a DOM element, it can be in three modes
// NotReady - it's data has not been returned to us by the debugger yet.
// Ready - the debugger has given us the DOMNode reference.
// Invalidated - The Element has been destroyed.
// Certain actions require that the Element be populated (getting nodename/type)
// If you need this information, wait for IsReady() to return true
type Element struct {
	lock           *sync.RWMutex     // for protecting read/write access to this Element
	attributes     map[string]string // dom attributes
	nodeName       string            // the DOM tag name
	characterData  string            // the character data (if any, #text only)
	childNodeCount int               // the number of children this element has
	nodeType       int               // the DOM nodeType
	depth          int               // depth of the node in the document
	tab            *Tab              // reference to the containing tab
	node           *gcdapi.DOMNode   // the dom node, taken from the document
	readyGate      chan struct{}     // gate to close upon recieving all information from the debugger service
	ID             int               // nodeId in chrome
	ready          bool              // has this elements data been populated by setChildNodes or GetDocument?
	invalidated    bool              // has this node been invalidated (removed?)
}

func newElement(tab *Tab, nodeID, depth int) *Element {
	e := &Element{}
	e.tab = tab
	e.attributes = make(map[string]string)
	e.readyGate = make(chan struct{})
	e.lock = &sync.RWMutex{}
	e.ID = nodeID
	e.depth = depth
	return e
}

func newReadyElement(tab *Tab, node *gcdapi.DOMNode, depth int) *Element {
	e := &Element{}
	e.tab = tab
	e.attributes = make(map[string]string)
	e.readyGate = make(chan struct{})
	e.nodeName = strings.ToLower(node.NodeName)
	e.lock = &sync.RWMutex{}
	e.populateElement(node, depth)
	return e
}

// populate the Element with node data.
func (e *Element) populateElement(node *gcdapi.DOMNode, depth int) {
	e.lock.Lock()
	e.node = node
	e.ID = node.NodeId
	e.depth = depth
	e.nodeType = node.NodeType
	e.nodeName = strings.ToLower(node.NodeName)
	e.childNodeCount = node.ChildNodeCount
	if node.NodeType == int(NodeText) {
		e.characterData = node.NodeValue
	}

	e.lock.Unlock()

	for i := 0; i < len(node.Attributes); i += 2 {
		e.updateAttribute(node.Attributes[i], node.Attributes[i+1])
	}

	// close it
	if !e.ready {
		close(e.readyGate)
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.ready = true
}

// IsReady has the Chrome Debugger notified us of this Elements data yet?
func (e *Element) IsReady() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return (e.ready && !e.invalidated)
}

// IsReadyInvalid has Chrome notified us, but the element is invalid?
func (e *Element) IsReadyInvalid() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return (e.ready && e.invalidated)
}

// IsInvalid has the debugger invalidated (removed) the element from the DOM?
func (e *Element) IsInvalid() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.invalidated
}

// The element has become invalid.
func (e *Element) setInvalidated(invalid bool) {
	e.lock.Lock()
	e.invalidated = invalid
	e.lock.Unlock()
}

// Depth of this node as relative to the <html> doc
func (e *Element) Depth() int {
	return e.depth
}

// WaitForReady If we are ready, just return, if we are not, wait for the readyGate
// to be closed or for the timeout timer to fired.
func (e *Element) WaitForReady() error {
	e.lock.RLock()
	ready := e.ready
	e.lock.RUnlock()

	if ready {
		return nil
	}

	timeout := time.NewTimer(e.tab.elementTimeout)
	defer timeout.Stop()

	select {
	case <-e.readyGate:
		return nil
	case <-e.tab.ctx.Ctx.Done():
		return &ErrElementNotReady{}
	case <-timeout.C:
		return &ErrElementNotReady{}
	case <-e.tab.exitCh:
		return &ErrElementNotReady{}
	}
}

// GetSource returns the outer html of the element.
func (e *Element) GetSource() (string, error) {
	e.lock.RLock()
	id := e.ID
	e.lock.RUnlock()

	if e.invalidated {
		return "", &ErrInvalidElement{}
	}

	outerParams := &gcdapi.DOMGetOuterHTMLParams{NodeId: id}
	return e.tab.t.DOM.GetOuterHTMLWithParams(e.tab.ctx.Ctx, outerParams)
}

// IsDocument Is this Element a #document?
func (e *Element) IsDocument() (bool, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready {
		return false, &ErrElementNotReady{}
	}
	return (e.nodeType == int(NodeDocument)), nil
}

// FrameID If this is a #document, returns the underlying chrome frameId.
func (e *Element) FrameID() (string, error) {
	if !e.IsReady() {
		return "", &ErrElementNotReady{}
	}

	isDoc, err := e.IsDocument()
	if err != nil {
		return "", err
	}

	if !isDoc {
		return "", nil
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.node.FrameId, nil
}

// GetFrameDocumentNodeID if this element is a frame or iframe, return the ContentDocument node id
func (e *Element) GetFrameDocumentNodeID() (int, error) {
	if !e.IsReady() {
		return -1, &ErrElementNotReady{}
	}
	e.lock.RLock()
	defer e.lock.RUnlock()

	if e.node != nil && e.node.ContentDocument != nil {
		return e.node.ContentDocument.NodeId, nil
	}
	return -1, &ErrIncorrectElementType{ExpectedName: "(i)frame", NodeName: e.nodeName}
}

// NodeID returns the underlying chrome debugger node id of this Element
func (e *Element) NodeID() int {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.ID
}

// GetEventListeners returns event listeners for the element, both static and dynamically bound.
func (e *Element) GetEventListeners() ([]*gcdapi.DOMDebuggerEventListener, error) {
	e.lock.RLock()
	id := e.ID
	e.lock.RUnlock()

	params := &gcdapi.DOMResolveNodeParams{
		NodeId: id,
	}

	rro, err := e.tab.t.DOM.ResolveNodeWithParams(e.tab.ctx.Ctx, params)
	if err != nil {
		return nil, err
	}
	eventListeners, err := e.tab.t.DOMDebugger.GetEventListeners(e.tab.ctx.Ctx, rro.ObjectId, 1, false)
	if err != nil {
		return nil, err
	}
	return eventListeners, nil
}

// GetDebuggerDOMNode returns the underlying DOMNode for this element. Note this is potentially
// unsafe to access as we give up the ability to lock.
func (e *Element) GetDebuggerDOMNode() (*gcdapi.DOMNode, error) {
	if !e.IsReady() {
		return nil, &ErrElementNotReady{}
	}

	e.lock.RLock()
	defer e.lock.RUnlock()
	if e.invalidated {
		return nil, &ErrInvalidElement{}
	}
	return e.node, nil
}

// updates the attribute name/value pair
func (e *Element) updateAttribute(name, value string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.attributes[name] = value
}

// removes the attribute from our attributes list.
func (e *Element) removeAttribute(name string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	delete(e.attributes, name)
}

// updates character data
func (e *Element) updateCharacterData(newValue string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.characterData = newValue
}

// updates child node counts.
func (e *Element) updateChildNodeCount(newValue int) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.childNodeCount = newValue
}

// adds the child to our DOMNode.
func (e *Element) addChild(child *gcdapi.DOMNode) {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.node == nil {
		return
	}

	if e.node.Children == nil {
		e.node.Children = make([]*gcdapi.DOMNode, 0)
	}
	e.node.Children = append(e.node.Children, child)
	e.childNodeCount++
}

// adds the children to our DOMNode
func (e *Element) addChildren(childNodes []*gcdapi.DOMNode) {
	for _, child := range childNodes {
		if child != nil {
			e.addChild(child)
		}
	}
}

// removes the child from our DOMNode
func (e *Element) removeChild(removedNodeID int) {
	var idx int
	var child *gcdapi.DOMNode

	e.lock.Lock()
	defer e.lock.Unlock()

	if e.node == nil || e.node.Children == nil {
		return
	}

	for idx, child = range e.node.Children {
		if child != nil && child.NodeId == removedNodeID {
			e.node.Children = append(e.node.Children[:idx], e.node.Children[idx+1:]...)
			e.childNodeCount = e.childNodeCount - 1
			break
		}
	}
}

// GetChildNodeIds returns nil if node is not ready
func (e *Element) GetChildNodeIds() ([]int, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready {
		return nil, &ErrElementNotReady{}
	}

	if e.node == nil || e.node.Children == nil {
		return nil, &ErrElementHasNoChildren{}
	}

	ids := make([]int, len(e.node.Children))
	for k, child := range e.node.Children {
		ids[k] = child.NodeId
	}
	return ids, nil
}

// GetTagName returns the tag name (input, div etc) if the element is in a ready state.
func (e *Element) GetTagName() (string, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready && e.nodeName == "" {
		return "", &ErrElementNotReady{}
	}
	return e.nodeName, nil
}

// GetNodeType returns the node type if the element is in a ready state.
func (e *Element) GetNodeType() (int, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready {
		return -1, &ErrElementNotReady{}
	}
	return e.nodeType, nil
}

// GetCharacterData returns the character data if the element is in a ready state.
func (e *Element) GetCharacterData() (string, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready {
		return "", &ErrElementNotReady{}
	}

	return e.characterData, nil
}

// GetInnerText of this element
func (e *Element) GetInnerText() string {
	return e.tab.GetChildrensCharacterData(e)
}

// IsEnabled returns true if the node is enabled, only makes sense for form controls.
// Element must be in a ready state.
func (e *Element) IsEnabled() (bool, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready {
		return false, &ErrElementNotReady{}
	}
	disabled, ok := e.attributes["disabled"]
	// if the attribute doesn't exist, it's enabled.
	if !ok {
		return true, nil
	}
	if disabled == "true" || disabled == "" {
		return false, nil
	}
	return true, nil
}

// IsSelected simulate WebDrivers checked propertyname check
func (e *Element) IsSelected() (bool, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.ready {
		return false, &ErrElementNotReady{}
	}

	checked, ok := e.attributes["checked"]
	if ok == true && checked != "false" {
		return true, nil
	}
	return false, nil
}

// GetCSSInlineStyleText returns the CSS Style Text of the element, returns the inline style first
// and the attribute style second, or error.
func (e *Element) GetCSSInlineStyleText() (string, string, error) {
	e.lock.RLock()
	inline, attribute, err := e.tab.t.CSS.GetInlineStylesForNode(e.tab.ctx.Ctx, e.ID)
	e.lock.RUnlock()

	if err != nil {
		return "", "", err
	}
	return inline.CssText, attribute.CssText, nil
}

// GetComputedCSSStyle returns all of the computed css styles in form of name value map.
func (e *Element) GetComputedCSSStyle() (map[string]string, error) {
	e.lock.RLock()
	styles, err := e.tab.t.CSS.GetComputedStyleForNode(e.tab.ctx.Ctx, e.ID)
	e.lock.RUnlock()

	if err != nil {
		return nil, err
	}
	styleMap := make(map[string]string, len(styles))
	for _, style := range styles {
		styleMap[style.Name] = style.Value
	}
	return styleMap, nil
}

// GetAttributes of the node returning a map of name,value pairs.
func (e *Element) GetAttributes() (map[string]string, error) {
	e.lock.RLock()
	attr, err := e.tab.t.DOM.GetAttributes(e.tab.ctx.Ctx, e.ID)
	e.lock.RUnlock()

	if err != nil {
		return nil, err
	}
	for i := 0; i < len(attr); i += 2 {
		e.updateAttribute(attr[i], attr[i+1])
	}

	return e.attributes, nil
}

// GetAttribute a single attribute by name, returns empty string if it does not exist
// or is empty.
func (e *Element) GetAttribute(name string) string {
	attr, err := e.GetAttributes()
	if err != nil {
		return ""
	}
	return attr[name]
}

// HasAttribute similar to above, but works for boolean properties (checked, async etc)
// Returns true if the attribute is set in our known list of attributes
// for this element.
func (e *Element) HasAttribute(name string) bool {
	attr, err := e.GetAttributes()
	if err != nil {
		return false
	}
	_, exists := attr[name]
	return exists
}

// SetAttributeValue sets an element's attribute with name to value.
func (e *Element) SetAttributeValue(name, value string) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	_, err := e.tab.t.DOM.SetAttributeValue(e.tab.ctx.Ctx, e.ID, name, value)
	if err != nil {
		return err
	}

	e.attributes[name] = value

	return nil
}

// Clear works like WebDriver's clear(), simply sets the attribute value for input
// or clears the value for textarea. This element must be ready so we can
// properly read the nodeName value.
func (e *Element) Clear() error {
	e.lock.RLock()
	defer e.lock.RUnlock()

	var err error

	if !e.ready {
		return &ErrElementNotReady{}
	}

	if e.nodeName == "textarea" {
		_, err = e.tab.t.DOM.SetNodeValue(e.tab.ctx.Ctx, e.ID, "")
	} else if e.nodeName == "input" {
		_, err = e.tab.t.DOM.SetAttributeValue(e.tab.ctx.Ctx, e.ID, "value", "")
	} else {
		err = &ErrIncorrectElementType{ExpectedName: "textarea or input", NodeName: e.nodeName}
	}
	return err
}

// Click the center of the element.
func (e *Element) Click() error {
	x, y, err := e.getCenter()
	if err != nil {
		return err
	}

	// click the centroid of the element.
	return e.tab.Click(float64(x), float64(y))
}

// ClickParent node
func (e *Element) ClickParent() error {
	parentID := e.GetParentNodeID()

	params := &gcdapi.DOMGetBoxModelParams{
		NodeId: parentID,
	}
	box, err := e.tab.t.DOM.GetBoxModelWithParams(e.tab.ctx.Ctx, params)

	if err != nil {
		return err
	}
	points := box.Content
	x, y, err := centroid(points)
	return e.tab.Click(float64(x), float64(y))
}

// GetParentNodeID of this element
func (e *Element) GetParentNodeID() int {
	e.lock.RLock()
	parentID := e.node.ParentId
	e.lock.RUnlock()
	return parentID
}

// DoubleClick the center of the element.
func (e *Element) DoubleClick() error {
	x, y, err := e.getCenter()
	if err != nil {
		return err
	}

	return e.tab.DoubleClick(float64(x), float64(y))
}

// Focus on the element.
func (e *Element) Focus() error {
	e.lock.RLock()
	defer e.lock.RUnlock()

	params := &gcdapi.DOMFocusParams{
		NodeId: e.ID,
	}
	_, err := e.tab.t.DOM.FocusWithParams(e.tab.ctx.Ctx, params)
	return err
}

// ScrollTo the element if needed
func (e *Element) ScrollTo() error {
	e.lock.RLock()
	defer e.lock.RUnlock()

	params := &gcdapi.DOMScrollIntoViewIfNeededParams{
		NodeId: e.ID,
	}
	_, err := e.tab.t.DOM.ScrollIntoViewIfNeededWithParams(e.tab.ctx.Ctx, params)
	return err
}

// MouseOver the center of the element.
func (e *Element) MouseOver() error {
	x, y, err := e.getCenter()
	if err != nil {
		return err
	}

	return e.tab.MoveMouse(float64(x), float64(y))
}

// Dimensions returns the dimensions of the element.
func (e *Element) Dimensions() ([]float64, error) {
	var points []float64
	e.lock.RLock()

	params := &gcdapi.DOMGetBoxModelParams{
		NodeId: e.ID,
	}
	box, err := e.tab.t.DOM.GetBoxModelWithParams(e.tab.ctx.Ctx, params)

	e.lock.RUnlock()

	if err != nil {
		return nil, err
	}
	points = box.Content
	return points, nil
}

// gets the center of the element
func (e *Element) getCenter() (int, int, error) {
	points, err := e.Dimensions()
	if err != nil {
		return 0, 0, err
	}

	x, y, err := centroid(points)
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

// SendKeys sends each individual character after focusing (clicking) on the element.
// Extremely basic, doesn't take into account most/all system keys except enter, tab or backspace.
func (e *Element) SendKeys(text string) error {
	e.Focus()
	err := e.Click()
	if err != nil {
		return err
	}
	return e.tab.SendKeys(text)
}

func (e *Element) SendRawKeys(keys string) error {
	e.Focus()
	err := e.Click()
	if err != nil {
		return err
	}
	for _, c := range keys {
		toSend := keymap.KeyEncode(c)
		for _, key := range toSend {
			_, err = e.tab.t.Input.DispatchKeyEventWithParams(e.tab.ctx.Ctx, key)
			time.Sleep(time.Millisecond * 70) // small delay inbetween key presses
		}
	}
	return err
}

// String gnarly output mode activated
func (e *Element) String() string {
	e.lock.RLock()
	defer e.lock.RUnlock()
	output := fmt.Sprintf("NodeId: %d Invalid: %t Ready: %t", e.ID, e.invalidated, e.ready)
	if !e.ready {
		return output
	}
	attrs := ""
	for key, value := range e.attributes {
		attrs = attrs + "\t" + key + "=" + value + "\n"
	}
	output = fmt.Sprintf("%s NodeType: %d TagName: %s characterData: %s childNodeCount: %d attributes (%d): \n%s", output, e.nodeType, e.nodeName, e.characterData, e.childNodeCount, len(e.attributes), attrs)
	if e.nodeType == int(NodeDocument) {
		output = fmt.Sprintf("%s FrameId: %s documentURL: %s\n", output, e.node.FrameId, e.node.DocumentURL)
	}
	//output = fmt.Sprintf("%s %#v", output, e.node)
	return output
}

// finds the centroid of an arbitrary number of points.
// Assumes points[i] = x, points[i+1] = y.
func centroid(points []float64) (int, int, error) {
	pointLen := len(points)
	if pointLen%2 != 0 {
		return -1, -1, &ErrInvalidDimensions{"number of points are not divisible by two"}
	}
	x := 0
	y := 0
	for i := 0; i < pointLen; i = i + 2 {
		x += int(points[i])
		y += int(points[i+1])
	}
	return x / (pointLen / 2), y / (pointLen / 2), nil
}
