package sqli

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
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
	return "BR-A-0001"
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
	// expr := injector.InjectionExpr()
	for _, attack := range p.attacks {
		injector.BCtx().Log.Info().Str("attack", attack.Attack).Msg("attempting SQLi")
		if attack.IsTiming {
			found, err := p.doTimingAttack(injector, attack)
			if err != nil {
				injector.BCtx().Log.Warn().Err(err).Msg("attack failed")
			}
			if found {
				return true, nil
			}
		}
		/*
			TODO: handle generic sql i and match against common exceptions/error messages}
		*/
	}
	return false, nil
}

// TODO get 'median' response time for all requests by capturing stats across all response timing.
// this check will run the attack 4x alternating between slow sleep(15) and fast sleep(0) accounting for baseline
// timing etc. If 2 of the slow attacks work, and the fast attack returns 'fast' then we can be semi-certain it's a legit
// finding. Tweak this algo as necessary with various network conditions.
func (p *Plugin) doTimingAttack(injector browserk.Injector, attack *SQLIAttack) (bool, error) {
	expr := injector.InjectionExpr()

	originalBaseline := (time.Millisecond * time.Duration(injector.Message().Response.ResponseTimeMs()))

	longSleep := (time.Second * p.sleepTimeSec) + originalBaseline

	injector.BCtx().Log.Info().
		Str("attack", attack.Attack).
		Str("original_resp_ms", originalBaseline.String()).
		Str("sleep_time", longSleep.String()).
		Msg("attempting SQLi long sleep")

	// long sleep
	expr.Inject(attack.Prefix+fmt.Sprintf(attack.Attack, p.sleepTimeSec)+attack.Suffix, browserk.InjectValue)
	t, success := p.sendTiming(longSleep, injector)
	if success && t < longSleep {
		return false, nil
	}

	injector.BCtx().Log.Info().Str("attack", attack.Attack).Msg("attempting SQLi short sleep")
	// short sleep
	shortSleepFailCount := 0
	expr.Inject(attack.Prefix+fmt.Sprintf(attack.Attack, 0)+attack.Suffix, browserk.InjectValue)
	t, success = p.sendTiming(longSleep, injector)

	// if the connection timed out on the short sleep attack, it's probably just a busted page
	if !success || (success && t-originalBaseline > longSleep) {
		shortSleepFailCount++
	}
	injector.BCtx().Log.Info().Int("short_fail_cnt", shortSleepFailCount).Str("attack", attack.Attack).Msg("attempting SQLi short sleep x2")

	expr.Inject(attack.Prefix+fmt.Sprintf(attack.Attack, 0)+attack.Suffix, browserk.InjectValue)
	t, success = p.sendTiming(longSleep, injector)
	if !success || (success && t-originalBaseline > longSleep) {
		shortSleepFailCount++
	}

	injector.BCtx().Log.Info().Int("short_fail_cnt", shortSleepFailCount).Str("attack", attack.Attack).Msg("short sleep results")
	if shortSleepFailCount == 2 {
		injector.BCtx().Log.Info().Str("attack", attack.Attack).Msg("attack not successful, short sleep failed 2x")
		return false, nil
	}

	injector.BCtx().Log.Info().Int("short_fail_cnt", shortSleepFailCount).Str("attack", attack.Attack).Msg("attempting long sleep x 2")
	expr.Inject(attack.Prefix+fmt.Sprintf(attack.Attack, p.sleepTimeSec)+attack.Suffix, browserk.InjectValue)
	t, success = p.sendTiming(longSleep, injector)
	if success && t < longSleep {
		injector.BCtx().Log.Info().Str("attack", attack.Attack).Msg("attack not successful, time was < longSleep for 2nd attack")
		return false, nil
	}
	injector.BCtx().Log.Info().Str("attack", attack.Attack).Msg("attack successful, creating report")
	p.reportTimingSuccess(injector, attack)
	return true, nil
}

func (p *Plugin) reportTimingSuccess(injector browserk.Injector, attack *SQLIAttack) {
	injector.BCtx().PluginServicer.Store().AddReport(&browserk.Report{
		CheckID:     2,
		CWE:         89,
		URL:         injector.Message().Request.Request.Url,
		Description: fmt.Sprintf("you have sql injection in %s. %s", attack.DBTech.String(), attack.Description),
		Remediation: "don't have sqli injection",
		Nav:         injector.Nav(),
		NavResultID: injector.NavResultID(),
		Evidence:    browserk.NewEvidence(fmt.Sprintf("%s", attack.Prefix+attack.Attack+attack.Suffix)),
		Reported:    time.Now(),
	})
}

func (p *Plugin) sendTiming(timeout time.Duration, injector browserk.Injector) (time.Duration, bool) {
	ctx, cancel := context.WithTimeout(injector.BCtx().Ctx, timeout+(time.Second*5)) // give it 5 extra seconds to timeout
	defer cancel()
	start := time.Now()
	m, err := injector.SendWithCtx(ctx, false)
	log.Info().Err(ctx.Err()).Msg("ctx error?")
	if err != nil && err != browserk.ErrInjectionTimeout {
		return 0, false
	}

	// quite possible we ran over the timelimit still a potential sqli
	if err == browserk.ErrInjectionTimeout || m == nil {
		return start.Sub(time.Now()), false
		// do something
	}
	return m.Response.RecvTimestamp.Sub(m.Request.SentTimestamp), true
}

// OnEvent handles passive events
func (p *Plugin) OnEvent(evt *browserk.PluginEvent) {
}
