# Shipyard Server

This is a backend REST server to manage a shopping cart system. There is a
custom Oauth2 identity provider implementation included in the `idp` package.
It's modularized so it can be swapped out for an actual solution, like
[Auth0](https://auth0.com), fairly easily.


### Running Locally

```sh
make runsqlite3
```
