# Design

The browserker crawler is based off of `Actions` that are executed as a part of a `Navigation.` Navigations are unique actions taken by the crawler to produce results. Navigations build upon each other to create paths.

These paths branch out to create a graph and will continue until there are no unique navigations left to take. This ends up looking like something below:

```
"path":"ActLoadURL [http://localhost:8080/]"
"path":"ActLoadURL [http://localhost:8080/] -> ActLeftClick [A href=2.php]"
"path":"ActLoadURL [http://localhost:8080/] -> ActLeftClick [A href=2.php] -> ActLeftClick [A href=../innerpages/2_1f84b.php] "
"path":"ActLoadURL [http://localhost:8080/] -> ActLeftClick [A href=2.php] -> ActLeftClick [A href=../innerpages/2_1f84b.php] -> ActLeftClick [A href=2_1f84b.php]"
"path":"ActLoadURL [http://localhost:8080/] -> ActLeftClick [A href=13.php]"
"path":"ActLoadURL [http://localhost:8080/] -> ActLeftClick [A href=13.php] -> ActLeftClick [INPUT value=click meonclick=doxhr()type=button]"
"path":"ActLoadURL [http://localhost:8080/] -> ActLeftClick [A href=18.php] -> ActSendKeys "
...
```

As the crawler executes navigations it marks them as either visited or failed in the crawl graph. Each 'path', once it reaches it's final stage captures various data about the state, extracts new potential actions to take and stores it in the graph. The browser itself is then killed and a new browser is taken from the pool of available browsers. Each path executes in it's own isolated browser process. This is to prevent leaking and in general makes managing browsers easier. (They crash a lot).

Once there are no more navigation nodes with the Unvisited state left, it exits the crawler loop.

### Uniqueness

Knowing whether a potential action is new is something any crawler must account for. Browserker's crawler uses a few methods. During each step or iteration of a navigation path, instrumentation is only enabled on the last navigation entry. This allows us to take a snapshot of the loaded DOM prior to executing our action, creating unique hashes of each element that exists, then execute our action.

We can then extract the potential actions and compare it with what we've already seen. If we haven't seen it (hashes don't match any action we snapshot'd) then we have a new action we can take.

This is not fool-proof, as a backup measure we also store navigations (again uniqueness hashes) as keys into the graph data store. If a key already exists, we simply ignore it.

### Handling inputs

Obviously a crawler must be able to click elements, and input values into elements which require input. A user configurable set of input field values can be supplied for various topics (address/name/credit card etc). When extracting forms and before generating new navigation entries, the form is analyzed for context specific information.

For example it looks if an input element has an associated label and combines the name/id/label information into a string and attempts to match it against a set of regexes. In other cases where input elements have a strict type defined (datetime/email etc) it's quite easy for us to supply a legitimate value. Once all the input fields have been analyzed and values set, the data is added to the next navigation entry and stored in the graphdb for later retrieval by the crawler.
