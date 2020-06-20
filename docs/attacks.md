# Design

After crawling has completed, navigation paths are re-executed with the final navigation used for attacking. Like the crawler, the attack component scans the crawl graph looking for Visited paths, then pulls them out and sets the state to InProcess. Once all plugins have attacked, the navigation paths are set to the Audited state.

## What to attack

Browser's loading a web page make a number of additional requests besides just the document alone. Loading JS / CSS / Image resources, embedded iframe's and initiating web socket connections, these all become vectors for attack.

The DevTools protocol is kind enough to tag what type of resource and who requested it in the Network API domain. We can use this information to prioritize attacks to just the top level documents or specific resources. (eh, probably just attack everything to be honest.)

On top of the requests, there are the fields and parameters that make up the request. What Browserker intends to attack:

1. HTTP Method (ex: `GET`, `POST`)
2. HTTP Path fields (ex: `/path1/` `/path1/path2/`)
3. HTTP File fields (ex: `/path1/file.jsp` `/path1/file`)
4. Query Parameter Names: (ex `/file1?attackhere=1` `/file1?x=y&attackhere=2`)
5. Query Parameter Values: (ex `/file1?x=attackhere` `/file1?x=y&z=attackhere`)
6. Query Parameter Index Names: (ex `/file1?x[attackhere]=1` `/file1?x[0]=1&x[attackhere]=2`)
7. Fragment Paths (ex: `/file#/some/attackhere/` `/file#/attackhere`)
8. Fragment Parameters (ex: `/file#/some/path?x=attackhere` `/file#/some/path?attackhere=value`)
9. TODO: Query Parameter Value JSON Fields: (ex: `/file1?x={"some": "attackhere"} /file1?x={"attackhere": "value"}` )
10. TODO: Header Names/Values (ex: `AttackHere: value` `HeaderName: attackhere`)
11. TODO: Cookie Names/Values (ex: `Cookie: attackhere=1234` `Cookie: JSESSIONID=attackhere`)
12. TODO: Body Names/Values (same as query)
13. TODO: JSON Inside x-www-url-encoded values (ex: `name=value&json={"some": "attackhere"}`)
14. TODO: XML inside x-www-url-encoded values (ex: `name=value&xml=\<xml\>attackhere\</xml\>`)
15. TODO: JSON Body
16. TODO: XML Body
17. TODO: GraphQL Body
18. TODO: JSON in WebSockets
19. ... TODO Add more ...

## Correlating Attacks with Requests

One of the difficulties of crawl then attack based web app scanners is correlating which request we want to attack with which requests the browser initiates. Creating semi-unique hashes is an integral part of Browserker and HTTP Requests and Responses are no different. Provided a decent hashing methodology and the proper fields are hashed, the hope is that correlation will be as simple as hashing the target request (taken from the crawl phase) then intercepting all requests and comparing the hash.

Hashing HTTP Requests will probably be done against:

1. Paths (if not determined to be dynamic/random)
2. Query names only
3. Unique header names only
4. Body parameter names only

The hope is that this will be sufficient to determine a unique request for a particular navigation.

TODO: The above is a work in process, the logic may change if hashing doesn't work well.

TODO: WebSockets aren't really handled yet, but the idea will be to capture websocket requests during the crawl phase, diff them, and correlate web socket messages with each navigation action. When it comes time to attack, we will probably hook the websocket connection and sniff/intercept from the JS side to attack. Unfortunately, the DevTools protocol does not allow intercepting websocket messages at this time. Worst case scenario, we build our own websocket client and re-play the collected messages from the navigations.

## How Injections work

Using similar techniques to language processing, HTTP request parts are parsed and stored as an AST like structure. This allows us to break into the individual components of a URI or HTTP body and replace, cut, prepend, append and inject into where ever we want.

A plugin will receive a copy of the original request parts, parsed and ready for injection. Once the values have been modified, the AST like structure is re-serialized and injected into the appropriate field.

## What plugins get

The interface is still a WIP, but how plugins configure themselves is slowly solidifying. On registration a plugin will define to the system exactly what it needs.

```
// PluginExecutionType determines how often/when a plugin should be called/executed
type PluginExecutionType int8

const (
	ExecOnce PluginExecutionType = iota
	ExecOncePath
	ExecOnceFile
	ExecOncePerNavPath
	ExecPerRequest
	ExecAlways
)

type PluginOpts struct {
	IsolatedRequests bool                // Initiates it's own requests, isolated from a crawl state
	WriteResponses   bool                // writes/injects into http/websocket responses
	WriteRequests    bool                // writes/injects into http/websocket responses
	WriteJS          bool                // writes/injects JS into the browser
	ListenResponses  bool                // reads http/websocket responses
	ListenRequests   bool                // reads http/websocket requests
	ListenStorage    bool                // listens for local/sessionStorage write/read events
	ListenCookies    bool                // listens for cookie write events
	ListenConsole    bool                // listens for console.log events
	ListenURL        bool                // listens for URL change/updates
	ListenJS         bool                // listens to JS events
	ExecutionType    PluginExecutionType // How often/when this plugin executes
	Mimes            []string            // list of mime types this plugin will execute on if ExecutionType = ONLY_INJECTION
	Injections       []string            // list of injection points this plugin will execute on
}
```

They can define when they should be called (PluginExecutionType) and what capabilities they offer (WriteRequests/WriteJS etc).

## TODO

- How to handle plugins that make multiple requests (timing attacks)?
- Define exactly how plugins can interact with the browser (XSS will need full access basically)
- Plugins are currently depth first attacks (all iterating over the same navigation's). May want to have a breadth first where they attack individually, makes storing plugin navigation state harder.
