# Design

## Scanner Flow

- Starts up with configuration.
- Opens up crawl graph or adds the LOAD_URL from the command line / config to insert a new node into the crawl graph
- Opens up N browsers
- Create browserk.Context with all necessary services and handlers for various hooks (Pre-Auth hook etc)
- Crawler queries the crawl graph for UNVISITED paths
- Returned paths (in order they were found) are then iterated over, executing each entry in the same browser.
  - A path is a series of navigation steps LOAD URL -> CLICK BTN -> ENTER FORM etc
- Before crawler executes the last entry of the path, it enables instrumentation to capture new potential navigations

Active Plugins

- Much like the crawler, the Active Plugin System queries the crawl graph for UNAUDITED

## Uniqueness

URLs are meaningless for some things, but not others

## Browsers

A browser is an implementation of a gcd.Tab. The browser pool handles acquiring new browsers and returning old ones. The pool gets browsers from the leaser service which handles starting new ones and closing old ones.

## Storage

Pretty much every data type should be stored in our custom DB built on badger

- Crawl Graph
- Attack Graph / Plugin Work Graph
- Store req/resp in seperate badger nodes with requestID as key? keep graphs light?
- Findings

## Plugins

Should support running external commands to get easy wins for coverage (TLS testing etc)
Should be configurable for types:

- run once
- run once per path
- once per page
- run only on X mimetypes
- run only on X injection point types
- need ability to send direct requests for certain plugin types (might have to rewrite devtool methods/inject capabilities)
- should plugins have dependencies (on other plugins)?

Active Plugins are directly handed the navigation actions done by a crawler and re-execute them with a reference to the browser, this way they have full control over the entire navigation path and can know when to attack without requiring tracking which requests/responses to attack. Effectively making plugins first class citizens.

Passive Plugins register listeners for the types of data they want (storage events, network events, cookie events) and a passive manager filters out duplicates then dispatches new events to them to process.

Uniqueness is important here because we don't want to inundate plugins with the same events over and over again. Each event type, and a unique set of properties for that event is hashed together to create keys for: host, path, file, query, fragment, request and response. This allows us to only do a O(1) look up (per uniqueness check) and return a bitmask of uniqueness types. The plugin manager can then dispatch events to plugins which have their execution policy set to whatever uniqueness level they defined.

## Authentication and Session Management

TODO: maybe support loading selenium/injecting selenium into browserker so we get selenium capabilities for scripting?
Supporting things like JWT should be easy (we can inject whatever we want into browser processes)

## Attacking

TODO: What should plugins _get_? A list of injection points? A page? A browser? Register for specific events? Needs access to responses.
Needs ability to read response for _their_ injected request.

Current thinking is to replay a crawl navigation, then on the final step generate an HTML/JS page with a series of xhr/fetch functions which we can tag to be intercepted. The DevTools Fetch API would be enacted to trap the request and modify to whatever the plugin wanted.

There are some questions though:

1. Should we execute the _entire_ navigation path _every_ time? This is super expensive but will reduce state related issues where an attack may fail unless path A -> B -> C is followed in proper succession (think anti-CSRF tokens). Maybe replay the navigation 1 time per plugin, then have each plugin attack. Maybe have some sort of heuristic detect if we should replay the entire path again.
2. Detect anti-CSRF tokens and replay only the necessary steps.
3. If possible, generate JS functions that attack directly via XHR/Whatever.
4. Call out to a standard Go http.Client and attack from that, bypassing the browser. This will work on API/REST calls, but requires us to track session state and pass it along, also adds more complexity.

### Parsing Requests

Parse methods, request headers, x-www-form-urlencoded, url's/path queries and fragments, json and XML by hand using injast (parses data into an AST).

### Injection Points

Each k/v for parsed types should be an injection point. Injection Manager should handle re-encoding (similar to how astutil works in Go)

## Reporting

Report manager should be available to plugins, plugins can report their specific checks with evidence.
