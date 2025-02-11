# Guide to the Outlived implementation

Outlived is a [Google App Engine](https://cloud.google.com/appengine) application.

It has a frontend written in [Typescript](https://www.typescriptlang.org/) and [React](https://react.dev/)
and a backend written in [Go](https://go.dev/).
The frontend uses UI components from the [Material Design library](https://mui.com/material-ui/).

## Frontend

The main entrypoint to the UI is the [App](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L68) function.
Among other things,
it initializes a variable called [loaded](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L72)
to `false`.
That variable being `false` triggers [this call](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L123) to `getData`.
A `false` value for `loaded` also hides the main UI [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L138),
instead showing a circular spinner, [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L191)
while data is being loaded from the server.

The [getData function](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L78) sends a `POST` request to the `/s/data` endpoint on the server, [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L80).
If that succeeds, information is parsed out of the response and used to set various pieces of state in the UI,
chiefly [the logged-in user](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L87) (if any)
and [the figures](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L85)
who died on this date in history
(and optionally also those who died at or just below the logged-in user’s current age).
Then [loaded is set to true](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L88)
which hides the circular spinner and reveals [the main UI](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/web/src/App.tsx#L140).

## Backend

Execution of the backend service begins [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/cmd/outlived/main.go#L15).
Execution quickly gets to [this point](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/cmd/outlived/serve.go#L12)
where a `Server` object is created with `NewServer` then listens for and serves requests [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/cmd/outlived/serve.go#L16).

The implementation of `NewServer` is [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/server.go#L19),
and its `Serve` method is [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/server.go#L64).
Among other things, `Serve` sets up a handler for `/s/data` requests in terms of a function called `s.handleData`, [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/server.go#L64).

That function is defined [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/data.go#L54).
It queries the database for “figures” who died on today’s date in past years, [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/data.go#L67).
If there is a current login session ([here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/data.go#L76)),
then the function also gets “user data”
(see `getUserData` and `getUserData2` in the same file).
The fetched data is bundled up and returned to the caller as [a dataResp object](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/site/data.go#L15).

The definition of `FiguresDiedOn`,
which produces the list of historical figures who died on a given month/day pair,
is [here](https://github.com/bobg/outlived/blob/e86a7e9a36f3e801bc32729f5ea7f4ec5b90e135/figure.go#L84).
