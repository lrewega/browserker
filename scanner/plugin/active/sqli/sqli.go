package sqli

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"gitlab.com/browserker/browserk"
)

type SQLIAttack struct {
	DBTech      browserk.TechType
	Prefix      string
	Suffix      string
	Attack      string
	Description string
	IsTiming    bool
}

type Plugin struct {
	service      browserk.PluginServicer
	sleepTimeSec time.Duration
	attacks      []*SQLIAttack
}

func New(service browserk.PluginServicer) *Plugin {
	p := &Plugin{service: service, attacks: make([]*SQLIAttack, 0), sleepTimeSec: 15}
	p.initAttacks()
	service.Register(p)
	return p
}

// Name of the plugin
func (p *Plugin) Name() string {
	return "SQLInjectionPlugin"
}

// ID unique to browserker
func (p *Plugin) ID() string {
	return "BR-A-0003"
}

// Config for this plugin
func (p *Plugin) Config() *browserk.PluginConfig {
	return nil
}

func (p *Plugin) InitContext(bctx *browserk.Context) {

}

// Options for the plugin manager to take into consideration when dispatching
func (p *Plugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		WriteRequests: true,
		ExecutionType: browserk.ExecAlways,
		Injections:    []browserk.InjectionLocation{browserk.InjectNameValue},
	}
}

// Ready to attack
func (p *Plugin) Ready(injector browserk.Injector) (bool, error) {
	// msg := injector.Message() // get original req/resp
	expr := injector.InjectionExpr()
	for _, attack := range p.attacks {
		if attack.IsTiming {
			found, err := p.doTimingAttack(injector, attack)
			if err != nil {
				injector.BCtx().Log.Warn().Err(err).Msg("attack failed")
			}
			if found {
				return true, nil
			}
		}

		expr.Inject(attack.Prefix+attack.Attack+attack.Suffix, browserk.InjectValue)

		ctx, cancel := context.WithTimeout(injector.BCtx().Ctx, time.Second*15)
		defer cancel()
		m, err := injector.Send(ctx, false)
		if err != nil {
			injector.BCtx().Log.Error().Err(err).Msg("failed to inject")
			return false, nil
		}
		injector.BCtx().Log.Info().Msg("attacked!")
		body := m.Response.Body
		if m.Response.BodyEncoded {
			b, err := base64.StdEncoding.DecodeString(m.Response.Body)
			if err != nil {
				injector.BCtx().Log.Error().Err(err).Msg("failed to decode body response")
				return true, nil
			} else {
				body = string(b)
			}
		}

		if strings.Contains(body, "root:") {
			injector.BCtx().Reporter.Add(&browserk.Report{
				CheckID:     "1",
				CWE:         78,
				URL:         injector.Message().Request.DocumentURL,
				Description: "you have sql injection",
				Remediation: "don't have sqli injection",
				Nav:         injector.Nav(),
				NavResultID: injector.NavResultID(),
				Evidence: &browserk.Evidence{
					String: body,
				},
				Reported: time.Now(),
			})
			return true, nil
		}
	}
	return true, nil
}

// TODO get 'median' response time for all requests by capturing stats across all response timing.
// this check will run the attack 3x alternating between slow sleep(15) and fast sleep(0) accounting for baseline
// timing etc. If 2 of the slow attacks work, and the fast attack returns 'fast' then we can be semi-certain it's a legit
// finding. Tweak this algo as necessary with various network conditions.
func (p *Plugin) doTimingAttack(injector browserk.Injector, attack *SQLIAttack) (bool, error) {
	expr := injector.InjectionExpr()
	expr.Inject(attack.Prefix+fmt.Sprintf(attack.Attack, p.sleepTimeSec)+attack.Suffix, browserk.InjectValue)

	originalBaseline := injector.Message().Response.ResponseTimeMs()
	longSleep := (time.Second * p.sleepTimeSec) + (time.Millisecond * time.Duration(originalBaseline))
	ctx, cancel := context.WithTimeout(injector.BCtx().Ctx, longSleep+(time.Second*5)) // give it 5 extra seconds to timeout
	defer cancel()
	m, err := injector.Send(ctx, false)
	if err != nil && err != browserk.ErrInjectionTimeout {
		return false, nil
	}

	// quite possible we ran over the timelimit still a potential sqli
	if err == browserk.ErrInjectionTimeout || m == nil {
		// do something
	} else {
		// TODO: verify this is right
		if m.Response.RecvTimestamp.Sub(m.Request.SentTimestamp) > time.Duration(p.sleepTimeSec)*time.Second {
			// do stuff
		}
	}

	return false, nil
}

// OnEvent handles passive events
func (p *Plugin) OnEvent(evt *browserk.PluginEvent) {
}
