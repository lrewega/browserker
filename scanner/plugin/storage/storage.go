package storage

import (
	"encoding/json"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gitlab.com/browserker/browserk"
)

type Plugin struct {
	service browserk.PluginServicer
}

func New(service browserk.PluginServicer) *Plugin {
	p := &Plugin{service: service}
	service.Register(p)
	return p
}

// Name of the plugin
func (h *Plugin) Name() string {
	return "StoragePlugin"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-P-0003"
}

// Config for this plugin
func (h *Plugin) Config() *browserk.PluginConfig {
	return nil
}

func (h *Plugin) InitContext(bctx *browserk.Context) {

}

// Options for the plugin manager to take into consideration when dispatching
func (h *Plugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenStorage: true,
		ExecutionType: browserk.ExecAlways,
	}
}

// Ready to attack
func (h *Plugin) Ready(injector browserk.Injector) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
	h.checkJWTTokens(evt)
}

// Checks storage updates for JWT tokens being added to local/sessionStorage
func (h *Plugin) checkJWTTokens(evt *browserk.PluginEvent) {
	tokenData := ""
	jwt.Parse(evt.EventData.Storage.NewValue, func(token *jwt.Token) (interface{}, error) {
		if _, valid := token.Header["alg"]; valid {
			header, _ := json.Marshal(token.Header)
			claims, _ := json.Marshal(token.Claims)
			tokenData = string(token.Raw) + "\nheader: " + string(header) + "\nclaims: " + string(claims)
		}
		return nil, nil
	})

	if tokenData == "" {
		return
	}

	storageType := "sessionStorage"
	remediation := "Be extremely careful that no XSS vulnerabilities exist as it will be possible to extract the JWT tokens directly from the sessionStorage"
	severity := "LOW"
	checkID := 1
	cwe := 200

	if evt.EventData.Storage.IsLocalStorage {
		storageType = "localStorage"
		remediation = "JWT tokens should be stored in sessionStorage, not localStorage as they persist in the browser until they are cleared."
		severity = "MEDIUM"
		cwe = 922
		checkID = 2
	}

	report := &browserk.Report{
		Plugin:      h.Name(),
		CheckID:     checkID,
		CWE:         cwe,
		Description: "JWT Token found in " + storageType,
		Remediation: remediation,
		Severity:    severity,
		URL:         evt.URL,
		Nav:         evt.Nav,
		Result:      nil,
		Evidence: &browserk.Evidence{
			ID:     nil,
			String: tokenData,
		},
		Reported: time.Now(),
	}
	report.Hash()
	evt.BCtx.Reporter.Add(report)
}
