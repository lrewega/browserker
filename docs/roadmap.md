## Project Roadmap

Iterate quickly, get small things working as soon as possible to show progress.

### Goals

- [x] Get browser hooked up and store results as graph
- [x] Get minimal crawling working and storing results
- [x] Auto fill forms
- [x] Get passive plugins working
- [x] Get JS based plugins (via goja) working
- [x] Implement reporter system for reporting flaws
- [x] Implement uniqueness for reports
- [x] Implement injast for treating request params as an AST (JSON/XML/x-www-url-encoded etc)
- [x] Get active plugin attacks working (minimal, query only to start)
- [x] Implement uniqueness for plugin events (host, path, file, query)
- [x] Handle floating forms
- [x] Handle 'SPA' like pages better
- [x] Parse body for / json / xml
- [x] Handle frames better (Yeah, Nah.)
- [x] Export crawl graph in DOT format
- [ ] Diff out duplicate requests for attack phase (currently wasting lots of time attacking the same requests)
- [ ] Actually test / get JS Active Plugins working
- [ ] Get timing attack plugins working (SQL/OS/Code Injection etc)
- [ ] Get browser based attacks working (injecting into URL fragments)
- [ ] Get XSS plugin working (for stored XSS/inject js listeners etc)
- [ ] Get authentication working
- [ ] Get custom authentication scripts working
- [ ] Handle failures in navigations better
- [ ] Get page uniqueness working
- [ ] Get 404 detection working
- [ ] Hook common JS functions (window.print, window.open etc)
- [ ] Handle SPAs better (vuejs, react etc) hook routers etc (this may not be necessary after recent improvements)
- [ ] Get websocket events/attacks working
- [ ] Handle marking navigations if anti-CSRF tokens are identified for replaying
- [ ] Other?
